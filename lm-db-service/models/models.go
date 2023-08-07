package models

import (
	"context"
	"leavemanagement/lm-db-service/pkg/pb"
)

type DatabaseIF interface {
	Connect(string, string) error
	Test() error
	ApplyLeave(context.Context, *pb.ApplyLeaveRequest) error
	ChangeLeaveStatus(context.Context, *pb.ChangeLeaveStatusRequest) error
	GetLeaveById(context.Context, *pb.GetLeaveByIdRequest) (*pb.GetLeaveByIdResponse, error)
	DeleteLeave(context.Context, *pb.DeleteLeaveRequest) error
	UpdateLeave(context.Context, *pb.UpdateLeaveRequest) error
	LeavesList(context.Context, *pb.LeavesListRequest) (*pb.LeavesListResponse, error)
}

type ValidateApplyLeave struct {
	EmployeeId  string `validate:"required"`
	LeaveTypeId int    `validate:"required,gte=1,lte=6"`
	FromDate    string `validate:"required"`
	ToDate      string `validate:"required"`
	Comment     string `validate:"required"`
}
type ValidateLeavesList struct {
	EmployeeId  string `validate:"required"`
	LeaveStatus int    `validate:"required,gte=0,lte=2"`
}
type ValidateChangeLeaveStatus struct {
	EmployeeId    string `validate:"required"`
	ApplicationId string `validate:"required"`
	LeaveStatus   int    `validate:"required,gte=0,lte=2"`
}
type ValidateDeleteLeave struct {
	EmployeeId    string `validate:"required"`
	ApplicationId string `validate:"required"`
}
type ValidateUpdateLeave struct {
	ApplicationId string `validate:"required"`
	EmployeeId    string `validate:"required"`
	LeaveTypeId   int    `validate:"required,gte=1,lte=6"`
	FromDate      string `validate:"required"`
	ToDate        string `validate:"required"`
	Comment       string `validate:"required"`
}
