package services

import (
	"context"
	"leavemanagement/lm-db-service/models"
	"leavemanagement/lm-db-service/pkg/pb"
)

type Server struct {
	pb.UnimplementedLeaveManagementSerivceServer
	DB models.DatabaseIF
}

func (svc Server) LeavesList(ctx context.Context, req *pb.LeavesListRequest) (*pb.LeavesListResponse, error) {
	leaves, err := svc.DB.LeavesList(ctx, req)
	return leaves, err
}

func (svc Server) GetLeaveById(ctx context.Context, req *pb.GetLeaveByIdRequest) (*pb.GetLeaveByIdResponse, error) {
	leave, err := svc.DB.GetLeaveById(ctx, req)
	return leave, err
}

func (svc Server) ApplyLeave(ctx context.Context, req *pb.ApplyLeaveRequest) (*pb.ApplyLeaveResponse, error) {
	err := svc.DB.ApplyLeave(ctx, req)
	return &pb.ApplyLeaveResponse{}, err
}

func (svc Server) ChangeLeaveStatus(ctx context.Context, req *pb.ChangeLeaveStatusRequest) (*pb.ChangeLeaveStatusResponse, error) {
	err := svc.DB.ChangeLeaveStatus(ctx, req)
	return &pb.ChangeLeaveStatusResponse{}, err
}

func (svc Server) DeleteLeave(ctx context.Context, req *pb.DeleteLeaveRequest) (*pb.DeleteLeaveResponse, error) {
	err := svc.DB.DeleteLeave(ctx, req)
	return &pb.DeleteLeaveResponse{}, err
}

func (svc Server) UpdateLeave(ctx context.Context, req *pb.UpdateLeaveRequest) (*pb.UpdateLeaveResponse, error) {
	err := svc.DB.UpdateLeave(ctx, req)
	return &pb.UpdateLeaveResponse{}, err
}
