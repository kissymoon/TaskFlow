package main

import (
	"database/sql"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/gorilla/mux"
)

var db *sql.DB
var logger *logrus.Logger

func init() {
	// 加载环境变量
	err := godotenv.Load()
	if err != nil {
		logrus.Fatalf("Error loading .env file: %v", err)
	}

	// 初始化日志
	initLogger()

	// 初始化数据库连接
	initDB()
}

func initLogger() {
	logger = logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})
	
	logLevel, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logger.Warnf("Invalid log level, using default: info. Error: %v", err)
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)
}

func initDB() {
	// 配置MySQL连接
	cfg := mysql.Config{
		User:   os.Getenv("DB_USER"),
		Passwd: os.Getenv("DB_PASSWORD"),
		Net:    "tcp",
		Addr:   os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT"),
		DBName: os.Getenv("DB_NAME"),
	}

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		logger.Fatalf("Failed to open database: %v", err)
	}

	// 测试连接
	pingErr := db.Ping()
	if pingErr != nil {
		logger.Fatalf("Failed to ping database: %v", pingErr)
	}

	logger.Info("Database connection established successfully")
	
	// 创建表
	createTables()
}

func createTables() {
	// 创建用户表
	userTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(50) NOT NULL UNIQUE,
		email VARCHAR(100) NOT NULL UNIQUE,
		password_hash VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	);`
	
	_, err := db.Exec(userTableSQL)
	if err != nil {
		logger.Fatalf("Failed to create users table: %v", err)
	}

	// 创建任务表
	taskTableSQL := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(100) NOT NULL,
		description TEXT,
		completed BOOLEAN DEFAULT FALSE,
		user_id INT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`
	
	_, err = db.Exec(taskTableSQL)
	if err != nil {
		logger.Fatalf("Failed to create tasks table: %v", err)
	}

	logger.Info("Tables created or already exist")
}

func main() {
	defer db.Close()

	// 初始化路由
	r := mux.NewRouter()

	// 注册中间件
	r.Use(loggingMiddleware)
	r.Use(recoverMiddleware)

	// 注册路由
	registerRoutes(r)

	// 启动服务器
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	logger.Infof("Server starting on port %s", port)
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}
