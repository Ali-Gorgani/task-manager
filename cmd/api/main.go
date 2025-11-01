package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ali-Gorgani/task-manager/internal/cache"
	"github.com/Ali-Gorgani/task-manager/internal/config"
	"github.com/Ali-Gorgani/task-manager/internal/handlers"
	"github.com/Ali-Gorgani/task-manager/internal/metrics"
	"github.com/Ali-Gorgani/task-manager/internal/repository"
	"github.com/Ali-Gorgani/task-manager/internal/service"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Ali-Gorgani/task-manager/docs" // Swagger docs
)

// @title Task Manager API
// @version 1.0
// @description A RESTful API for managing tasks (to-do items) with full CRUD operations
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3000
// @BasePath /

// @schemes http
func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Set Gin mode
	if !cfg.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL database")

	// Initialize schema
	taskRepo := repository.NewPostgresTaskRepository(db)
	if err := taskRepo.InitSchema(context.Background()); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}
	log.Println("Database schema initialized successfully")

	// Initialize Redis cache
	var redisCache *cache.RedisCache
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisURL,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis connection failed: %v. Running without cache.", err)
		redisCache = nil
	} else {
		redisCache = cache.NewRedisCache(redisClient)
		log.Println("Successfully connected to Redis")
	}

	// Initialize service and handler
	taskService := service.NewTaskService(taskRepo, redisCache)
	taskHandler := handlers.NewTaskHandler(taskService)

	// Setup router
	router := gin.Default()

	// Add Prometheus middleware
	router.Use(metrics.PrometheusMiddleware())

	// Health check
	router.GET("/health", taskHandler.HealthCheck)

	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		tasks := v1.Group("/tasks")
		{
			tasks.POST("", taskHandler.CreateTask)
			tasks.GET("", taskHandler.ListTasks)
			tasks.GET("/:id", taskHandler.GetTask)
			tasks.PUT("/:id", taskHandler.UpdateTask)
			tasks.DELETE("/:id", taskHandler.DeleteTask)
		}
	}

	// Start periodic task count update for metrics
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			count, err := taskService.GetTaskCount(context.Background())
			if err == nil {
				metrics.UpdateTasksCount(count)
			}
		}
	}()

	// Setup HTTP server
	srv := &http.Server{
		Addr:    cfg.GetServerAddress(),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s", cfg.GetServerAddress())
		log.Printf("Swagger documentation available at http://localhost:%s/swagger/index.html", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited successfully")
}
