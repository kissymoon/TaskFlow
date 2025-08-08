package main

import (
	"net/http"
	"time"
	"fmt"
)

// loggingMiddleware 记录请求日志
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// 创建一个响应写入器来捕获状态码
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// 调用下一个处理器
		next.ServeHTTP(lrw, r)
		
		// 计算请求处理时间
		duration := time.Since(start)
		
		// 记录日志
		logger.WithFields(logrus.Fields{
			"method":     r.Method,
			"path":       r.URL.Path,
			"status":     lrw.statusCode,
			"duration":   duration,
			"remote_addr": r.RemoteAddr,
		}).Info("Request processed")
	})
}

// loggingResponseWriter 用于捕获HTTP响应状态码
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// recoverMiddleware 捕获恐慌并返回适当的错误响应
func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf("Panic recovered: %v", err)
				sendErrorResponse(w, http.StatusInternalServerError, "Internal server error")
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}

// sendErrorResponse 发送标准化的错误响应
func sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := map[string]string{
		"error": message,
		"code":  fmt.Sprintf("%d", statusCode),
	}
	
	json.NewEncoder(w).Encode(response)
}
