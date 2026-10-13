package scheduler

import (
	"log"
	"octoops/internal/infra/postgres"
	systemModel "octoops/internal/model/system"
	taskModel "octoops/internal/model/task"
	syssetting "octoops/internal/pkg/setting"
	"time"

	"github.com/robfig/cron/v3"
)

// 内置维护任务：每天 03:10 清理过期任务日志
const taskLogCleanupCron = "0 10 3 * * *"
const taskLogCleanupName = "任务日志清理"

// maintenanceEntries 内置维护任务的 cron entry，key 为 entryID
var maintenanceEntries = map[cron.EntryID]string{}

// registerMaintenanceTasks 注册内置维护任务，并立即执行一次
func registerMaintenanceTasks() {
	if cronScheduler == nil {
		log.Printf("[Scheduler][维护任务] 调度器未初始化，跳过注册")
		return
	}
	entryID, err := cronScheduler.AddFunc(taskLogCleanupCron, CleanupTaskLogs)
	if err != nil {
		log.Printf("[Scheduler][维护任务] 注册失败: %v", err)
		return
	}
	mapsMu.Lock()
	maintenanceEntries[entryID] = taskLogCleanupName
	mapsMu.Unlock()
	log.Printf("[Scheduler][维护任务] 已注册 %s, cron=%s", taskLogCleanupName, taskLogCleanupCron)

	// 启动时先清理一次，避免历史日志堆积
	CleanupTaskLogs()
}

// taskLogRetentionDays 从系统设置读取保留天数
// 与 service/system 保持同样的解析规则：调度器不能被 service/system 反向依赖，这里直接读设置表
func taskLogRetentionDays() int {
	var rs systemModel.RetentionSetting
	if !syssetting.Get(systemModel.KeyTaskLogRetentionDays, &rs) {
		return systemModel.DefaultTaskLogRetentionDays
	}
	if rs.Days != systemModel.TaskLogRetentionDisabled && rs.Days < 1 {
		return systemModel.DefaultTaskLogRetentionDays
	}
	return rs.Days
}

// CleanupTaskLogs 按保留天数清理任务日志，-1 表示不清理
func CleanupTaskLogs() {
	days := taskLogRetentionDays()
	if days == systemModel.TaskLogRetentionDisabled {
		log.Printf("[任务日志清理] 未开启（保留天数为 -1），跳过")
		return
	}
	deadline := time.Now().AddDate(0, 0, -days)
	res := postgres.DB.Where("created_at < ?", deadline).Delete(&taskModel.TaskLog{})
	if res.Error != nil {
		log.Printf("[任务日志清理] 清理失败: %v", res.Error)
		return
	}
	if res.RowsAffected > 0 {
		log.Printf(
			"[任务日志清理] 已删除 %d 条日志（保留 %d 天，早于 %s）",
			res.RowsAffected, days, deadline.Format("2006-01-02 15:04:05"),
		)
	}
}
