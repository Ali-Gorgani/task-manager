package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/Ali-Gorgani/task-manager/internal/models"
	_ "github.com/lib/pq"
)

// setupBenchmarkDB creates a test database connection for benchmarks
func setupBenchmarkDB(b *testing.B) (*sql.DB, *PostgresTaskRepository) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/taskmanager?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		b.Skipf("Skipping benchmark: could not connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		b.Skipf("Skipping benchmark: database not available: %v", err)
	}

	repo := NewPostgresTaskRepository(db)
	if err := repo.InitSchema(context.Background()); err != nil {
		b.Skipf("Skipping benchmark: could not initialize schema: %v", err)
	}

	// Clean up test data
	_, _ = db.Exec("DELETE FROM tasks WHERE title LIKE 'Benchmark%'")

	return db, repo
}

func BenchmarkPostgresCreate(b *testing.B) {
	db, repo := setupBenchmarkDB(b)
	defer db.Close()

	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		task := models.NewTask(
			fmt.Sprintf("Benchmark Task %d", i),
			"Benchmark description",
			"benchmark@example.com",
			models.TaskStatusPending,
		)
		_ = repo.Create(ctx, task)
	}

	b.StopTimer()
	// Cleanup
	_, _ = db.Exec("DELETE FROM tasks WHERE title LIKE 'Benchmark%'")
}

func BenchmarkPostgresGetByID(b *testing.B) {
	db, repo := setupBenchmarkDB(b)
	defer db.Close()

	ctx := context.Background()

	// Create a task to benchmark retrieval
	task := models.NewTask(
		"Benchmark GetByID Task",
		"Description for GetByID benchmark",
		"benchmark@example.com",
		models.TaskStatusPending,
	)
	if err := repo.Create(ctx, task); err != nil {
		b.Fatalf("Failed to create test task: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByID(ctx, task.ID)
	}

	b.StopTimer()
	// Cleanup
	_ = repo.Delete(ctx, task.ID)
}

func BenchmarkPostgresGetAll(b *testing.B) {
	db, repo := setupBenchmarkDB(b)
	defer db.Close()

	ctx := context.Background()

	// Create 100 tasks for realistic benchmark
	taskIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		task := models.NewTask(
			fmt.Sprintf("Benchmark GetAll Task %d", i),
			"Description",
			"benchmark@example.com",
			models.TaskStatusPending,
		)
		_ = repo.Create(ctx, task)
		taskIDs[i] = task.ID
	}

	filter := &models.TaskFilter{
		Page:     1,
		PageSize: 10,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _ = repo.GetAll(ctx, filter)
	}

	b.StopTimer()
	// Cleanup
	for _, id := range taskIDs {
		_ = repo.Delete(ctx, id)
	}
}

func BenchmarkPostgresGetAllWithFilter(b *testing.B) {
	db, repo := setupBenchmarkDB(b)
	defer db.Close()

	ctx := context.Background()

	// Create 100 tasks with different statuses
	taskIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		status := models.TaskStatusPending
		if i%2 == 0 {
			status = models.TaskStatusCompleted
		}
		task := models.NewTask(
			fmt.Sprintf("Benchmark Filter Task %d", i),
			"Description",
			"benchmark@example.com",
			status,
		)
		_ = repo.Create(ctx, task)
		taskIDs[i] = task.ID
	}

	status := models.TaskStatusPending
	filter := &models.TaskFilter{
		Status:   &status,
		Page:     1,
		PageSize: 10,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _ = repo.GetAll(ctx, filter)
	}

	b.StopTimer()
	// Cleanup
	for _, id := range taskIDs {
		_ = repo.Delete(ctx, id)
	}
}

func BenchmarkPostgresUpdate(b *testing.B) {
	db, repo := setupBenchmarkDB(b)
	defer db.Close()

	ctx := context.Background()

	// Create a task to benchmark updates
	task := models.NewTask(
		"Benchmark Update Task",
		"Description for update benchmark",
		"benchmark@example.com",
		models.TaskStatusPending,
	)
	if err := repo.Create(ctx, task); err != nil {
		b.Fatalf("Failed to create test task: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Toggle status between pending and in_progress
		if i%2 == 0 {
			task.Status = models.TaskStatusInProgress
		} else {
			task.Status = models.TaskStatusPending
		}
		_ = repo.Update(ctx, task)
	}

	b.StopTimer()
	// Cleanup
	_ = repo.Delete(ctx, task.ID)
}

func BenchmarkPostgresDelete(b *testing.B) {
	db, repo := setupBenchmarkDB(b)
	defer db.Close()

	ctx := context.Background()

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Create a task to delete
		task := models.NewTask(
			fmt.Sprintf("Benchmark Delete Task %d", i),
			"Description",
			"benchmark@example.com",
			models.TaskStatusPending,
		)
		_ = repo.Create(ctx, task)
		b.StartTimer()

		_ = repo.Delete(ctx, task.ID)
	}
}

