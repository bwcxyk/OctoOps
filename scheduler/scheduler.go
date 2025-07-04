package scheduler

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"log"
	"octoops/db"
	"octoops/model"
	seatunnelModel "octoops/model/seatunnel"
	aliyunService "octoops/service/aliyun"
	seatunnelService "octoops/service/seatunnel"
	"time"
)

var cronScheduler *cron.Cron
var taskEntryMap map[uint]cron.EntryID // 存储任务ID到entry ID的映射

// 统一可管理任务结构
var customTasks = map[uint]*CustomTask{}

// 新增：ETL任务映射
var etlTasksMap = map[uint]*seatunnelModel.EtlTask{}

type CustomTask struct {
	ID         uint          `json:"id"`
	Name       string        `json:"name"`
	Type       string        `json:"type"`
	Spec       string        `json:"spec"`
	Enabled    bool          `json:"enabled"`
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
	taskEntryMap = make(map[uint]cron.EntryID)

	// 加载数据库自定义任务
	loadCustomTasksFromDB()

	// 启动时加载所有活跃的定时任务
	loadActiveTasks()

	log.Println("定时任务调度器已启动")
}

// 加载所有活跃的定时任务
func loadActiveTasks() {
	var tasks []seatunnelModel.EtlTask
	db.DB.Where("task_type = ? AND status = ? AND cron_expr != ?", "batch", 1, "").Find(&tasks)

	for _, task := range tasks {
		AddTask(task)
	}

	log.Printf("加载了 %d 个活跃的ETL定时任务", len(tasks))
}

// 添加定时任务
func AddTask(task seatunnelModel.EtlTask) error {
	if task.CronExpr == "" {
		return fmt.Errorf("cron表达式不能为空")
	}

	// 创建任务函数
	taskFunc := func() {
		executeTask(task)
	}

	// 添加定时任务
	entryID, err := cronScheduler.AddFunc(task.CronExpr, taskFunc)
	if err != nil {
		return fmt.Errorf("添加ETL定时任务失败: %v", err)
	}

	// 保存任务ID到entry ID的映射
	taskEntryMap[task.ID] = entryID
	// 注册到 etlTasksMap
	etlTasksMap[task.ID] = &task

	log.Printf("成功添加ETL定时任务: ID=%d, 任务名称=%s, Cron表达式=%s", entryID, task.Name, task.CronExpr)
	return nil
}

// 移除定时任务
func RemoveTask(taskID uint) {
	if entryID, exists := taskEntryMap[taskID]; exists {
		cronScheduler.Remove(entryID)
		delete(taskEntryMap, taskID)
		delete(etlTasksMap, taskID)
		log.Printf("成功移除定时任务: 任务ID=%d, EntryID=%d", taskID, entryID)
	}
}

// 重新加载所有任务
func ReloadTasks() {
	reloadTasks()
}

// 重新加载所有任务
func reloadTasks() {
	// 停止当前调度器
	cronScheduler.Stop()

	// 创建新的调度器
	cronScheduler = cron.New(cron.WithSeconds())
	cronScheduler.Start()

	// 清空映射
	taskEntryMap = make(map[uint]cron.EntryID)

	// 重新加载自定义任务
	loadCustomTasksFromDB()
	// 重新加载etl任务
	loadActiveTasks()
}

// 执行任务
func executeTask(task seatunnelModel.EtlTask) {
	log.Printf("开始执行定时任务: ID=%d, 名称=%s", task.ID, task.Name)

	// 更新最后运行时间
	now := time.Now()
	db.DB.Model(&task).Update("last_run_time", now)

	// 调用提交作业服务
	respBody, err := seatunnelService.SubmitJobInternal(task.ID, false) // 默认不使用savepoint
	if err != nil {
		log.Printf("执行定时任务失败: ID=%d, 名称=%s, 错误=%v", task.ID, task.Name, err)
	} else {
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
		// 先查 customTasks
		for _, t := range customTasks {
			if t.EntryID == entry.ID {
				taskName = t.Name
				taskType = "custom"
				break
			}
		}
		// 再查 ETL 任务
		if taskName == "" {
			for _, t := range etlTasksMap {
				if taskEntryMap[t.ID] == entry.ID {
					taskName = t.Name
					taskType = "etl"
					break
				}
			}
		}
		activeTasks = append(activeTasks, map[string]interface{}{
			"entry_id":  entry.ID,
			"task_name": taskName,
			"task_type": taskType,
			"next_run":  entry.Next,
		})
	}

	return map[string]interface{}{
		"active_tasks_count": len(activeTasks),
		"active_tasks":       activeTasks,
	}
}

