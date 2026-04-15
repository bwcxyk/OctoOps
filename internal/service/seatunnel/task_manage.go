package seatunnel

import (
	"octoops/internal/db"
	seatunnelModel "octoops/internal/model/seatunnel"
)

type TaskListFilter struct {
	TaskType  string
	Name      string
	Status    *int
	JobID     string
	JobStatus string
	Page      int
	PageSize  int
}

func ListTasks(filter TaskListFilter) ([]seatunnelModel.EtlTask, int64, error) {
	var tasks []seatunnelModel.EtlTask
	query := db.DB
	if filter.TaskType != "" {
		query = query.Where("task_type = ?", filter.TaskType)
	}
	if filter.Name != "" {
		query = query.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.JobID != "" {
		query = query.Where("job_id = ?", filter.JobID)
	}
	if filter.JobStatus != "" {
		query = query.Where("job_status = ?", filter.JobStatus)
	}
	query = query.Order("created_at desc")

	var total int64
	if err := query.Model(&seatunnelModel.EtlTask{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

func CreateTask(task *seatunnelModel.EtlTask) error {
	return db.DB.Create(task).Error
}

func GetTaskByID(id interface{}) (seatunnelModel.EtlTask, error) {
	var task seatunnelModel.EtlTask
	err := db.DB.First(&task, id).Error
	return task, err
}

func UpdateTask(task *seatunnelModel.EtlTask, updates map[string]interface{}) error {
	return db.DB.Model(task).Updates(updates).Error
}

func DeleteTask(task *seatunnelModel.EtlTask) error {
	return db.DB.Delete(task).Error
}
