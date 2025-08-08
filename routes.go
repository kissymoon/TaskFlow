package main

import (
	"net/http"
	"github.com/gorilla/mux"
)

func registerRoutes(r *mux.Router) {
	// 健康检查路由
	r.HandleFunc("/health", healthCheckHandler).Methods("GET")

	// 用户相关路由
	userRouter := r.PathPrefix("/api/users").Subrouter()
	userRouter.HandleFunc("", createUserHandler).Methods("POST")
	userRouter.HandleFunc("", getAllUsersHandler).Methods("GET")
	userRouter.HandleFunc("/{id:[0-9]+}", getUserHandler).Methods("GET")
	userRouter.HandleFunc("/{id:[0-9]+}", updateUserHandler).Methods("PUT")
	userRouter.HandleFunc("/{id:[0-9]+}", deleteUserHandler).Methods("DELETE")

	// 任务相关路由
	taskRouter := r.PathPrefix("/api/tasks").Subrouter()
	taskRouter.HandleFunc("", createTaskHandler).Methods("POST")
	taskRouter.HandleFunc("", getAllTasksHandler).Methods("GET")
	taskRouter.HandleFunc("/{id:[0-9]+}", getTaskHandler).Methods("GET")
	taskRouter.HandleFunc("/{id:[0-9]+}", updateTaskHandler).Methods("PUT")
	taskRouter.HandleFunc("/{id:[0-9]+}", deleteTaskHandler).Methods("DELETE")
	taskRouter.HandleFunc("/user/{userId:[0-9]+}", getTasksByUserHandler).Methods("GET")
}
