package scheduler

import (
	"fmt"
	"log"
	"octoops/internal/db"
	seatunnelModel "octoops/internal/model/seatunnel"
	taskModel "octoops/internal/model/task"
	aliyunService "octoops/internal/service/aliyun"
	seatunnelService "octoops/internal/service/seatunnel"
	"sync"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"
)

var cronScheduler *cron.Cron
var taskEntryMap map[uint]cron.EntryID // taskID -> entryID
var schedulerRunning atomic.Bool
var customTasks = map[uint]*CustomTask{}             // unified custom task store
var etlTasksMap = map[uint]*seatunnelModel.EtlTask{} // taskID -> ETL task

var mapsMu sync.RWMutex

type CustomTask struct {
	ID         uint          `json:"id"`
	Name       string        `json:"name"`
	Type       string        `json:"type"`
	Spec       string        `json:"spec"`
	Status     int           `json:"status"`
	LastRun    time.Time     `json:"last_run"`
	NextRun    time.Time     `json:"next_run"`
	LastResult string        `json:"last_result"`
	EntryID    cron.EntryID  `json:"entry_id"`
	Job        func() string `json:"-"`
}

// 修改 QueryOctoopsJobStatus 返回结构，支持 finishedTime
type JobStatusResult struct {
	JobStatus    string `json:"jobStatus"`
	FinishedTime string `json:"finishedTime"`
}

// 初始化调度器
func InitScheduler() {
	cronScheduler = cron.New(cron.WithSeconds())
	cronScheduler.Start()
	schedulerRunning.Store(true)
	mapsMu.Lock()
	taskEntryMap = make(map[uint]cron.EntryID)
	customTasks = map[uint]*CustomTask{}
	etlTasksMap = map[uint]*seatunnelModel.EtlTask{}
	mapsMu.Unlock()

	// 加载数据库自定义任务
	loadCustomTasksFromDB()

	// 启动时加载所有活跃的定时任务
	loadActiveTasks()

	log.Println("定时任务调度器已启动")
}

// 加载所有活跃的定时任务
func loadActiveTasks() {
	var tasks []seatunnelModel.EtlTask
	if err := db.DB.Where("task_type = ? AND status = ? AND cron_expr != ?", "batch", 1, "").Find(&tasks).Error; err != nil {
		log.Printf("[Scheduler] failed to load tasks: %v", err)
		return
	}

	for _, task := range tasks {
		if err := AddTask(task); err != nil {
			log.Printf("[Scheduler] failed to add task name=%s id=%d: %v", task.Name, task.ID, err)
			continue
		}
	}

	log.Printf("加载了 %d 个活跃的ETL定时任务", len(tasks))
}

// 添加定时任务
func AddTask(task seatunnelModel.EtlTask) error {
	if task.CronExpr == "" {
		return fmt.Errorf("cron表达式不能为空")
	}

	// 创建任务函数
	taskCopy := task
	taskFunc := func() {
		if !schedulerRunning.Load() {
			return
		}
		executeTask(taskCopy)
	}

	// 添加定时任务
	entryID, err := cronScheduler.AddFunc(task.CronExpr, taskFunc)
	if err != nil {
		log.Printf("[Scheduler][ETL任务] 添加失败 id=%d, name=%s, cron=%s, err=%v", task.ID, task.Name, task.CronExpr, err)
		return fmt.Errorf("添加ETL定时任务失败: %v", err)
	}

	// Save task ID to entry ID mapping
	mapsMu.Lock()
	taskEntryMap[task.ID] = entryID
	// Register to etlTasksMap
	etlTasksMap[task.ID] = &taskCopy
	mapsMu.Unlock()
	// nextRun 只保留日期和时间
	entry := cronScheduler.Entry(entryID)
	nextRunTime := computeNextRunFromEntry(entry, time.Now())
	nextRun := nextRunTime.Format("2006-01-02 15:04:05")
	log.Printf("[Scheduler][ETL任务] 添加成功 id=%d, name=%s, cron=%s, nextRun=%s", task.ID, task.Name, task.CronExpr, nextRun)
	return nil
}

// 移除定时任务
func RemoveTask(taskID uint) {
	mapsMu.Lock()
	entryID, exists := taskEntryMap[taskID]
	name := ""
	if exists {
		delete(taskEntryMap, taskID)
		if t, ok := etlTasksMap[taskID]; ok {
			name = t.Name
		}
		delete(etlTasksMap, taskID)
	}
	mapsMu.Unlock()
	if exists {
		cronScheduler.Remove(entryID)
		log.Printf("[Scheduler][ETL] removed id=%d, name=%s", taskID, name)
	}
}

