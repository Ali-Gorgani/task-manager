package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Ali-Gorgani/task-manager/internal/models"
	"github.com/Ali-Gorgani/task-manager/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTaskRepository is a mock implementation of TaskRepository
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

func TestCreateTask_Success(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo, nil)

	req := &models.CreateTaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
		Assignee:    "test@example.com",
		Status:      models.TaskStatusPending,
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Task")).Return(nil)

	task, err := service.CreateTask(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, req.Title, task.Title)
	assert.Equal(t, req.Description, task.Description)
	mockRepo.AssertExpectations(t)
}

func TestCreateTask_EmptyTitle(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo, nil)

	req := &models.CreateTaskRequest{
		Title: "",
	}

	task, err := service.CreateTask(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "title is required")
}

func TestCreateTask_InvalidStatus(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo, nil)

	req := &models.CreateTaskRequest{
		Title:  "Test",
		Status: "invalid_status",
	}

	task, err := service.CreateTask(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "invalid status")
}

func TestGetTask_Success(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo, nil)

	expectedTask := models.NewTask("Test", "Desc", "test@example.com", models.TaskStatusPending)
	mockRepo.On("GetByID", mock.Anything, expectedTask.ID).Return(expectedTask, nil)

	task, err := service.GetTask(context.Background(), expectedTask.ID)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, expectedTask.ID, task.ID)
	mockRepo.AssertExpectations(t)
}

func TestGetTask_NotFound(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo, nil)

	mockRepo.On("GetByID", mock.Anything, "non-existent").Return(nil, repository.ErrTaskNotFound)

	task, err := service.GetTask(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Nil(t, task)
	mockRepo.AssertExpectations(t)
}

func TestListTasks_Success(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo, nil)

	tasks := []models.Task{
		*models.NewTask("Task 1", "Desc 1", "user1@example.com", models.TaskStatusPending),
		*models.NewTask("Task 2", "Desc 2", "user2@example.com", models.TaskStatusCompleted),
	}

	filter := &models.TaskFilter{Page: 1, PageSize: 10}
	mockRepo.On("GetAll", mock.Anything, filter).Return(tasks, 2, nil)

	response, err := service.ListTasks(context.Background(), filter)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Tasks, 2)
	assert.Equal(t, 2, response.Total)
	mockRepo.AssertExpectations(t)
}

func TestUpdateTask_Success(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo, nil)

	existingTask := models.NewTask("Old Title", "Old Desc", "old@example.com", models.TaskStatusPending)
	newTitle := "New Title"
	newStatus := models.TaskStatusCompleted

	mockRepo.On("GetByID", mock.Anything, existingTask.ID).Return(existingTask, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).Return(nil)

	req := &models.UpdateTaskRequest{
		Title:  &newTitle,
		Status: &newStatus,
	}

	task, err := service.UpdateTask(context.Background(), existingTask.ID, req)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, newTitle, task.Title)
	assert.Equal(t, newStatus, task.Status)
	mockRepo.AssertExpectations(t)
}

func TestUpdateTask_NotFound(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo, nil)

	mockRepo.On("GetByID", mock.Anything, "non-existent").Return(nil, repository.ErrTaskNotFound)

	req := &models.UpdateTaskRequest{}
	task, err := service.UpdateTask(context.Background(), "non-existent", req)
	assert.Error(t, err)
	assert.Nil(t, task)
	mockRepo.AssertExpectations(t)
}

func TestDeleteTask_Success(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo, nil)

	taskID := "test-id"
	mockRepo.On("Delete", mock.Anything, taskID).Return(nil)

	err := service.DeleteTask(context.Background(), taskID)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteTask_NotFound(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo, nil)

	mockRepo.On("Delete", mock.Anything, "non-existent").Return(repository.ErrTaskNotFound)

	err := service.DeleteTask(context.Background(), "non-existent")
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGetTaskCount(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo, nil)

	mockRepo.On("Count", mock.Anything).Return(42, nil)

	count, err := service.GetTaskCount(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 42, count)
	mockRepo.AssertExpectations(t)
}

func TestGetTaskCount_Error(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo, nil)

	mockRepo.On("Count", mock.Anything).Return(0, errors.New("database error"))

	count, err := service.GetTaskCount(context.Background())
	assert.Error(t, err)
	assert.Equal(t, 0, count)
	mockRepo.AssertExpectations(t)
}

func TestListTasks_InvalidStatus(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo, nil)

	invalidStatus := models.TaskStatus("invalid_status")
	filter := &models.TaskFilter{
		Status:   &invalidStatus,
		Page:     1,
		PageSize: 10,
	}

	response, err := service.ListTasks(context.Background(), filter)
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid status filter")
}

func TestListTasks_DefaultPagination(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo, nil)

	tasks := []models.Task{
		*models.NewTask("Task 1", "Desc 1", "user1@example.com", models.TaskStatusPending),
	}

	// Test with page < 1 and pageSize < 1
	filter := &models.TaskFilter{
		Page:     0,
		PageSize: 0,
	}

	mockRepo.On("GetAll", mock.Anything, mock.MatchedBy(func(f *models.TaskFilter) bool {
		return f.Page == 1 && f.PageSize == 10
	})).Return(tasks, 1, nil)

	response, err := service.ListTasks(context.Background(), filter)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 10, response.PageSize)
	mockRepo.AssertExpectations(t)
}

