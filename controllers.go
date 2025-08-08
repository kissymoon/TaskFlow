package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"golang.org/x/crypto/bcrypt"
	"github.com/gorilla/mux"
)

// 健康检查处理函数
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// 用户控制器

// createUserHandler 创建新用户
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		logger.Errorf("Error decoding user: %v", err)
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 验证请求数据
	if user.Username == "" || user.Email == "" || user.PasswordHash == "" {
		sendErrorResponse(w, http.StatusBadRequest, "Username, email and password are required")
		return
	}

	// 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("Error hashing password: %v", err)
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to process password")
		return
	}
	user.PasswordHash = string(hashedPassword)

	// 创建用户
	err = CreateUser(&user)
	if err != nil {
		logger.Errorf("Error creating user: %v", err)
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// getAllUsersHandler 获取所有用户
func getAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := GetAllUsers()
	if err != nil {
		logger.Errorf("Error getting users: %v", err)
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to get users")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

// getUserHandler 根据ID获取用户
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := GetUserByID(id)
	if err != nil {
		logger.Errorf("Error getting user: %v", err)
		if err.Error() == "user not found" {
			sendErrorResponse(w, http.StatusNotFound, "User not found")
		} else {
			sendErrorResponse(w, http.StatusInternalServerError, "Failed to get user")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// updateUserHandler 更新用户
func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// 检查用户是否存在
	existingUser, err := GetUserByID(id)
	if err != nil {
		logger.Errorf("Error checking user existence: %v", err)
		if err.Error() == "user not found" {
			sendErrorResponse(w, http.StatusNotFound, "User not found")
		} else {
			sendErrorResponse(w, http.StatusInternalServerError, "Failed to check user existence")
		}
		return
	}

	var updatedUser User
	err = json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		logger.Errorf("Error decoding user update: %v", err)
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 保留原ID
	updatedUser.ID = id
	// 如果提供了新密码，则哈希新密码
	if updatedUser.PasswordHash != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updatedUser.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
			logger.Errorf("Error hashing password: %v", err)
			sendErrorResponse(w, http.StatusInternalServerError, "Failed to process password")
			return
		}
		updatedUser.PasswordHash = string(hashedPassword)
	} else {
		// 否则使用原密码
		updatedUser.PasswordHash = existingUser.PasswordHash
	}

	// 更新用户
	err = UpdateUser(&updatedUser)
	if err != nil {
		logger.Errorf("Error updating user: %v", err)
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	// 返回更新后的用户信息（不含密码）
	resultUser, err := GetUserByID(id)
	if err != nil {
		logger.Errorf("Error getting updated user: %v", err)
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to get updated user")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resultUser)
}

// deleteUserHandler 删除用户
func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// 检查用户是否存在
	_, err = GetUserByID(id)
	if err != nil {
		logger.Errorf("Error checking user existence: %v", err)
		if err.Error() == "user not found" {
			sendErrorResponse(w, http.StatusNotFound, "User not found")
		} else {
			sendErrorResponse(w, http.StatusInternalServerError, "Failed to check user existence")
		}
		return
	}

	// 删除用户
	err = DeleteUser(id)
	if err != nil {
		logger.Errorf("Error deleting user: %v", err)
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// 任务控制器

// createTaskHandler 创建新任务
func createTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		logger.Errorf("Error decoding task: %v", err)
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 验证请求数据
	if task.Title == "" || task.UserID == 0 {
		sendErrorResponse(w, http.StatusBadRequest, "Title and user ID are required")
		return
	}

	// 检查用户是否存在
	_, err = GetUserByID(task.UserID)
	if err != nil {
		logger.Errorf("Error checking user existence: %v", err)
		if err.Error() == "user not found" {
			sendErrorResponse(w, http.StatusNotFound, "User not found")
		} else {
			sendErrorResponse(w, http.StatusInternalServerError, "Failed to check user existence")
		}
		return
	}

	// 创建任务
	err = CreateTask(&task)
	if err != nil {
		logger.Errorf("Error creating task: %v", err)
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to create task")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// getAllTasksHandler 获取所有任务
func getAllTasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := GetAllTasks()
	if err != nil {
		logger.Errorf("Error getting tasks: %v", err)
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to get tasks")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

// getTaskHandler 根据ID获取任务
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	task, err := GetTaskByID(id)
	if err != nil {
		logger.Errorf("Error getting task: %v", err)
		if err.Error() == "task not found" {
			sendErrorResponse(w, http.StatusNotFound, "Task not found")
		} else {
			sendErrorResponse(w, http.StatusInternalServerError, "Failed to get task")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

// getTasksByUserHandler 根据用户ID获取任务
func getTasksByUserHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID, err := strconv.Atoi(params["userId"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// 检查用户是否存在
	_, err = GetUserByID(userID)
	if err != nil {
		logger.Errorf("Error checking user existence: %v", err)
		if err.Error() == "user not found" {
			sendErrorResponse(w, http.StatusNotFound, "User not found")
		} else {
			sendErrorResponse(w, http.StatusInternalServerError, "Failed to check user existence")
		}
		return
	}

	tasks, err := GetTasksByUserID(userID)
	if err != nil {
		logger.Errorf("Error getting user tasks: %v", err)
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to get user tasks")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

// updateTaskHandler 更新任务
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	// 检查任务是否存在
	_, err = GetTaskByID(id)
	if err != nil {
		logger.Errorf("Error checking task existence: %v", err)
		if err.Error() == "task not found" {
			sendErrorResponse(w, http.StatusNotFound, "Task not found")
		} else {
			sendErrorResponse(w, http.StatusInternalServerError, "Failed to check task existence")
		}
		return
	}

	var updatedTask Task
	err = json.NewDecoder(r.Body).Decode(&updatedTask)
	if err != nil {
		logger.Errorf("Error decoding task update: %v", err)
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 保留原ID
	updatedTask.ID = id

	// 如果指定了新的用户ID，检查用户是否存在
	if updatedTask.UserID != 0 {
		_, err = GetUserByID(updatedTask.UserID)
		if err != nil {
			logger.Errorf("Error checking user existence: %v", err)
			if err.Error() == "user not found" {
				sendErrorResponse(w, http.StatusNotFound, "User not found")
			} else {
				sendErrorResponse(w, http.StatusInternalServerError, "Failed to check user existence")
			}
			return
		}
	}

	// 更新任务
	err = UpdateTask(&updatedTask)
	if err != nil {
		logger.Errorf("Error updating task: %v", err)
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to update task")
		return
	}

	// 返回更新后的任务信息
	resultTask, err := GetTaskByID(id)
	if err != nil {
		logger.Errorf("Error getting updated task: %v", err)
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to get updated task")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resultTask)
}

// deleteTaskHandler 删除任务
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	// 检查任务是否存在
	_, err = GetTaskByID(id)
	if err != nil {
		logger.Errorf("Error checking task existence: %v", err)
		if err.Error() == "task not found" {
			sendErrorResponse(w, http.StatusNotFound, "Task not found")
		} else {
			sendErrorResponse(w, http.StatusInternalServerError, "Failed to check task existence")
		}
		return
	}

	// 删除任务
	err = DeleteTask(id)
	if err != nil {
		logger.Errorf("Error deleting task: %v", err)
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to delete task")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
