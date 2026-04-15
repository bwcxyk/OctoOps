package scheduler

import (
	"log"
	seatunnelModel "octoops/internal/model/seatunnel"
	"time"

	"github.com/robfig/cron/v3"
)

// InitScheduler 初始化调度器
func InitScheduler() {
	cronScheduler = cron.New(cron.WithSeconds())
	cronScheduler.Start()
	schedulerRunning.Store(true)
	mapsMu.Lock()
	taskEntryMap = make(map[uint]cron.EntryID)
	customTasks = map[uint]*CustomTask{}
	etlTasksMap = map[uint]*seatunnelModel.EtlTask{}
	mapsMu.Unlock()

	loadCustomTasksFromDB()
	loadActiveTasks()

	log.Println("定时任务调度器已启动")
}

// ReloadTasks 重新加载所有任务
func ReloadTasks() {
	reloadTasks()
}

func reloadTasks() {
	if cronScheduler != nil {
		ctx := cronScheduler.Stop()
		select {
		case <-ctx.Done():
			log.Println("[Scheduler] previous scheduler stopped")
		case <-time.After(5 * time.Second):
			log.Println("[Scheduler] stop timeout, switching to new scheduler")
		}
	}

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

	loadCustomTasksFromDB()
	loadActiveTasks()
	cronScheduler.Start()
	schedulerRunning.Store(true)
}

// GetSchedulerStatus 获取调度器状态
func GetSchedulerStatus() map[string]interface{} {
	entries := cronScheduler.Entries()

	var activeTasks []map[string]interface{}
	for _, entry := range entries {
		taskName := ""
		taskType := ""
		mapsMu.RLock()
		for _, t := range customTasks {
			if t.EntryID == entry.ID {
				taskName = t.Name
				taskType = "custom"
				break
			}
		}
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

// StartScheduler 启动调度器
func StartScheduler() {
	if cronScheduler != nil {
		cronScheduler.Start()
		schedulerRunning.Store(true)
		log.Println("[Scheduler] started")
	} else {
		log.Println("[Scheduler] start requested but scheduler is nil")
	}
}

// StopScheduler 停止调度器
func StopScheduler() {
	if cronScheduler != nil {
		cronScheduler.Stop()
		schedulerRunning.Store(false)
		log.Println("[Scheduler] stopped")
	} else {
		log.Println("[Scheduler] stop requested but scheduler is nil")
	}
}
