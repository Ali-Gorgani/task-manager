package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/Ali-Gorgani/task-manager/internal/models"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return db, mock
}

func TestCreate(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewPostgresTaskRepository(db)
	task := models.NewTask("Test Task", "Description", "test@example.com", models.TaskStatusPending)

	mock.ExpectExec("INSERT INTO tasks").
		WithArgs(task.ID, task.Title, task.Description, task.Status, task.Assignee, task.CreatedAt, task.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), task)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewPostgresTaskRepository(db)
	expectedTask := models.NewTask("Test Task", "Description", "test@example.com", models.TaskStatusPending)

	rows := sqlmock.NewRows([]string{"id", "title", "description", "status", "assignee", "created_at", "updated_at"}).
		AddRow(expectedTask.ID, expectedTask.Title, expectedTask.Description, expectedTask.Status, expectedTask.Assignee, expectedTask.CreatedAt, expectedTask.UpdatedAt)

	mock.ExpectQuery("SELECT (.+) FROM tasks WHERE id = \\$1").
		WithArgs(expectedTask.ID).
		WillReturnRows(rows)

	task, err := repo.GetByID(context.Background(), expectedTask.ID)
	assert.NoError(t, err)
	assert.Equal(t, expectedTask.ID, task.ID)
	assert.Equal(t, expectedTask.Title, task.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewPostgresTaskRepository(db)

	mock.ExpectQuery("SELECT (.+) FROM tasks WHERE id = \\$1").
		WithArgs("non-existent-id").
		WillReturnError(sql.ErrNoRows)

	task, err := repo.GetByID(context.Background(), "non-existent-id")
	assert.Error(t, err)
	assert.Equal(t, ErrTaskNotFound, err)
	assert.Nil(t, task)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAll_WithFilters(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewPostgresTaskRepository(db)
	status := models.TaskStatusPending
	filter := &models.TaskFilter{
		Status:   &status,
		Page:     1,
		PageSize: 10,
	}

	// Mock count query
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM tasks WHERE status = \\$1").
		WithArgs(status).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// Mock select query
	task := models.NewTask("Test", "Desc", "test@example.com", status)
	rows := sqlmock.NewRows([]string{"id", "title", "description", "status", "assignee", "created_at", "updated_at"}).
		AddRow(task.ID, task.Title, task.Description, task.Status, task.Assignee, task.CreatedAt, task.UpdatedAt)

	mock.ExpectQuery("SELECT (.+) FROM tasks WHERE status = \\$1 ORDER BY created_at DESC LIMIT \\$2 OFFSET \\$3").
		WithArgs(status, 10, 0).
		WillReturnRows(rows)

	tasks, total, err := repo.GetAll(context.Background(), filter)
	assert.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, tasks, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewPostgresTaskRepository(db)
	task := models.NewTask("Updated Task", "Updated Desc", "test@example.com", models.TaskStatusCompleted)

	mock.ExpectExec("UPDATE tasks SET").
		WithArgs(task.Title, task.Description, task.Status, task.Assignee, task.UpdatedAt, task.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Update(context.Background(), task)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewPostgresTaskRepository(db)
	task := models.NewTask("Task", "Desc", "test@example.com", models.TaskStatusPending)

	mock.ExpectExec("UPDATE tasks SET").
		WithArgs(task.Title, task.Description, task.Status, task.Assignee, task.UpdatedAt, task.ID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Update(context.Background(), task)
	assert.Error(t, err)
	assert.Equal(t, ErrTaskNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewPostgresTaskRepository(db)
	taskID := "test-id"

	mock.ExpectExec("DELETE FROM tasks WHERE id = \\$1").
		WithArgs(taskID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), taskID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewPostgresTaskRepository(db)
	taskID := "non-existent"

	mock.ExpectExec("DELETE FROM tasks WHERE id = \\$1").
		WithArgs(taskID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), taskID)
	assert.Error(t, err)
	assert.Equal(t, ErrTaskNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCount(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewPostgresTaskRepository(db)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM tasks").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(42))

	count, err := repo.Count(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 42, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInitSchema(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewPostgresTaskRepository(db)

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS tasks").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.InitSchema(context.Background())
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInitSchema_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewPostgresTaskRepository(db)

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS tasks").
		WillReturnError(sql.ErrConnDone)

	err := repo.InitSchema(context.Background())
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAll_NoFilters(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewPostgresTaskRepository(db)
	filter := &models.TaskFilter{
		Page:     1,
		PageSize: 10,
	}

	// Mock count query
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM tasks").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Mock select query
	task1 := models.NewTask("Task 1", "Desc 1", "test1@example.com", models.TaskStatusPending)
	task2 := models.NewTask("Task 2", "Desc 2", "test2@example.com", models.TaskStatusCompleted)
	rows := sqlmock.NewRows([]string{"id", "title", "description", "status", "assignee", "created_at", "updated_at"}).
		AddRow(task1.ID, task1.Title, task1.Description, task1.Status, task1.Assignee, task1.CreatedAt, task1.UpdatedAt).
		AddRow(task2.ID, task2.Title, task2.Description, task2.Status, task2.Assignee, task2.CreatedAt, task2.UpdatedAt)

	mock.ExpectQuery("SELECT (.+) FROM tasks ORDER BY created_at DESC LIMIT \\$1 OFFSET \\$2").
		WithArgs(10, 0).
		WillReturnRows(rows)

	tasks, total, err := repo.GetAll(context.Background(), filter)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, tasks, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAll_WithAssigneeFilter(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewPostgresTaskRepository(db)
	assignee := "test@example.com"
	filter := &models.TaskFilter{
		Assignee: &assignee,
		Page:     1,
		PageSize: 10,
	}

	// Mock count query
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM tasks WHERE assignee = \\$1").
		WithArgs(assignee).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Mock select query
	task := models.NewTask("Test", "Desc", assignee, models.TaskStatusPending)
	rows := sqlmock.NewRows([]string{"id", "title", "description", "status", "assignee", "created_at", "updated_at"}).
		AddRow(task.ID, task.Title, task.Description, task.Status, task.Assignee, task.CreatedAt, task.UpdatedAt)

	mock.ExpectQuery("SELECT (.+) FROM tasks WHERE assignee = \\$1 ORDER BY created_at DESC LIMIT \\$2 OFFSET \\$3").
		WithArgs(assignee, 10, 0).
		WillReturnRows(rows)

	tasks, total, err := repo.GetAll(context.Background(), filter)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, tasks, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAll_WithBothFilters(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewPostgresTaskRepository(db)
	status := models.TaskStatusCompleted
	assignee := "test@example.com"
	filter := &models.TaskFilter{
		Status:   &status,
		Assignee: &assignee,
		Page:     2,
		PageSize: 5,
	}

	// Mock count query
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM tasks WHERE status = \\$1 AND assignee = \\$2").
		WithArgs(status, assignee).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	// Mock select query
	rows := sqlmock.NewRows([]string{"id", "title", "description", "status", "assignee", "created_at", "updated_at"})

	mock.ExpectQuery("SELECT (.+) FROM tasks WHERE status = \\$1 AND assignee = \\$2 ORDER BY created_at DESC LIMIT \\$3 OFFSET \\$4").
		WithArgs(status, assignee, 5, 5).
		WillReturnRows(rows)

	tasks, total, err := repo.GetAll(context.Background(), filter)
	assert.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, tasks, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAll_CountError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewPostgresTaskRepository(db)
	filter := &models.TaskFilter{
		Page:     1,
		PageSize: 10,
	}

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM tasks").
		WillReturnError(sql.ErrConnDone)

	tasks, total, err := repo.GetAll(context.Background(), filter)
	assert.Error(t, err)
	assert.Equal(t, 0, total)
	assert.Nil(t, tasks)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAll_QueryError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewPostgresTaskRepository(db)
	filter := &models.TaskFilter{
		Page:     1,
		PageSize: 10,
	}

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM tasks").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	mock.ExpectQuery("SELECT (.+) FROM tasks ORDER BY created_at DESC LIMIT \\$1 OFFSET \\$2").
		WithArgs(10, 0).
		WillReturnError(sql.ErrConnDone)

	tasks, total, err := repo.GetAll(context.Background(), filter)
	assert.Error(t, err)
	assert.Equal(t, 0, total)
	assert.Nil(t, tasks)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewPostgresTaskRepository(db)
	task := models.NewTask("Test Task", "Description", "test@example.com", models.TaskStatusPending)

	mock.ExpectExec("INSERT INTO tasks").
		WithArgs(task.ID, task.Title, task.Description, task.Status, task.Assignee, task.CreatedAt, task.UpdatedAt).
		WillReturnError(sql.ErrConnDone)

	err := repo.Create(context.Background(), task)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewPostgresTaskRepository(db)

	mock.ExpectQuery("SELECT (.+) FROM tasks WHERE id = \\$1").
		WithArgs("error-id").
		WillReturnError(sql.ErrConnDone)

	task, err := repo.GetByID(context.Background(), "error-id")
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewPostgresTaskRepository(db)
	task := models.NewTask("Task", "Desc", "test@example.com", models.TaskStatusPending)

	mock.ExpectExec("UPDATE tasks SET").
		WithArgs(task.Title, task.Description, task.Status, task.Assignee, task.UpdatedAt, task.ID).
		WillReturnError(sql.ErrConnDone)

	err := repo.Update(context.Background(), task)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewPostgresTaskRepository(db)

	mock.ExpectExec("DELETE FROM tasks WHERE id = \\$1").
		WithArgs("error-id").
		WillReturnError(sql.ErrConnDone)

	err := repo.Delete(context.Background(), "error-id")
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCount_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewPostgresTaskRepository(db)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM tasks").
		WillReturnError(sql.ErrConnDone)

	count, err := repo.Count(context.Background())
	assert.Error(t, err)
	assert.Equal(t, 0, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}
