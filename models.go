package main

import (
	"database/sql"
	"errors"
	"time"
)

// User 模型定义
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // 不返回密码哈希
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Task 模型定义
type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	UserID      int       `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateUser 创建新用户
func CreateUser(user *User) error {
	result, err := db.Exec(
		"INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)",
		user.Username, user.Email, user.PasswordHash,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = int(id)
	return nil
}

// GetAllUsers 获取所有用户
func GetAllUsers() ([]User, error) {
	rows, err := db.Query("SELECT id, username, email, created_at, updated_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// GetUserByID 根据ID获取用户
func GetUserByID(id int) (*User, error) {
	var user User
	err := db.QueryRow(
		"SELECT id, username, email, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// UpdateUser 更新用户信息
func UpdateUser(user *User) error {
	_, err := db.Exec(
		"UPDATE users SET username = ?, email = ?, password_hash = ? WHERE id = ?",
		user.Username, user.Email, user.PasswordHash, user.ID,
	)
	return err
}

// DeleteUser 删除用户
func DeleteUser(id int) error {
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

// CreateTask 创建新任务
func CreateTask(task *Task) error {
	result, err := db.Exec(
		"INSERT INTO tasks (title, description, completed, user_id) VALUES (?, ?, ?, ?)",
		task.Title, task.Description, task.Completed, task.UserID,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	task.ID = int(id)
	return nil
}

// GetAllTasks 获取所有任务
func GetAllTasks() ([]Task, error) {
	rows, err := db.Query("SELECT id, title, description, completed, user_id, created_at, updated_at FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Completed,
			&task.UserID, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetTaskByID 根据ID获取任务
func GetTaskByID(id int) (*Task, error) {
	var task Task
	err := db.QueryRow(
		"SELECT id, title, description, completed, user_id, created_at, updated_at FROM tasks WHERE id = ?",
		id,
	).Scan(
		&task.ID, &task.Title, &task.Description, &task.Completed,
		&task.UserID, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("task not found")
		}
		return nil, err
	}

	return &task, nil
}

// GetTasksByUserID 根据用户ID获取任务
func GetTasksByUserID(userID int) ([]Task, error) {
	rows, err := db.Query(
		"SELECT id, title, description, completed, user_id, created_at, updated_at FROM tasks WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Completed,
			&task.UserID, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// UpdateTask 更新任务
func UpdateTask(task *Task) error {
	_, err := db.Exec(
		"UPDATE tasks SET title = ?, description = ?, completed = ?, user_id = ? WHERE id = ?",
		task.Title, task.Description, task.Completed, task.UserID, task.ID,
	)
	return err
}

// DeleteTask 删除任务
func DeleteTask(id int) error {
	_, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}
