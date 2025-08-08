package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"task-scheduler/api"
	"task-scheduler/config"
	"task-scheduler/internal/scheduler"
	"task-scheduler/pkg/database"
	"task-scheduler/pkg/logger"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger.InitLogger(cfg.Log.Level)
	
	// 初始化数据库连接
	db, err := database.InitDB(cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB(db)
	
	// 初始化任务调度器
	scheduler := scheduler.NewScheduler(db)
	defer scheduler.Stop()
	
	// 从数据库加载并启动所有启用的任务
	if err := scheduler.LoadAndStartTasks(); err != nil {
		logger.Warnf("Failed to load and start tasks: %v", err)
	}
	
	// 设置路由
	router := api.SetupRouter(db, scheduler)
	
	// 创建HTTP服务器
	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}
	
	// 启动服务器（非阻塞）
	go func() {
		logger.Infof("Server is running on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()
	
	// 等待中断信号优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")
	
	// 优雅关闭服务器，等待10秒
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}
	
	logger.Info("Server exiting")
}
