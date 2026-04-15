package scheduler

import (
	"log"
	"octoops/internal/db"
	taskModel "octoops/internal/model/task"
	aliyunService "octoops/internal/service/aliyun"
	seatunnelService "octoops/internal/service/seatunnel"
	"time"

	"github.com/robfig/cron/v3"
)

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

		db.DB.Model(&taskModel.CustomTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
			"last_run_time": task.LastRun,
			"last_result":   task.LastResult,
		})

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
