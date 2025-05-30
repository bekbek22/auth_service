// cmd/main.go
package main

import (
	"fmt"
	"log"
	"net"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"

	pb "github.com/bekbek22/auth_service/api/proto"
	"github.com/bekbek22/auth_service/config"
	"github.com/bekbek22/auth_service/internal/controller"
	"github.com/bekbek22/auth_service/internal/repository"
	"github.com/bekbek22/auth_service/internal/service"
)

func main() {
	//Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found, using default config")
	}

	//Load Config
	cfg := config.Load()

	//Connect MongoDB
	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	client, err := mongo.Connect(cfg.Ctx, clientOptions)
	if err != nil {
		log.Fatalf("❌ Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(cfg.Ctx)
	db := client.Database(cfg.MongoDBName)

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg)
	authController := controller.NewAuthController(authService)

	//Create gRPC Server
	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("❌ Failed to listen on port %s: %v", cfg.GRPCPort, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, authController)

	fmt.Printf("gRPC server is running on port %s\n", cfg.GRPCPort)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}
