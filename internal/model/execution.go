package model

import (
	"time"

	"gorm.io/gorm"
)

// ExecutionStatus 执行状态
type ExecutionStatus string

const (
	ExecutionStatusRunning ExecutionStatus = "running"
	ExecutionStatusSuccess ExecutionStatus = "success"
	ExecutionStatusFailed  ExecutionStatus = "failed"
)

// TaskExecution 记录任务的一次执行
type TaskExecution struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	TaskID    uint           `gorm:"not null;index" json:"task_id"`
	StartTime time.Time      `json:"start_time"`
	EndTime   *time.Time     `json:"end_time,omitempty"`
	Status    ExecutionStatus `gorm:"default:running" json:"status"`
	Output    string         `gorm:"type:text" json:"output,omitempty"`
	Error     string         `gorm:"type:text" json:"error,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	// 关联
	Task *Task `gorm:"foreignKey:TaskID" json:"task,omitempty"`
}
