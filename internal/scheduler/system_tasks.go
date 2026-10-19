package scheduler

import (
	"encoding/json"
	"fmt"
	"log"
	"octoops/internal/infra/postgres"
	systemModel "octoops/internal/model/system"
	taskModel "octoops/internal/model/task"
	aliyunService "octoops/internal/service/aliyun"
	seatunnelService "octoops/internal/service/seatunnel"
	"strings"
	"time"
)

// SystemTaskDef 系统内置任务定义，新增系统任务在这里登记即可
type SystemTaskDef struct {
	Type string
	Name string
}

var systemTaskDefs = []SystemTaskDef{
	{
		Type: "job_status_sync",
		Name: "ETL 实时作业状态同步",
	},
	{
		Type: "ecs_sg_sync",
		Name: "ECS 安全组同步",
	},
}

// SystemTaskDefByType 按类型查找系统任务定义
func SystemTaskDefByType(taskType string) (SystemTaskDef, bool) {
	for _, def := range systemTaskDefs {
		if def.Type == taskType {
			return def, true
		}
	}
	return SystemTaskDef{}, false
}

// loadSystemTasksFromSettings 从系统设置读取配置并注册系统任务
func loadSystemTasksFromSettings() {
	keys := make([]string, 0, len(systemTaskDefs))
	for _, def := range systemTaskDefs {
		keys = append(keys, systemTaskSettingKey(def.Type))
	}
	var settings []systemModel.Setting
	if err := postgres.DB.Where("\"key\" IN ?", keys).Find(&settings).Error; err != nil {
		log.Printf("[Scheduler][系统任务] 读取设置失败: %v", err)
		return
	}
	values := make(map[string]systemModel.TaskSetting, len(settings))
	for _, s := range settings {
		var ts systemModel.TaskSetting
		if err := json.Unmarshal([]byte(s.Value), &ts); err != nil {
			log.Printf("[Scheduler][系统任务] 设置项 %s 解析失败: %v", s.Key, err)
			continue
		}
		values[s.Key] = ts
	}

	enabled := 0
	for _, def := range systemTaskDefs {
		ts, ok := values[systemTaskSettingKey(def.Type)]
		if !ok || !ts.Enabled || strings.TrimSpace(ts.Cron) == "" {
			continue
		}
		if registerSystemTask(def, strings.TrimSpace(ts.Cron)) {
			enabled++
		}
	}
	log.Printf("[Scheduler] 系统任务加载完成，启用数量: %d/%d", enabled, len(systemTaskDefs))
}

// ApplySystemTask 按设置热更新某个系统任务（先摘旧调度，再按需注册）
func ApplySystemTask(taskType string, enabled bool, cronExpr string) {
	def, ok := SystemTaskDefByType(taskType)
	if !ok {
		log.Printf("[Scheduler][系统任务] 未知任务类型: %s", taskType)
		return
	}
	DisableSystemTask(taskType)
	if !enabled || strings.TrimSpace(cronExpr) == "" {
		log.Printf("[Scheduler][系统任务] 任务已停用: %s", taskType)
		return
	}
	registerSystemTask(def, strings.TrimSpace(cronExpr))
}

// DisableSystemTask 移除某个系统任务的调度
func DisableSystemTask(taskType string) {
	mapsMu.Lock()
	task := systemTasks[taskType]
	delete(systemTasks, taskType)
	mapsMu.Unlock()
	if task == nil || task.EntryID == 0 || cronScheduler == nil {
		return
	}
	cronScheduler.Remove(task.EntryID)
	log.Printf("[Scheduler][系统任务] 已移除调度: %s", taskType)
}

// runSystemTask 执行系统任务并捕获 panic，避免异常直接中断调度 goroutine 导致日志缺失
func runSystemTask(task *SystemTask) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Scheduler][Panic] 系统任务异常 type=%s, name=%s, err=%v", task.Type, task.Name, r)
			err = fmt.Errorf("任务执行异常: %v", r)
		}
	}()
	return task.Job()
}

func registerSystemTask(def SystemTaskDef, cronExpr string) bool {
	if cronScheduler == nil {
		log.Printf("[Scheduler][系统任务] 调度器未初始化，跳过注册: %s", def.Type)
		return false
	}
	task := &SystemTask{
		Type:   def.Type,
		Name:   def.Name,
		Spec:   cronExpr,
		Status: 1,
		Job:    GetJobFuncByType(def.Type),
	}
	mapsMu.Lock()
	systemTasks[def.Type] = task
	mapsMu.Unlock()
	if !addSystemTaskToCron(task) {
		mapsMu.Lock()
		delete(systemTasks, def.Type)
		mapsMu.Unlock()
		return false
	}
	return true
}

func addSystemTaskToCron(task *SystemTask) bool {
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
		result, runErr := runSystemTask(task)
		status := taskModel.TaskStatusSuccess
		if runErr != nil {
			status = taskModel.TaskStatusFailed
			result = runErr.Error()
		}
		mapsMu.Lock()
		task.LastResult = result
		if currentEntryID != 0 {
			entry := cronScheduler.Entry(currentEntryID)
			task.NextRun = computeNextRunFromEntry(entry, time.Now())
		}
		mapsMu.Unlock()

		if err := postgres.DB.Create(&taskModel.TaskLog{
			TaskName: task.Name,
			TaskType: taskModel.TaskTypeSystem,
			Result:   result,
			Status:   status,
		}).Error; err != nil {
			log.Printf("[Scheduler][系统任务] 写入任务日志失败 type=%s, err=%v", task.Type, err)
		}
	}
	entryID, err := cronScheduler.AddFunc(task.Spec, jobFunc)
	if err != nil {
		log.Printf("[Scheduler][系统任务] 添加失败 type=%s, name=%s, cron=%s, err=%v", task.Type, task.Name, task.Spec, err)
		return false
	}
	mapsMu.Lock()
	task.EntryID = entryID
	entry := cronScheduler.Entry(entryID)
	task.NextRun = computeNextRunFromEntry(entry, time.Now())
	mapsMu.Unlock()
	log.Printf(
		"[Scheduler][系统任务] 添加成功 type=%s, name=%s, cron=%s, nextRun=%s",
		task.Type, task.Name, task.Spec, task.NextRun.Format("2006-01-02 15:04:05"),
	)
	return true
}

// GetJobFuncByType 返回系统任务的执行函数，返回展示结果与错误
func GetJobFuncByType(taskType string) func() (string, error) {
	switch taskType {
	case "ecs_sg_sync":
		return aliyunService.SyncECSSecurityGroups
	case "job_status_sync":
		return seatunnelService.SyncAllJobStatus
	default:
		return func() (string, error) {
			return "系统任务执行完成", nil
		}
	}
}

func systemTaskSettingKey(taskType string) string { return "task." + taskType }