func BenchmarkPostgresCount(b *testing.B) {
	db, repo := setupBenchmarkDB(b)
	defer db.Close()

	ctx := context.Background()

	// Create some tasks
	taskIDs := make([]string, 50)
	for i := 0; i < 50; i++ {
		task := models.NewTask(
			fmt.Sprintf("Benchmark Count Task %d", i),
			"Description",
			"benchmark@example.com",
			models.TaskStatusPending,
		)
		_ = repo.Create(ctx, task)
		taskIDs[i] = task.ID
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = repo.Count(ctx)
	}

	b.StopTimer()
	// Cleanup
	for _, id := range taskIDs {
		_ = repo.Delete(ctx, id)
	}
}

// Benchmark concurrent operations
func BenchmarkPostgresConcurrentReads(b *testing.B) {
	db, repo := setupBenchmarkDB(b)
	defer db.Close()

	ctx := context.Background()

	// Create a task
	task := models.NewTask(
		"Benchmark Concurrent Task",
		"Description",
		"benchmark@example.com",
		models.TaskStatusPending,
	)
	_ = repo.Create(ctx, task)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = repo.GetByID(ctx, task.ID)
		}
	})

	b.StopTimer()
	_ = repo.Delete(ctx, task.ID)
}

func BenchmarkPostgresConcurrentWrites(b *testing.B) {
	db, repo := setupBenchmarkDB(b)
	defer db.Close()

	ctx := context.Background()
	taskIDs := make(chan string, b.N)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			task := models.NewTask(
				fmt.Sprintf("Concurrent Task %d", i),
				"Description",
				"benchmark@example.com",
				models.TaskStatusPending,
			)
			_ = repo.Create(ctx, task)
			taskIDs <- task.ID
			i++
		}
	})

	b.StopTimer()
	// Cleanup
	close(taskIDs)
	for id := range taskIDs {
		_ = repo.Delete(ctx, id)
	}
}

// Benchmark model operations (no database)
func BenchmarkTaskCreation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		task := models.NewTask(
			"Benchmark Task",
			"Description",
			"benchmark@example.com",
			models.TaskStatusPending,
		)
		_ = task
	}
}

func BenchmarkTaskFilterCreation(b *testing.B) {
	status := models.TaskStatusPending
	assignee := "test@example.com"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		filter := &models.TaskFilter{
			Status:   &status,
			Assignee: &assignee,
			Page:     1,
			PageSize: 10,
		}
		_ = filter
	}
}

// Benchmark pagination with different page sizes
func BenchmarkPostgresPagination(b *testing.B) {
	db, repo := setupBenchmarkDB(b)
	defer db.Close()

	ctx := context.Background()

	// Create 1000 tasks
	taskIDs := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		task := models.NewTask(
			fmt.Sprintf("Benchmark Pagination Task %d", i),
			"Description",
			"benchmark@example.com",
			models.TaskStatusPending,
		)
		_ = repo.Create(ctx, task)
		taskIDs[i] = task.ID
	}

	benchmarks := []struct {
		name     string
		pageSize int
	}{
		{"PageSize10", 10},
		{"PageSize50", 50},
		{"PageSize100", 100},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			filter := &models.TaskFilter{
				Page:     1,
				PageSize: bm.pageSize,
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _, _ = repo.GetAll(ctx, filter)
			}
		})
	}

	b.StopTimer()
	// Cleanup
	for _, id := range taskIDs {
		_ = repo.Delete(ctx, id)
	}
}

// Benchmark with different query patterns
func BenchmarkPostgresQueryPatterns(b *testing.B) {
	db, repo := setupBenchmarkDB(b)
	defer db.Close()

	ctx := context.Background()

	// Create diverse dataset
	taskIDs := make([]string, 200)
	for i := 0; i < 200; i++ {
		status := models.TaskStatusPending
		assignee := fmt.Sprintf("user%d@example.com", i%5)

		switch i % 4 {
		case 0:
			status = models.TaskStatusPending
		case 1:
			status = models.TaskStatusInProgress
		case 2:
			status = models.TaskStatusCompleted
		case 3:
			status = models.TaskStatusCancelled
		}

		task := models.NewTask(
			fmt.Sprintf("Query Pattern Task %d", i),
			"Description",
			assignee,
			status,
		)
		_ = repo.Create(ctx, task)
		taskIDs[i] = task.ID
	}

	status := models.TaskStatusPending
	assignee := "user1@example.com"

	b.Run("NoFilter", func(b *testing.B) {
		filter := &models.TaskFilter{Page: 1, PageSize: 10}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = repo.GetAll(ctx, filter)
		}
	})

	b.Run("StatusFilter", func(b *testing.B) {
		filter := &models.TaskFilter{
			Status:   &status,
			Page:     1,
			PageSize: 10,
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = repo.GetAll(ctx, filter)
		}
	})

	b.Run("AssigneeFilter", func(b *testing.B) {
		filter := &models.TaskFilter{
			Assignee: &assignee,
			Page:     1,
			PageSize: 10,
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = repo.GetAll(ctx, filter)
		}
	})

	b.Run("CombinedFilter", func(b *testing.B) {
		filter := &models.TaskFilter{
			Status:   &status,
			Assignee: &assignee,
			Page:     1,
			PageSize: 10,
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = repo.GetAll(ctx, filter)
		}
	})

	b.StopTimer()
	// Cleanup
	for _, id := range taskIDs {
		_ = repo.Delete(ctx, id)
	}
}
