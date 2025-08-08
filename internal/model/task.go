package model

import (
	"time"

	"gorm.io/gorm"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusRunning TaskStatus = "running"
	TaskStatusStopped TaskStatus = "stopped"
)

// Task 表示一个定时任务
type Task struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:100;not null;unique" json:"name"`
	Description string         `gorm:"size:500" json:"description"`
	CronExpr    string         `gorm:"not null" json:"cron_expr"` // cron表达式
	Command     string         `gorm:"not null" json:"command"`   // 要执行的命令
	IsEnabled   bool           `gorm:"default:true" json:"is_enabled"`
	Status      TaskStatus     `gorm:"default:stopped" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate 创建前的钩子
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.Status == "" {
		t.Status = TaskStatusStopped
	}
	return nil
}
