package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const (
	baseURL        = "http://localhost:3000"
	numWorkers     = 50
	numRequests    = 1000
	requestTimeout = 10 * time.Second
)

type LoadTestResult struct {
	TotalRequests     int
	SuccessfulReqs    int
	FailedReqs        int
	TotalDuration     time.Duration
	AvgResponseTime   time.Duration
	MinResponseTime   time.Duration
	MaxResponseTime   time.Duration
	RequestsPerSecond float64
}

type TaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Assignee    string `json:"assignee"`
}

func main() {
	fmt.Println("🚀 Starting Load Test for Task Manager API")
	fmt.Println("===========================================")
	fmt.Printf("Base URL: %s\n", baseURL)
	fmt.Printf("Workers: %d\n", numWorkers)
	fmt.Printf("Total Requests: %d\n\n", numRequests)

	// Wait for service to be ready
	fmt.Println("Checking if service is ready...")
	if !waitForService() {
		fmt.Println("❌ Service is not responding. Please start the service first.")
		return
	}
	fmt.Println("✅ Service is ready!")

	// Run load tests
	fmt.Println("Running load tests...")

	fmt.Println("Test 1: Create Tasks")
	createResult := runLoadTest("POST", "/api/v1/tasks", true)
	printResults(createResult)

	time.Sleep(2 * time.Second)

	fmt.Println("\nTest 2: Get All Tasks")
	listResult := runLoadTest("GET", "/api/v1/tasks", false)
	printResults(listResult)

	time.Sleep(2 * time.Second)

	fmt.Println("\nTest 3: Get Tasks with Filtering")
	filterResult := runLoadTest("GET", "/api/v1/tasks?status=pending&page=1&page_size=10", false)
	printResults(filterResult)

	fmt.Println("\n===========================================")
	fmt.Println("✅ Load test completed!")
	fmt.Println("\nView Prometheus metrics at: http://localhost:9090")
	fmt.Println("View service metrics at: http://localhost:3000/metrics")
}

func waitForService() bool {
	for i := 0; i < 10; i++ {
		resp, err := http.Get(baseURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return true
		}
		time.Sleep(1 * time.Second)
	}
	return false
}

func runLoadTest(method, path string, includeBody bool) LoadTestResult {
	startTime := time.Now()

	var wg sync.WaitGroup
	requestsChan := make(chan int, numRequests)
	resultsChan := make(chan time.Duration, numRequests)
	errorsChan := make(chan error, numRequests)

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(&wg, requestsChan, resultsChan, errorsChan, method, path, includeBody)
	}

	// Send requests
	for i := 0; i < numRequests; i++ {
		requestsChan <- i
	}
	close(requestsChan)

	// Wait for all workers to finish
	wg.Wait()
	close(resultsChan)
	close(errorsChan)

	// Collect results
	var responseTimes []time.Duration
	for duration := range resultsChan {
		responseTimes = append(responseTimes, duration)
	}

	failedCount := 0
	for range errorsChan {
		failedCount++
	}

	totalDuration := time.Since(startTime)
	result := LoadTestResult{
		TotalRequests:  numRequests,
		SuccessfulReqs: len(responseTimes),
		FailedReqs:     failedCount,
		TotalDuration:  totalDuration,
	}

	if len(responseTimes) > 0 {
		result.MinResponseTime = responseTimes[0]
		result.MaxResponseTime = responseTimes[0]
		var totalTime time.Duration

		for _, rt := range responseTimes {
			totalTime += rt
			if rt < result.MinResponseTime {
				result.MinResponseTime = rt
			}
			if rt > result.MaxResponseTime {
				result.MaxResponseTime = rt
			}
		}

		result.AvgResponseTime = totalTime / time.Duration(len(responseTimes))
		result.RequestsPerSecond = float64(result.SuccessfulReqs) / totalDuration.Seconds()
	}

	return result
}

func worker(wg *sync.WaitGroup, requests <-chan int, results chan<- time.Duration, errors chan<- error, method, path string, includeBody bool) {
	defer wg.Done()

	client := &http.Client{
		Timeout: requestTimeout,
	}

	for range requests {
		start := time.Now()

		var req *http.Request
		var err error

		if includeBody && method == "POST" {
			task := TaskRequest{
				Title:       fmt.Sprintf("Load Test Task %d", rand.Intn(10000)),
				Description: "This is a load test task",
				Status:      "pending",
				Assignee:    fmt.Sprintf("user%d@example.com", rand.Intn(100)),
			}
			body, _ := json.Marshal(task)
			req, err = http.NewRequest(method, baseURL+path, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req, err = http.NewRequest(method, baseURL+path, nil)
		}

		if err != nil {
			errors <- err
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			errors <- err
			continue
		}

		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		duration := time.Since(start)

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			results <- duration
		} else {
			errors <- fmt.Errorf("status code: %d", resp.StatusCode)
		}
	}
}

func printResults(result LoadTestResult) {
	fmt.Printf("Total Requests:       %d\n", result.TotalRequests)
	fmt.Printf("Successful Requests:  %d (%.2f%%)\n", result.SuccessfulReqs, float64(result.SuccessfulReqs)/float64(result.TotalRequests)*100)
	fmt.Printf("Failed Requests:      %d (%.2f%%)\n", result.FailedReqs, float64(result.FailedReqs)/float64(result.TotalRequests)*100)
	fmt.Printf("Total Duration:       %v\n", result.TotalDuration)
	fmt.Printf("Avg Response Time:    %v\n", result.AvgResponseTime)
	fmt.Printf("Min Response Time:    %v\n", result.MinResponseTime)
	fmt.Printf("Max Response Time:    %v\n", result.MaxResponseTime)
	fmt.Printf("Requests/Second:      %.2f\n", result.RequestsPerSecond)
}
