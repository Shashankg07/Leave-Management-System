package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"leavemanagement/lm-db-service/internal/storage/validation"
	"leavemanagement/lm-db-service/models"
	"leavemanagement/lm-db-service/pkg/pb"
	"math"
	"strconv"
	"time"

	"github.com/go-playground/validator"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang/mock/mockgen/model"
)

type MysqlDB struct {
	DB *sql.DB
}

const (
	pending = string('0' + iota)
	approved
	declined
)
const (
	employeeId = string('1' + iota)
	hrId
	managerId
)
const (
	user           = "root"
	protocol       = "tcp"
	host           = "localhost"
	port           = "3306"
	database       = "leave_management"
	dateTimeFormat = "2006-01-02 15:04:05 -0700 MST"
	dateFormat     = "2006-01-02"
)

func NewMysqlDB(db string) (*MysqlDB, error) {
	mysql := &MysqlDB{}
	if err := mysql.Connect(db, port); err != nil {
		return nil, err
	}
	return mysql, nil
}
func (mysql *MysqlDB) Connect(mySql, port string) error {
	addr := fmt.Sprintf("%v@%v(%v:%v)/%v?charset=utf8&parseTime=True&loc=Local", user, protocol, host, port, database)
	db, err := sql.Open(mySql, addr)
	if err != nil {
		return errors.New("could not create connection to mysql DB")
	}

	mysql.DB = db

	if err := mysql.Test(); err != nil {
		return err
	}
	return nil
}
func (mysql *MysqlDB) Test() error {
	if err := mysql.DB.Ping(); err != nil {
		return errors.New("ping to mysql DB failed")
	}
	return nil
}
func (d MysqlDB) getDesignationId(employeeId string) (string, error) {
	var designationId string
	getDesignationIdQuery := `SELECT designation_id FROM lm_employee where employee_id=?`
	err := d.DB.QueryRow(getDesignationIdQuery, employeeId).Scan(&designationId)
	if err != nil {
		return "", err
	}
	return designationId, nil
}
func (d MysqlDB) getTotalLeavesTaken(employeeId, leaveTypeId string) (int, error) {
	var totalLeavesTaken int
	totalLeavesTakenQuery := `
						SELECT 
							IFNULL(SUM(no_of_days),0) 
						FROM lm_leave_application 
						WHERE 
							employee_id =? 
							AND leave_type_id=?`
	err := d.DB.QueryRow(totalLeavesTakenQuery, employeeId, leaveTypeId).Scan(&totalLeavesTaken)
	if err != nil {
		return 0, err
	}
	return totalLeavesTaken, nil
}
func (d MysqlDB) getAllowedDays(leaveTypeId string) (int, error) {
	var noOfDaysAllowed int
	allowedDaysQuery := `SELECT number_of_days_allowed FROM lm_leave_type WHERE leave_type_id=?`
	err := d.DB.QueryRow(allowedDaysQuery, leaveTypeId).Scan(&noOfDaysAllowed)
	if err != nil {
		return 0, err
	}
	return noOfDaysAllowed, err
}

