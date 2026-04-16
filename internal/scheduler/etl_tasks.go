package scheduler

import (
	"fmt"
	"log"
	"octoops/internal/infra/postgres"
	seatunnelModel "octoops/internal/model/seatunnel"
	seatunnelService "octoops/internal/service/seatunnel"
	"time"
)

func loadActiveTasks() {
	var tasks []seatunnelModel.EtlTask
	if err := postgres.DB.Where("task_type = ? AND status = ? AND cron_expr != ?", "batch", 1, "").Find(&tasks).Error; err != nil {
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

func AddTask(task seatunnelModel.EtlTask) error {
	if task.CronExpr == "" {
		return fmt.Errorf("cron表达式不能为空")
	}

	taskCopy := task
	taskFunc := func() {
		if !schedulerRunning.Load() {
			return
		}
		executeTask(taskCopy)
	}

	entryID, err := cronScheduler.AddFunc(task.CronExpr, taskFunc)
	if err != nil {
		log.Printf("[Scheduler][ETL任务] 添加失败 id=%d, name=%s, cron=%s, err=%v", task.ID, task.Name, task.CronExpr, err)
		return fmt.Errorf("添加ETL定时任务失败: %v", err)
	}

	mapsMu.Lock()
	taskEntryMap[task.ID] = entryID
	etlTasksMap[task.ID] = &taskCopy
	mapsMu.Unlock()

	entry := cronScheduler.Entry(entryID)
	nextRunTime := computeNextRunFromEntry(entry, time.Now())
	nextRun := nextRunTime.Format("2006-01-02 15:04:05")
	log.Printf("[Scheduler][ETL任务] 添加成功 id=%d, name=%s, cron=%s, nextRun=%s", task.ID, task.Name, task.CronExpr, nextRun)
	return nil
}

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

func executeTask(task seatunnelModel.EtlTask) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Scheduler][Panic] ETL task crashed id=%d, name=%s, err=%v", task.ID, task.Name, r)
		}
	}()
	log.Printf("开始执行定时任务: ID=%d, 名称=%s", task.ID, task.Name)

	now := time.Now()
	postgres.DB.Model(&task).Update("last_run_time", now)

	respBody, err := seatunnelService.SubmitJobInternal(task.ID, false)
	if err != nil {
		log.Printf("执行定时任务失败: ID=%d, 名称=%s, 错误=%v", task.ID, task.Name, err)
		seatunnelService.WriteTaskLogWithStatus(task, []byte(err.Error()), "failed")
		return
	}

	seatunnelService.UpdateJobIdFromResponse(task.ID, respBody)
	seatunnelService.WriteTaskLog(task, respBody)
	log.Printf("定时任务执行成功: ID=%d, 名称=%s", task.ID, task.Name)
}

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
