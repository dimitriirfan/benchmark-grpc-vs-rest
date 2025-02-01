package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/dimitriirfan/benchmark-grpc-vs-rest-server/entity"
	pb "github.com/dimitriirfan/benchmark-grpc-vs-rest-server/proto"
	"github.com/dimitriirfan/benchmark-grpc-vs-rest-server/testutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/proto"
)

// Global response cache for both servers
var (
	globalJSONResponse *entity.GetPopulationResponse
	globalPBResponse   *pb.GetPopulationResponse
	globalRawData      []byte
)

// Initialize loads the data at startup
func initialize(size int) error {
	// Generate fixtures for different sizes
	testutil.GenerateFixtures([]int{size})

	// Load JSON data
	jsonData, err := os.ReadFile(fmt.Sprintf("testutil/fixtures/fixtures_population_%d.json", size))
	if err != nil {
		return err
	}
	globalJSONResponse = &entity.GetPopulationResponse{}
	if err := json.Unmarshal(jsonData, globalJSONResponse); err != nil {
		return err
	}

	// Load protobuf data
	pbData, err := os.ReadFile(fmt.Sprintf("testutil/fixtures/fixtures_population_%d.pb", size))
	if err != nil {
		return err
	}
	globalRawData = pbData
	globalPBResponse = &pb.GetPopulationResponse{}
	if err := proto.Unmarshal(pbData, globalPBResponse); err != nil {
		return err
	}

	return nil
}

// REST handler
func handleGetBenchmark(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(globalJSONResponse)
}

// gRPC server implementation
type grpcServer struct {
	pb.UnimplementedPopulationServiceServer
}

func (s *grpcServer) GetPopulation(ctx context.Context, req *pb.GetPopulationRequest) (*pb.GetPopulationResponse, error) {
	return globalPBResponse, nil
}

func (s *grpcServer) GetPopulationRaw(ctx context.Context, req *pb.GetPopulationRequest) (*pb.RawResponse, error) {
	return &pb.RawResponse{Data: globalRawData}, nil
}

func main() {
	config := entity.Config{}
	if err := env.Parse(&config); err != nil {
		log.Fatalf("Failed to parse environment variables: %v", err)
	}

	// Load data at startup
	if err := initialize(config.MockSize); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	// Create REST server
	handler := http.NewServeMux()
	handler.HandleFunc("/benchmark", handleGetBenchmark)

	restServer := &http.Server{
		Addr:              ":8080",
		Handler:           handler,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	// Create gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcSrv := grpc.NewServer(
		grpc.MaxRecvMsgSize(1024*1024*10),
		grpc.MaxSendMsgSize(1024*1024*10),
		grpc.MaxConcurrentStreams(100000),
		grpc.NumStreamWorkers(32),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Minute,
			MaxConnectionAgeGrace: 5 * time.Minute,
			Time:                  30 * time.Second,
			Timeout:               20 * time.Second,
		}),
	)
	pb.RegisterPopulationServiceServer(grpcSrv, &grpcServer{})

	// Start both servers
	go func() {
		log.Printf("Starting REST server on port 8080")
		if err := restServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("REST server error: %v", err)
		}
	}()

	go func() {
		log.Printf("Starting gRPC server on port 50051")
		if err := grpcSrv.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down servers...")

	// Graceful shutdown
	grpcSrv.GracefulStop()
	if err := restServer.Shutdown(context.Background()); err != nil {
		log.Printf("Error shutting down REST server: %v", err)
	}
}