// 重新加载所有任务
func ReloadTasks() {
	reloadTasks()
}

// 重新加载所有任务
func reloadTasks() {
	// Stop current scheduler and wait with timeout
	if cronScheduler != nil {
		ctx := cronScheduler.Stop()
		select {
		case <-ctx.Done():
			log.Println("[Scheduler] previous scheduler stopped")
		case <-time.After(5 * time.Second):
			log.Println("[Scheduler] stop timeout, switching to new scheduler")
		}
	}

	// Clear mappings and remove existing entries
	mapsMu.Lock()
	entryIDs := make([]cron.EntryID, 0, len(taskEntryMap))
	for _, entryID := range taskEntryMap {
		entryIDs = append(entryIDs, entryID)
	}
	for _, t := range customTasks {
		if t.EntryID != 0 {
			entryIDs = append(entryIDs, t.EntryID)
		}
	}
	taskEntryMap = make(map[uint]cron.EntryID)
	customTasks = map[uint]*CustomTask{}
	etlTasksMap = map[uint]*seatunnelModel.EtlTask{}
	mapsMu.Unlock()
	for _, entryID := range entryIDs {
		cronScheduler.Remove(entryID)
	}

	// Reload custom tasks
	loadCustomTasksFromDB()
	// Reload ETL tasks
	loadActiveTasks()
	cronScheduler.Start()
	schedulerRunning.Store(true)
}

// 执行任务
func executeTask(task seatunnelModel.EtlTask) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Scheduler][Panic] ETL task crashed id=%d, name=%s, err=%v", task.ID, task.Name, r)
		}
	}()
	log.Printf("开始执行定时任务: ID=%d, 名称=%s", task.ID, task.Name)

	// 更新最后运行时间
	now := time.Now()
	db.DB.Model(&task).Update("last_run_time", now)

	// 调用提交作业服务
	respBody, err := seatunnelService.SubmitJobInternal(task.ID, false) // 默认不使用savepoint
	if err != nil {
		log.Printf("执行定时任务失败: ID=%d, 名称=%s, 错误=%v", task.ID, task.Name, err)
		// 记录失败日志
		seatunnelService.WriteTaskLogWithStatus(task, []byte(err.Error()), "failed")
	} else {
		// 从响应中提取 jobId 并更新到数据库
		seatunnelService.UpdateJobIdFromResponse(task.ID, respBody)

		// 写入作业日志
		seatunnelService.WriteTaskLog(task, respBody)
		log.Printf("定时任务执行成功: ID=%d, 名称=%s", task.ID, task.Name)
	}
}

// 获取调度器状态
func GetSchedulerStatus() map[string]interface{} {
	entries := cronScheduler.Entries()

	var activeTasks []map[string]interface{}
	for _, entry := range entries {
		taskName := ""
		taskType := ""
		mapsMu.RLock()
		// Check customTasks first
		for _, t := range customTasks {
			if t.EntryID == entry.ID {
				taskName = t.Name
				taskType = "custom"
				break
			}
		}
		// Then check ETL tasks
		if taskName == "" {
			for _, t := range etlTasksMap {
				if taskEntryMap[t.ID] == entry.ID {
					taskName = t.Name
					taskType = "etl"
					break
				}
			}
		}
		mapsMu.RUnlock()
		activeTasks = append(activeTasks, map[string]interface{}{
			"entry_id":  entry.ID,
			"task_name": taskName,
			"task_type": taskType,
			"next_run":  entry.Next,
		})
	}

	return map[string]interface{}{
		"scheduler_running":  schedulerRunning.Load(),
		"active_tasks_count": len(activeTasks),
		"active_tasks":       activeTasks,
	}
}

// 获取任务的下次执行时间
func GetTaskNextRunTime(taskID uint) *time.Time {
	mapsMu.RLock()
	entryID, exists := taskEntryMap[taskID]
	mapsMu.RUnlock()
	if exists {
		entries := cronScheduler.Entries()
		for _, entry := range entries {
			if entry.ID == entryID {
				return &entry.Next
			}
		}
	}
	return nil
}

// 获取所有任务的下次执行时间
func GetAllTasksNextRunTime() map[uint]*time.Time {
	result := make(map[uint]*time.Time)
	mapsMu.RLock()
	keys := make([]uint, 0, len(taskEntryMap))
	for taskID := range taskEntryMap {
		keys = append(keys, taskID)
	}
	mapsMu.RUnlock()
	for _, taskID := range keys {
		nextRun := GetTaskNextRunTime(taskID)
		if nextRun != nil {
			result[taskID] = nextRun
		}
	}
	return result
}

