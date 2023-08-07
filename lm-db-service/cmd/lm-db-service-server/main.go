package main

import (
	"leavemanagement/lm-db-service/internal/storage/database"
	"leavemanagement/lm-db-service/models"
	"leavemanagement/lm-db-service/pkg/pb"
	"log"
	"net"

	"leavemanagement/lm-db-service/cmd/lm-db-service-server/services"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	databaseType = "mysql"
	protocol     = "tcp"
	addr         = "0.0.0.0:50051"
)

func main() {
	var db models.DatabaseIF
	log.Print("Leave Management Server")
	lis, err := net.Listen(protocol, addr)
	if err != nil {
		log.Fatalf("failed to listen:%v", err)
	}
	db, err = database.NewMysqlDB(databaseType)
	if err != nil {
		log.Fatal(err)
	}
	s := grpc.NewServer()
	pb.RegisterLeaveManagementSerivceServer(s, &services.Server{
		DB: db,
	})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve:%v", err)
	}
}
