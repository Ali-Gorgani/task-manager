package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ali-Gorgani/task-manager/internal/models"
	"github.com/Ali-Gorgani/task-manager/internal/repository"
	"github.com/Ali-Gorgani/task-manager/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTaskRepository is a mock implementation for testing
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) Create(ctx context.Context, task *models.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) GetByID(ctx context.Context, id string) (*models.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskRepository) GetAll(ctx context.Context, filter *models.TaskFilter) ([]models.Task, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]models.Task), args.Int(1), args.Error(2)
}

func (m *MockTaskRepository) Update(ctx context.Context, task *models.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTaskRepository) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func setupRouter(taskService *service.TaskService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	handler := NewTaskHandler(taskService)

	router.GET("/health", handler.HealthCheck)
	v1 := router.Group("/api/v1")
	{
		tasks := v1.Group("/tasks")
		{
			tasks.POST("", handler.CreateTask)
			tasks.GET("", handler.ListTasks)
			tasks.GET("/:id", handler.GetTask)
			tasks.PUT("/:id", handler.UpdateTask)
			tasks.DELETE("/:id", handler.DeleteTask)
		}
	}

	return router
}

func TestHealthCheck(t *testing.T) {
	mockService := &service.TaskService{}
	router := setupRouter(mockService)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}

func TestCreateTask_Handler(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	mockService := service.NewTaskService(mockRepo, nil)
	router := setupRouter(mockService)

	t.Run("Success", func(t *testing.T) {
		reqBody := models.CreateTaskRequest{
			Title:       "Test Task",
			Description: "Test Description",
			Assignee:    "test@example.com",
			Status:      models.TaskStatusPending,
		}
		body, _ := json.Marshal(reqBody)

		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Task")).Return(nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Service Error", func(t *testing.T) {
		mockRepo2 := new(MockTaskRepository)
		mockService2 := service.NewTaskService(mockRepo2, nil)
		router2 := setupRouter(mockService2)

		reqBody := models.CreateTaskRequest{
			Title:       "Test Task",
			Description: "Test Description",
			Status:      models.TaskStatusPending,
		}
		body, _ := json.Marshal(reqBody)

		mockRepo2.On("Create", mock.Anything, mock.AnythingOfType("*models.Task")).Return(errors.New("database error"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router2.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo2.AssertExpectations(t)
	})
}

func TestGetTask_Handler(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	mockService := service.NewTaskService(mockRepo, nil)
	router := setupRouter(mockService)

	t.Run("Success", func(t *testing.T) {
		task := models.NewTask("Test Task", "Description", "test@example.com", models.TaskStatusPending)
		mockRepo.On("GetByID", mock.Anything, task.ID).Return(task, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/tasks/"+task.ID, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo2 := new(MockTaskRepository)
		mockService2 := service.NewTaskService(mockRepo2, nil)
		router2 := setupRouter(mockService2)

		mockRepo2.On("GetByID", mock.Anything, "nonexistent").Return(nil, repository.ErrTaskNotFound)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/tasks/nonexistent", nil)
		router2.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo2.AssertExpectations(t)
	})

	t.Run("Internal Error", func(t *testing.T) {
		mockRepo3 := new(MockTaskRepository)
		mockService3 := service.NewTaskService(mockRepo3, nil)
		router3 := setupRouter(mockService3)

		mockRepo3.On("GetByID", mock.Anything, "error-id").Return(nil, errors.New("database error"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/tasks/error-id", nil)
		router3.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo3.AssertExpectations(t)
	})
}

func TestListTasks_Handler(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	mockService := service.NewTaskService(mockRepo, nil)
	router := setupRouter(mockService)

	t.Run("Success", func(t *testing.T) {
		tasks := []models.Task{
			*models.NewTask("Task 1", "Desc 1", "user1@example.com", models.TaskStatusPending),
		}
		mockRepo.On("GetAll", mock.Anything, mock.AnythingOfType("*models.TaskFilter")).Return(tasks, 1, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/tasks", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("With Filters", func(t *testing.T) {
		mockRepo2 := new(MockTaskRepository)
		mockService2 := service.NewTaskService(mockRepo2, nil)
		router2 := setupRouter(mockService2)

		tasks := []models.Task{}
		mockRepo2.On("GetAll", mock.Anything, mock.AnythingOfType("*models.TaskFilter")).Return(tasks, 0, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/tasks?status=pending&page=1&page_size=10", nil)
		router2.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo2.AssertExpectations(t)
	})

	t.Run("Invalid Status", func(t *testing.T) {
		mockRepo3 := new(MockTaskRepository)
		mockService3 := service.NewTaskService(mockRepo3, nil)
		router3 := setupRouter(mockService3)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/tasks?status=invalid_status", nil)
		router3.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdateTask_Handler(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	mockService := service.NewTaskService(mockRepo, nil)
	router := setupRouter(mockService)

	t.Run("Success", func(t *testing.T) {
		task := models.NewTask("Old Title", "Old Desc", "old@example.com", models.TaskStatusPending)
		newTitle := "New Title"

		mockRepo.On("GetByID", mock.Anything, task.ID).Return(task, nil)
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).Return(nil)

		reqBody := models.UpdateTaskRequest{
			Title: &newTitle,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/v1/tasks/"+task.ID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo2 := new(MockTaskRepository)
		mockService2 := service.NewTaskService(mockRepo2, nil)
		router2 := setupRouter(mockService2)

		mockRepo2.On("GetByID", mock.Anything, "nonexistent").Return(nil, repository.ErrTaskNotFound)

		reqBody := models.UpdateTaskRequest{}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/v1/tasks/nonexistent", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router2.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo2.AssertExpectations(t)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/v1/tasks/some-id", bytes.NewBufferString("invalid"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Internal Error", func(t *testing.T) {
		mockRepo3 := new(MockTaskRepository)
		mockService3 := service.NewTaskService(mockRepo3, nil)
		router3 := setupRouter(mockService3)

		task := models.NewTask("Task", "Desc", "user@example.com", models.TaskStatusPending)
		mockRepo3.On("GetByID", mock.Anything, task.ID).Return(task, nil)
		mockRepo3.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).Return(errors.New("db error"))

		reqBody := models.UpdateTaskRequest{}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/v1/tasks/"+task.ID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router3.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo3.AssertExpectations(t)
	})
}

func TestDeleteTask_Handler(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	mockService := service.NewTaskService(mockRepo, nil)
	router := setupRouter(mockService)

	t.Run("Success", func(t *testing.T) {
		taskID := "test-id"
		mockRepo.On("Delete", mock.Anything, taskID).Return(nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/v1/tasks/"+taskID, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo2 := new(MockTaskRepository)
		mockService2 := service.NewTaskService(mockRepo2, nil)
		router2 := setupRouter(mockService2)

		mockRepo2.On("Delete", mock.Anything, "nonexistent").Return(repository.ErrTaskNotFound)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/v1/tasks/nonexistent", nil)
		router2.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo2.AssertExpectations(t)
	})

	t.Run("Internal Error", func(t *testing.T) {
		mockRepo3 := new(MockTaskRepository)
		mockService3 := service.NewTaskService(mockRepo3, nil)
		router3 := setupRouter(mockService3)

		mockRepo3.On("Delete", mock.Anything, "error-id").Return(errors.New("database error"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/v1/tasks/error-id", nil)
		router3.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo3.AssertExpectations(t)
	})
}

func TestNewTaskHandler(t *testing.T) {
	mockService := &service.TaskService{}
	handler := NewTaskHandler(mockService)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.service)
}