func (d MysqlDB) ApplyLeave(ctx context.Context, req *pb.ApplyLeaveRequest) error {
	var leaveBalance int
	leaveTypeId, _ := strconv.ParseInt(req.LeaveTypeId, 10, 32)
	validate := validator.New()
	fields := models.ValidateApplyLeave{
		EmployeeId:  req.EmployeeId,
		LeaveTypeId: int(leaveTypeId),
		FromDate:    req.FromDate,
		ToDate:      req.ToDate,
		Comment:     req.Comment,
	}
	err := validate.Struct(fields)
	if err != nil {
		return errors.New("invalid input")
	}
	err = validation.ValidateFromDate(req.FromDate)
	if err != nil {
		return err
	}
	err = validation.ValidateToDate(req.ToDate)
	if err != nil {
		return err
	}

	applyLeaveQuery := `
					INSERT INTO lm_leave_application (
						employee_id, 
						leave_type_id, 
						date_of_application, 
						from_date, 
						to_date,
						no_of_days,
						leave_balance,
						comment) 
					VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	startDate, _ := time.Parse(dateFormat, fields.FromDate)
	endDate, _ := time.Parse(dateFormat, fields.ToDate)
	noOfDays := int(math.Ceil(endDate.Sub(startDate).Hours()/24)) + 1

	noOfDaysAllowed, err := d.getAllowedDays(req.LeaveTypeId)
	if err != nil {
		return err
	}

	totalLeavesTaken, err := d.getTotalLeavesTaken(req.EmployeeId, req.LeaveTypeId)
	if err != nil {
		return err
	}

	if noOfDaysAllowed < totalLeavesTaken+noOfDays {
		return errors.New("leaves not remaining")
	} else {
		leaveBalance = noOfDaysAllowed - (totalLeavesTaken + noOfDays)
	}

	dateOfApplication := time.Now().Format(dateTimeFormat)
	_, err = d.DB.Exec(applyLeaveQuery, fields.EmployeeId, fields.LeaveTypeId, dateOfApplication, fields.FromDate, fields.ToDate, noOfDays, leaveBalance, fields.Comment)
	if err != nil {
		return err
	}
	return nil
}
func (d MysqlDB) LeavesList(ctx context.Context, req *pb.LeavesListRequest) (*pb.LeavesListResponse, error) {
	validate := validator.New()
	leaveStatus, _ := strconv.ParseInt(req.LeaveStatus, 10, 32)
	fields := models.ValidateLeavesList{
		EmployeeId:  req.EmployeeId,
		LeaveStatus: int(leaveStatus),
	}
	err := validate.Struct(fields)
	if err != nil {
		return &pb.LeavesListResponse{}, err
	}

	leaves := &pb.LeavesListResponse{}
	getAllLeaveQuery := `
					SELECT 
						first_name,
						last_name,
						application_id,
						employee_id, 
						leave_type_id, 
						date_of_application,
						from_date, 
						to_date,
						no_of_days,
						leave_balance,
						leave_status,
						comment, 
						IFNULL(date_of_approval,"N/A") 
					FROM lm_leave_application 
					INNER JOIN lm_employee USING (employee_id)`

	designationId, err := d.getDesignationId(req.EmployeeId)
	if err != nil {
		return &pb.LeavesListResponse{}, err
	}
	if designationId != hrId && designationId != managerId {
		return &pb.LeavesListResponse{}, errors.New("access denied")
	} else {
		if req.LeaveStatus == pending || req.LeaveStatus == approved || req.LeaveStatus == declined {
			getAllLeaveQuery = fmt.Sprintf("%v WHERE leave_status=%s", getAllLeaveQuery, req.LeaveStatus)
		}
		rows, err := d.DB.Query(getAllLeaveQuery)
		if err != nil {
			return &pb.LeavesListResponse{}, err
		}
		for rows.Next() {
			leave := pb.GetLeaveByIdResponse{}
			err = rows.Scan(
				&leave.FirstName,
				&leave.LastName,
				&leave.ApplicationId,
				&leave.EmployeeId,
				&leave.LeaveTypeId,
				&leave.DateOfApplication,
				&leave.FromDate,
				&leave.ToDate,
				&leave.NoOfDays,
				&leave.LeaveBalance,
				&leave.LeaveStatus,
				&leave.Comment,
				&leave.DateOfApproval)
			if err != nil {
				return &pb.LeavesListResponse{}, err
			}
			leaves.LeavesListResponse = append(leaves.LeavesListResponse, &leave)
		}
	}
	return leaves, nil
}
func (d MysqlDB) GetLeaveById(ctx context.Context, req *pb.GetLeaveByIdRequest) (*pb.GetLeaveByIdResponse, error) {
	var leave *pb.GetLeaveByIdResponse = &pb.GetLeaveByIdResponse{}
	getGetLeaveByIdQuery := `
					SELECT 
						first_name,
						last_name,
						application_id,
						employee_id, 
						leave_type_id, 
						date_of_application,
						from_date,
						to_date,
						no_of_days,
						leave_balance,
						leave_status,
						comment,
						IFNULL(date_of_approval,"N/A") 
					FROM lm_leave_application 
					INNER JOIN lm_employee 
					USING (employee_id) 
					WHERE application_id=?`
	row, err := d.DB.Query(getGetLeaveByIdQuery, req.ApplicationId)
	if err != nil {
		return &pb.GetLeaveByIdResponse{}, err
	}
	for row.Next() {
		err = row.Scan(
			&leave.FirstName,
			&leave.LastName,
			&leave.ApplicationId,
			&leave.EmployeeId,
			&leave.LeaveTypeId,
			&leave.DateOfApplication,
			&leave.FromDate,
			&leave.ToDate,
			&leave.NoOfDays,
			&leave.LeaveBalance,
			&leave.LeaveStatus,
			&leave.Comment,
			&leave.DateOfApproval)
		if err != nil {
			return &pb.GetLeaveByIdResponse{}, err
		}
	}
	return leave, nil
}
func (d MysqlDB) ChangeLeaveStatus(ctx context.Context, req *pb.ChangeLeaveStatusRequest) error {
	leaveStatus, _ := strconv.ParseInt(req.LeaveStatus, 10, 32)
	validate := validator.New()
	fields := models.ValidateChangeLeaveStatus{
		EmployeeId:    req.EmployeeId,
		ApplicationId: req.ApplicationId,
		LeaveStatus:   int(leaveStatus),
	}
	err := validate.Struct(fields)
	if err != nil {
		return errors.New("invalid input")
	}

	designationId, err := d.getDesignationId(req.EmployeeId)
	if err != nil {
		return err
	}

	if designationId != managerId {
		return errors.New("access denied")
	} else {
		changeLeaveStatusQuery := `
							UPDATE lm_leave_application 
							SET 
								leave_status=?, 
								date_of_approval=? 
							WHERE lm_leave_application.application_id=?`
		dateOfApproval := time.Now().Format(dateTimeFormat)
		_, err = d.DB.Exec(changeLeaveStatusQuery, req.LeaveStatus, dateOfApproval, req.ApplicationId)
		if err != nil {
			return err
		}
	}
	return nil
}
func (d MysqlDB) DeleteLeave(ctx context.Context, req *pb.DeleteLeaveRequest) error {
	validate := validator.New()
	fields := models.ValidateDeleteLeave{
		EmployeeId:    req.EmployeeId,
		ApplicationId: req.ApplicationId,
	}
	err := validate.Struct(fields)
	if err != nil {
		return errors.New("invalid input")
	}

	designationId, err := d.getDesignationId(req.EmployeeId)
	if err != nil {
		return err
	}

	if designationId != hrId {
		return errors.New("access denied")
	} else {
		deleteLeaveQuery := `DELETE FROM lm_leave_application WHERE lm_leave_application.application_id=?`
		_, err = d.DB.Exec(deleteLeaveQuery, req.ApplicationId)
		if err != nil {
			return err
		}
	}
	return nil
}
func (d MysqlDB) UpdateLeave(ctx context.Context, req *pb.UpdateLeaveRequest) error {
	var employeeId string
	leaveTypeId, _ := strconv.ParseInt(req.LeaveTypeId, 10, 32)

	validate := validator.New()
	fields := models.ValidateUpdateLeave{
		ApplicationId: req.ApplicationId,
		EmployeeId:    req.EmployeeId,
		LeaveTypeId:   int(leaveTypeId),
		FromDate:      req.FromDate,
		ToDate:        req.ToDate,
		Comment:       req.Comment,
	}
	err := validate.Struct(fields)
	if err != nil {
		return errors.New("invalid input")
	}
	err = validation.ValidateFromDate(req.FromDate)
	if err != nil {
		return err
	}
	err = validation.ValidateToDate(req.ToDate)
	if err != nil {
		return err
	}

	getEmployeeIdQuery := `SELECT employee_id FROM lm_leave_application where application_id=?`
	err = d.DB.QueryRow(getEmployeeIdQuery, req.ApplicationId).Scan(&employeeId)
	if err != nil {
		return err
	}

	if employeeId != req.EmployeeId {
		return errors.New("access denied")
	} else {
		updateLeaveQuery := `UPDATE lm_leave_application SET 
			leave_type_id=?, 
			comment=?, 
			from_date=?, 
			to_date=? 
			WHERE lm_leave_application.application_id=?`
		_, err := d.DB.Exec(updateLeaveQuery, req.LeaveTypeId, req.Comment, req.FromDate, req.ToDate, req.ApplicationId)
		if err != nil {
			return err
		}
	}
	return nil
}
