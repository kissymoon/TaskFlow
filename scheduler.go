package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"gorm.io/gorm"

	"task-scheduler/internal/model"
	"task-scheduler/internal/repository"
	"task-scheduler/pkg/logger"
)

// Scheduler 任务调度器
type Scheduler struct {
	gc          gocron.Scheduler
	db          *gorm.DB
	taskRepo    *repository.TaskRepository
	execRepo    *repository.ExecutionRepository
	jobs        map[uint]gocron.JobID // 任务ID到JobID的映射
	mu          sync.RWMutex
}

// NewScheduler 创建新的调度器
func NewScheduler(db *gorm.DB) *Scheduler {
	// 创建gocron调度器
	gc, err := gocron.NewScheduler()
	if err != nil {
		logger.Fatalf("Failed to create scheduler: %v", err)
	}
	
	// 启动调度器
	gc.Start()
	
	return &Scheduler{
		gc:          gc,
		db:          db,
		taskRepo:    repository.NewTaskRepository(db),
		execRepo:    repository.NewExecutionRepository(db),
		jobs:        make(map[uint]gocron.JobID),
	}
}

// LoadAndStartTasks 从数据库加载并启动所有启用的任务
func (s *Scheduler) LoadAndStartTasks() error {
	tasks, err := s.taskRepo.ListAllEnabled()
	if err != nil {
		return fmt.Errorf("failed to list enabled tasks: %w", err)
	}
	
	for _, task := range tasks {
		if err := s.StartTask(&task); err != nil {
			logger.Warnf("Failed to start task %d: %v", task.ID, err)
		} else {
			logger.Infof("Started task: %s (ID: %d)", task.Name, task.ID)
		}
	}
	
	return nil
}

// StartTask 启动指定任务
func (s *Scheduler) StartTask(task *model.Task) error {
	// 先停止可能已存在的任务
	if err := s.StopTaskByID(task.ID); err != nil {
		logger.Warnf("Failed to stop existing task %d: %v", task.ID, err)
	}
	
	// 创建作业
	job, err := s.gc.NewJob(
		gocron.CronJob(task.CronExpr, false),
		gocron.NewTask(
			ExecuteTask,
			s.db,
			task.ID,
			s.execRepo,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}
	
	// 记录作业ID
	s.mu.Lock()
	s.jobs[task.ID] = job.ID()
	s.mu.Unlock()
	
	// 更新任务状态
	task.Status = model.TaskStatusRunning
	_, err = s.taskRepo.Update(task)
	
	return err
}

// StopTaskByID 停止指定ID的任务
func (s *Scheduler) StopTaskByID(taskID uint) error {
	s.mu.RLock()
	jobID, exists := s.jobs[taskID]
	s.mu.RUnlock()
	
	if exists {
		if err := s.gc.RemoveJob(jobID); err != nil {
			return fmt.Errorf("failed to remove job: %w", err)
		}
		
		s.mu.Lock()
		delete(s.jobs, taskID)
		s.mu.Unlock()
	}
	
	// 更新任务状态
	task, err := s.taskRepo.GetById(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	
	task.Status = model.TaskStatusStopped
	_, err = s.taskRepo.Update(task)
	
	return err
}

// ExecuteTaskNow 立即执行任务（不影响定时调度）
func (s *Scheduler) ExecuteTaskNow(taskID uint) (uint, error) {
	task, err := s.taskRepo.GetById(taskID)
	if err != nil {
		return 0, fmt.Errorf("failed to get task: %w", err)
	}
	
	// 创建执行记录
	execution := &model.TaskExecution{
		TaskID:    taskID,
		StartTime: time.Now(),
		Status:    model.ExecutionStatusRunning,
	}
	
	execution, err = s.execRepo.Create(execution)
	if err != nil {
		return 0, fmt.Errorf("failed to create execution record: %w", err)
	}
	
	// 异步执行任务
	go func() {
		ctx := context.Background()
		output, err := executeCommand(ctx, task.Command)
		
		endTime := time.Now()
		status := model.ExecutionStatusSuccess
		errorMsg := ""
		
		if err != nil {
			status = model.ExecutionStatusFailed
			errorMsg = err.Error()
		}
		
		// 更新执行记录
		execution.EndTime = &endTime
		execution.Status = status
		execution.Output = output
		execution.Error = errorMsg
		
		if _, err := s.execRepo.Update(execution); err != nil {
			logger.Errorf("Failed to update execution record %d: %v", execution.ID, err)
		}
	}()
	
	return execution.ID, nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	if s.gc != nil {
		s.gc.Shutdown()
	}
}
