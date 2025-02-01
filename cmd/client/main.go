package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/dimitriirfan/benchmark-grpc-vs-rest-server/entity"
	pb "github.com/dimitriirfan/benchmark-grpc-vs-rest-server/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/proto"
)

const (
	ProtocolRest    = "rest"
	ProtocolGrpc    = "grpc"
	ProtocolGrpcRaw = "grpc-raw"
)

type ClientAnalytics struct {
	Protocol        string        `json:"protocol"`
	TotalRequests   int64         `json:"total_requests"`
	SuccessRequests int64         `json:"success_requests"`
	FailedRequests  int64         `json:"failed_requests"`
	AverageLatency  time.Duration `json:"average_latency"`
	MinLatency      time.Duration `json:"min_latency"`
	MaxLatency      time.Duration `json:"max_latency"`
	TotalLatency    time.Duration `json:"total_latency"`
	TotalBytes      int64         `json:"total_bytes"`
	AverageBodySize float64       `json:"average_body_size"`
	StartTime       time.Time     `json:"start_time"`
	EndTime         time.Time     `json:"end_time"`
	TotalDuration   time.Duration `json:"total_duration"`
	RequestsPerSec  float64       `json:"requests_per_sec"`
	BytesPerSec     float64       `json:"bytes_per_sec"`
	MockSize        int           `json:"mock_size"`
	mu              sync.RWMutex
}

func (a *ClientAnalytics) recordMetrics(latency time.Duration, bodySize int, success bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.TotalRequests++
	if success {
		a.SuccessRequests++
	} else {
		a.FailedRequests++
	}

	a.TotalLatency += latency
	a.TotalBytes += int64(bodySize)

	if latency < a.MinLatency || a.MinLatency == 0 {
		a.MinLatency = latency
	}
	if latency > a.MaxLatency {
		a.MaxLatency = latency
	}

	a.AverageLatency = time.Duration(int64(a.TotalLatency) / a.TotalRequests)
	a.AverageBodySize = float64(a.TotalBytes) / float64(a.TotalRequests)

	a.EndTime = time.Now()
	a.TotalDuration = a.EndTime.Sub(a.StartTime)
	a.RequestsPerSec = float64(a.TotalRequests) / a.TotalDuration.Seconds()
	a.BytesPerSec = float64(a.TotalBytes) / a.TotalDuration.Seconds()
}

var restClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	},
	Timeout: 10 * time.Second,
}

func main() {
	config := entity.Config{}
	if err := env.Parse(&config); err != nil {
		log.Fatalf("Failed to parse environment variables: %v", err)
	}

	analytics := make([]*ClientAnalytics, 0)
	restAnalytics := benchmarkRest(config.MockSize)
	grpcAnalytics := benchmarkGrpc(config.MockSize)
	grpcRawAnalytics := benchmarkGrpcRaw(config.MockSize)

	analytics = append(analytics, restAnalytics, grpcAnalytics, grpcRawAnalytics)

	// to json file
	jsonData, err := json.MarshalIndent(analytics, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal analytics to JSON: %v", err)
	}

	os.WriteFile(config.OutputDir+"/"+config.OutputFile, jsonData, 0644)
}

func benchmarkRest(size int) *ClientAnalytics {
	analytics := &ClientAnalytics{
		Protocol:   ProtocolRest,
		MinLatency: time.Hour,
		StartTime:  time.Now(),
		MockSize:   size,
	}

	// Number of concurrent requests
	concurrency := 100
	// Number of requests per goroutine
	requestsPerClient := 100

	var wg sync.WaitGroup
	wg.Add(concurrency)

	log.Printf("Starting REST benchmark with %d concurrent clients, %d requests each", concurrency, requestsPerClient)

	for i := 0; i < concurrency; i++ {
		go func(clientID int) {
			defer wg.Done()
			for j := 0; j < requestsPerClient; j++ {
				makeRestRequest(analytics)
			}
		}(i)
	}

	wg.Wait()
	printAnalytics(analytics)
	return analytics
}