// 获取任务的下次执行时间
func GetTaskNextRunTime(taskID uint) *time.Time {
	if entryID, exists := taskEntryMap[taskID]; exists {
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
	for taskID := range taskEntryMap {
		nextRun := GetTaskNextRunTime(taskID)
		if nextRun != nil {
			result[taskID] = nextRun
		}
	}
	return result
}

// 注册自定义任务
func RegisterCustomTask(id uint, name, typ, spec string, enabled bool, job func() string) {
	task := &CustomTask{
		ID:      id,
		Name:    name,
		Type:    typ,
		Spec:    spec,
		Enabled: enabled,
		Job:     job,
	}
	customTasks[id] = task
	log.Printf("[Scheduler] RegisterCustomTask: id=%d, name=%s, enabled=%v, spec=%s", id, name, enabled, spec)
	if enabled {
		addCustomTaskToCron(task)
	}
}

func addCustomTaskToCron(task *CustomTask) {
	entryID, err := cronScheduler.AddFunc(task.Spec, func() {
		task.LastRun = time.Now()
		result := task.Job()
		task.LastResult = result
		entry := cronScheduler.Entry(task.EntryID)
		task.NextRun = entry.Next

		// 新增：同步写回数据库
		db.DB.Model(&model.CustomTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
			"last_run_time": task.LastRun,
			"last_result":   task.LastResult,
		})

		// 新增：写入日志表
		db.DB.Create(&model.TaskLog{
			TaskID:   task.ID,
			JobID:    "",
			JobName:  task.Name,
			Result:   result,
			TaskType: "custom",
		})
	})
	if err == nil {
		task.EntryID = entryID
		task.NextRun = cronScheduler.Entry(entryID).Next
		log.Printf("[Scheduler] addCustomTaskToCron: 成功添加自定义任务到调度器 id=%d, entryID=%d, nextRun=%v", task.ID, entryID, task.NextRun)
	} else {
		log.Printf("[Scheduler] addCustomTaskToCron: 添加自定义任务失败 id=%d, err=%v", task.ID, err)
	}
}

func EnableCustomTask(id uint) {
	task, ok := customTasks[id]
	if ok && !task.Enabled {
		task.Enabled = true
		addCustomTaskToCron(task)
		// 同步更新数据库
		db.DB.Model(&model.CustomTask{}).Where("id = ?", id).Update("status", 1)
	}
}

func DisableCustomTask(id uint) {
	task, ok := customTasks[id]
	if ok && task.Enabled {
		cronScheduler.Remove(task.EntryID)
		task.Enabled = false
		// 同步更新数据库
		db.DB.Model(&model.CustomTask{}).Where("id = ?", id).Update("status", 0)
	}
}

func RunCustomTaskNow(id uint) string {
	task, ok := customTasks[id]
	if ok {
		task.LastRun = time.Now()
		result := task.Job()
		task.LastResult = result
		// 新增：同步写回数据库
		db.DB.Model(&model.CustomTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
			"last_run_time": task.LastRun,
			"last_result":   task.LastResult,
		})
		// 新增：写入日志表
		db.DB.Create(&model.TaskLog{
			TaskID:   task.ID,
			JobID:    "",
			JobName:  task.Name,
			Result:   result,
			TaskType: "custom",
		})
		return result
	}
	return "任务不存在"
}

func ListCustomTasks() []*CustomTask {
	tasks := []*CustomTask{}
	for _, t := range customTasks {
		tasks = append(tasks, t)
	}
	return tasks
}

func loadCustomTasksFromDB() {
	var tasks []model.CustomTask
	db.DB.Find(&tasks)
	log.Printf("[Scheduler] 数据库加载自定义任务数量: %d", len(tasks))
	for _, t := range tasks {
		log.Printf("[Scheduler] 尝试注册自定义任务: id=%d, name=%s, status=%d, cron=%s", t.ID, t.Name, t.Status, t.CronExpr)
		RegisterCustomTask(
			t.ID,
			t.Name,
			t.CustomType,
			t.CronExpr,
			t.Status == 1,
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
	}
}

// 停止调度器
func StopScheduler() {
	if cronScheduler != nil {
		cronScheduler.Stop()
	}
}