// 注册自定义任务
func RegisterCustomTask(id uint, name, typ, spec string, status int, job func() string) {
	task := &CustomTask{
		ID:     id,
		Name:   name,
		Type:   typ,
		Spec:   spec,
		Status: status,
		Job:    job,
	}
	mapsMu.Lock()
	customTasks[id] = task
	mapsMu.Unlock()
	if status == 1 {
		addCustomTaskToCron(task)
	}
}

func addCustomTaskToCron(task *CustomTask) {
	var err error
	jobFunc := func() {
		if !schedulerRunning.Load() {
			return
		}
		mapsMu.RLock()
		currentEntryID := task.EntryID
		mapsMu.RUnlock()

		mapsMu.Lock()
		task.LastRun = time.Now()
		mapsMu.Unlock()
		result := task.Job()
		mapsMu.Lock()
		task.LastResult = result
		if currentEntryID != 0 {
			entry := cronScheduler.Entry(currentEntryID)
			task.NextRun = computeNextRunFromEntry(entry, time.Now())
		}
		mapsMu.Unlock()

		// Write back to database
		db.DB.Model(&taskModel.CustomTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
			"last_run_time": task.LastRun,
			"last_result":   task.LastResult,
		})

		// Write task log
		db.DB.Create(&taskModel.TaskLog{
			TaskName: task.Name,
			Result:   result,
			Status:   "success",
		})
	}
	entryID, err := cronScheduler.AddFunc(task.Spec, jobFunc)
	if err == nil {
		mapsMu.Lock()
		task.EntryID = entryID
		entry := cronScheduler.Entry(entryID)
		task.NextRun = computeNextRunFromEntry(entry, time.Now())
		mapsMu.Unlock()
		nextRun := task.NextRun.Format("2006-01-02 15:04:05")
		log.Printf("[Scheduler][Custom] added id=%d, name=%s, cron=%s, nextRun=%s", task.ID, task.Name, task.Spec, nextRun)
	} else {
		log.Printf("[Scheduler][Custom] add failed id=%d, name=%s, cron=%s, err=%v", task.ID, task.Name, task.Spec, err)
	}
}

func DisableCustomTask(id uint) {
	mapsMu.Lock()
	task, ok := customTasks[id]
	entryID := cron.EntryID(0)
	if ok && task.Status == 1 {
		entryID = task.EntryID
		task.Status = 0
	}
	mapsMu.Unlock()
	if ok && entryID != 0 {
		cronScheduler.Remove(entryID)
		// Update database
		db.DB.Model(&taskModel.CustomTask{}).Where("id = ?", id).Update("status", 0)
	}
}

func loadCustomTasksFromDB() {
	var tasks []taskModel.CustomTask
	db.DB.Find(&tasks)
	log.Printf("[Scheduler] 数据库加载自定义任务数量: %d", len(tasks))
	for _, t := range tasks {
		RegisterCustomTask(
			t.ID,
			t.Name,
			t.CustomType,
			t.CronExpr,
			t.Status,
			GetJobFuncByType(t.CustomType),
		)
	}
}

func GetJobFuncByType(customType string) func() string {
	switch customType {
	case "ecs_sg_sync":
		return func() string {
			return aliyunService.SyncECSSecurityGroups()
		}
	case "job_status_sync":
		return func() string {
			seatunnelService.SyncAllJobStatus()
			return "作业状态同步完成"
		}
	default:
		return func() string {
			return "自定义任务执行完成"
		}
	}
}

// 启动调度器
func StartScheduler() {
	if cronScheduler != nil {
		cronScheduler.Start()
		schedulerRunning.Store(true)
		log.Println("[Scheduler] started")
	} else {
		log.Println("[Scheduler] start requested but scheduler is nil")
	}
}

// 停止调度器
func StopScheduler() {
	if cronScheduler != nil {
		cronScheduler.Stop()
		schedulerRunning.Store(false)
		log.Println("[Scheduler] stopped")
	} else {
		log.Println("[Scheduler] stop requested but scheduler is nil")
	}
}

func computeNextRunFromEntry(entry cron.Entry, now time.Time) time.Time {
	if entry.Schedule != nil {
		return entry.Schedule.Next(now)
	}
	if !entry.Next.IsZero() {
		return entry.Next
	}
	return time.Time{}
}