func benchmarkGrpc(size int) *ClientAnalytics {
	log.Printf("Starting gRPC benchmark")

	// Create gRPC connection with better options
	conn, err := grpc.Dial("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithInitialWindowSize(1<<23), // 8MB window size (up from 1MB)
		grpc.WithInitialConnWindowSize(1<<23),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             20 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create client
	client := pb.NewPopulationServiceClient(conn)
	log.Printf("Created gRPC client")

	// Test single request first
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	log.Printf("Making test request")
	resp, err := client.GetPopulation(ctx, &pb.GetPopulationRequest{})
	if err != nil {
		log.Fatalf("Test request failed: %v", err)
	}
	log.Printf("Test request successful, got %d people", len(resp.Population))

	// Continue with benchmark...
	analytics := &ClientAnalytics{
		Protocol:   ProtocolGrpc,
		MinLatency: time.Hour,
		StartTime:  time.Now(),
		MockSize:   size,
	}

	// Number of concurrent requests
	concurrency := 100
	// Number of requests per goroutine
	requestsPerClient := 100

	var wg sync.WaitGroup
	wg.Add(concurrency)

	log.Printf("Starting gRPC benchmark with %d concurrent clients, %d requests each", concurrency, requestsPerClient)

	for i := 0; i < concurrency; i++ {
		go func(clientID int) {
			defer wg.Done()
			for j := 0; j < requestsPerClient; j++ {
				makeGrpcRequest(client, analytics)
			}
		}(i)
	}

	wg.Wait()
	printAnalytics(analytics)
	return analytics
}

func benchmarkGrpcRaw(size int) *ClientAnalytics {
	log.Printf("Starting gRPC benchmark")

	// Create gRPC connection with better options
	conn, err := grpc.Dial("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithInitialWindowSize(1<<20),     // 1MB window size
		grpc.WithInitialConnWindowSize(1<<20), // 1MB connection window size
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(1024*1024*10), // 10MB max message size
		),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create client
	client := pb.NewPopulationServiceClient(conn)
	log.Printf("Created gRPC client")

	// Test single request first
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	log.Printf("Making test request")
	resp, err := client.GetPopulationRaw(ctx, &pb.GetPopulationRequest{})
	if err != nil {
		log.Fatalf("Test request failed: %v", err)
	}

	population := &pb.GetPopulationResponse{}
	if err := proto.Unmarshal(resp.Data, population); err != nil {
		log.Fatalf("Test request failed: %v", err)
	}

	log.Printf("Test request successful, got %d people", len(population.Population))

	// Continue with benchmark...
	analytics := &ClientAnalytics{
		Protocol:   ProtocolGrpcRaw,
		MinLatency: time.Hour,
		StartTime:  time.Now(),
		MockSize:   size,
	}

	// Number of concurrent requests
	concurrency := 100
	// Number of requests per goroutine
	requestsPerClient := 100

	var wg sync.WaitGroup
	wg.Add(concurrency)

	log.Printf("Starting gRPC benchmark with %d concurrent clients, %d requests each", concurrency, requestsPerClient)

	for i := 0; i < concurrency; i++ {
		go func(clientID int) {
			defer wg.Done()
			for j := 0; j < requestsPerClient; j++ {
				makeGrpcRequestRaw(client, analytics)
			}
		}(i)
	}

	wg.Wait()
	printAnalytics(analytics)
	return analytics
}

func printAnalytics(a *ClientAnalytics) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	fmt.Printf("\nBenchmark Results:\n")
	fmt.Printf("================\n")
	fmt.Printf("Protocol:           %s\n", a.Protocol)
	fmt.Printf("Total Requests:     %d\n", a.TotalRequests)
	fmt.Printf("Success Requests:   %d\n", a.SuccessRequests)
	fmt.Printf("Failed Requests:    %d\n", a.FailedRequests)
	fmt.Printf("Average Latency:    %.2fms\n", float64(a.AverageLatency.Microseconds())/1000)
	fmt.Printf("Min Latency:        %.2fms\n", float64(a.MinLatency.Microseconds())/1000)
	fmt.Printf("Max Latency:        %.2fms\n", float64(a.MaxLatency.Microseconds())/1000)
	fmt.Printf("Total Duration:     %.2fs\n", a.TotalDuration.Seconds())
	fmt.Printf("Requests/sec:       %.2f\n", a.RequestsPerSec)
	fmt.Printf("Average Body Size:  %.2f bytes\n", a.AverageBodySize)
	fmt.Printf("Transfer Rate:      %.2f MB/sec\n", a.BytesPerSec/1024/1024)
}

func makeRestRequest(analytics *ClientAnalytics) {
	startTime := time.Now()

	resp, err := restClient.Get("http://localhost:8080/benchmark")
	if err != nil {
		analytics.recordMetrics(time.Since(startTime), 0, false)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		analytics.recordMetrics(time.Since(startTime), 0, false)
		return
	}

	// Parse JSON but don't use the result
	var population entity.GetPopulationResponse
	if err := json.Unmarshal(body, &population); err != nil {
		analytics.recordMetrics(time.Since(startTime), 0, false)
		return
	}

	latency := time.Since(startTime)
	analytics.recordMetrics(latency, len(body), resp.StatusCode == http.StatusOK)
}

func makeGrpcRequest(client pb.PopulationServiceClient, analytics *ClientAnalytics) {
	startTime := time.Now()

	resp, err := client.GetPopulation(context.Background(), &pb.GetPopulationRequest{})
	if err != nil {
		analytics.recordMetrics(time.Since(startTime), 0, false)
		return
	}

	latency := time.Since(startTime)
	analytics.recordMetrics(latency, proto.Size(resp), true)
}

func makeGrpcRequestRaw(client pb.PopulationServiceClient, analytics *ClientAnalytics) {
	startTime := time.Now()

	resp, err := client.GetPopulationRaw(context.Background(), &pb.GetPopulationRequest{})
	if err != nil {
		analytics.recordMetrics(time.Since(startTime), 0, false)
		return
	}

	population := &pb.GetPopulationResponse{}
	if err := proto.Unmarshal(resp.Data, population); err != nil {
		analytics.recordMetrics(time.Since(startTime), 0, false)
		return
	}

	latency := time.Since(startTime)
	analytics.recordMetrics(latency, proto.Size(resp), true)
}