func TestListTasks_MaxPageSize(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo, nil)

	tasks := []models.Task{}

	// Test with pageSize > 100
	filter := &models.TaskFilter{
		Page:     1,
		PageSize: 200,
	}

	mockRepo.On("GetAll", mock.Anything, mock.MatchedBy(func(f *models.TaskFilter) bool {
		return f.PageSize == 100
	})).Return(tasks, 0, nil)

	response, err := service.ListTasks(context.Background(), filter)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	mockRepo.AssertExpectations(t)
}

func TestListTasks_NilFilter(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo, nil)

	tasks := []models.Task{
		*models.NewTask("Task 1", "Desc 1", "user1@example.com", models.TaskStatusPending),
	}

	mockRepo.On("GetAll", mock.Anything, mock.AnythingOfType("*models.TaskFilter")).Return(tasks, 1, nil)

	response, err := service.ListTasks(context.Background(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Tasks, 1)
	mockRepo.AssertExpectations(t)
}

func TestListTasks_RepositoryError(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo, nil)

	filter := &models.TaskFilter{
		Page:     1,
		PageSize: 10,
	}

	mockRepo.On("GetAll", mock.Anything, mock.AnythingOfType("*models.TaskFilter")).Return([]models.Task{}, 0, errors.New("database error"))

	response, err := service.ListTasks(context.Background(), filter)
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to list tasks")
	mockRepo.AssertExpectations(t)
}

func TestUpdateTask_InvalidStatus(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo, nil)

	existingTask := models.NewTask("Old Title", "Old Desc", "old@example.com", models.TaskStatusPending)
	invalidStatus := models.TaskStatus("invalid_status")

	mockRepo.On("GetByID", mock.Anything, existingTask.ID).Return(existingTask, nil)

	req := &models.UpdateTaskRequest{
		Status: &invalidStatus,
	}

	task, err := service.UpdateTask(context.Background(), existingTask.ID, req)
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "invalid status")
	mockRepo.AssertExpectations(t)
}

func TestUpdateTask_RepositoryError(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo, nil)

	existingTask := models.NewTask("Old Title", "Old Desc", "old@example.com", models.TaskStatusPending)
	newTitle := "New Title"

	mockRepo.On("GetByID", mock.Anything, existingTask.ID).Return(existingTask, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).Return(errors.New("database error"))

	req := &models.UpdateTaskRequest{
		Title: &newTitle,
	}

	task, err := service.UpdateTask(context.Background(), existingTask.ID, req)
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "failed to update task")
	mockRepo.AssertExpectations(t)
}

func TestUpdateTask_AllFields(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo, nil)

	existingTask := models.NewTask("Old Title", "Old Desc", "old@example.com", models.TaskStatusPending)
	newTitle := "New Title"
	newDesc := "New Description"
	newStatus := models.TaskStatusCompleted
	newAssignee := "new@example.com"

	mockRepo.On("GetByID", mock.Anything, existingTask.ID).Return(existingTask, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).Return(nil)

	req := &models.UpdateTaskRequest{
		Title:       &newTitle,
		Description: &newDesc,
		Status:      &newStatus,
		Assignee:    &newAssignee,
	}

	task, err := service.UpdateTask(context.Background(), existingTask.ID, req)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, newTitle, task.Title)
	assert.Equal(t, newDesc, task.Description)
	assert.Equal(t, newStatus, task.Status)
	assert.Equal(t, newAssignee, task.Assignee)
	mockRepo.AssertExpectations(t)
}

func TestCreateTask_RepositoryError(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo, nil)

	req := &models.CreateTaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
		Assignee:    "test@example.com",
		Status:      models.TaskStatusPending,
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Task")).Return(errors.New("database error"))

	task, err := service.CreateTask(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "failed to create task")
	mockRepo.AssertExpectations(t)
}

func TestListTasks_TotalPagesCalculation(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := NewTaskService(mockRepo, nil)

	tasks := []models.Task{}

	filter := &models.TaskFilter{
		Page:     1,
		PageSize: 10,
	}

	// Test when total is 0
	mockRepo.On("GetAll", mock.Anything, mock.AnythingOfType("*models.TaskFilter")).Return(tasks, 0, nil).Once()

	response, err := service.ListTasks(context.Background(), filter)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 1, response.TotalPages) // Should be 1 when total is 0

	// Test with multiple pages
	mockRepo.On("GetAll", mock.Anything, mock.AnythingOfType("*models.TaskFilter")).Return(tasks, 25, nil).Once()

	response2, err2 := service.ListTasks(context.Background(), filter)
	assert.NoError(t, err2)
	assert.NotNil(t, response2)
	assert.Equal(t, 3, response2.TotalPages) // 25 items / 10 per page = 3 pages

	mockRepo.AssertExpectations(t)
}
