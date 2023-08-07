package database

import (
	"context"
	"errors"
	"leavemanagement/lm-db-service/pkg/pb"
	"reflect"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql"
)

func getTestMysqlDB(t *testing.T) (*MysqlDB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal("failed to create mock DB")
	}
	testDB := &MysqlDB{
		DB: db,
	}
	return testDB, mock
}
func TestMysqlMock_NewMysqlDB(t *testing.T) {
	tests := []struct {
		description  string
		databaseType string
		isError      bool
	}{
		{
			description:  "success",
			databaseType: "mysql",
			isError:      false,
		},
		{
			description:  "error",
			databaseType: "",
			isError:      true,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			if test.isError == false {
				_, err := NewMysqlDB(test.databaseType)
				if err != nil {
					t.Error(err)
				}
			} else if test.isError == true {
				_, err := NewMysqlDB(test.databaseType)
				if err == nil {
					t.Error(err)
				}
			}
		})

	}
}
func TestMysqlDB_Connect(t *testing.T) {
	tests := []struct {
		description  string
		databaseType string
		port         string
		isError      bool
	}{
		{
			description:  "success",
			databaseType: "mysql",
			port:         "3306",
			isError:      false,
		},
		{
			description:  "connection error",
			databaseType: "",
			port:         "3306",
			isError:      true,
		},
		{
			description:  "ping error",
			databaseType: "mysql",
			port:         "3307",
			isError:      true,
		},
	}
	testDB, _ := getTestMysqlDB(t)
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			if test.isError == false {
				err := testDB.Connect(test.databaseType, test.port)
				if err != nil {
					t.Errorf("Connect() got error %v, %v", err, false)
				}
			} else if test.isError == true {
				err := testDB.Connect(test.databaseType, test.port)
				if err == nil {
					t.Errorf("Connect() got error %v, %v", err, true)
				}
			}
		})
	}
}
func TestMysqlMock_allowedDays(t *testing.T) {
	testDB, mock := getTestMysqlDB(t)
	tests := []struct {
		description string
		leaveTypeId string
		isError     bool
	}{
		{
			description: "success",
			leaveTypeId: "1",
			isError:     false,
		},
		{
			description: "error",
			leaveTypeId: "1",
			isError:     true,
		},
	}
	column := []string{
		"number_of_days_allowed",
	}
	row := sqlmock.NewRows(column).AddRow(
		"10",
	)
	allowedDaysQuery := `SELECT number_of_days_allowed FROM lm_leave_type WHERE leave_type_id=\?`
	mock.ExpectQuery(allowedDaysQuery).WithArgs("1").WillReturnRows(row)
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			if test.isError == false {
				expected := 10
				result, err := testDB.getAllowedDays(test.leaveTypeId)
				if err != nil {
					t.Errorf("got error %v: want error: %v", err, false)
				}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("expected %v: got %v", expected, result)
				}
			} else if test.isError == true {
				_, err := testDB.getAllowedDays(test.leaveTypeId)
				if err == nil {
					t.Errorf("got error %v: want error: %v", err, true)
				}
			}
		})
	}
}
func TestMysqlMock_totalLeavesTaken(t *testing.T) {
	testDB, mock := getTestMysqlDB(t)
	tests := []struct {
		description string
		employeeId  string
		leaveTypeId string
		isError     bool
	}{
		{
			description: "success",
			employeeId:  "1",
			leaveTypeId: "1",
			isError:     false,
		},
		{
			description: "error",
			employeeId:  "1",
			leaveTypeId: "1",
			isError:     true,
		},
	}
	column := []string{
		"no_of_days",
	}
	row := sqlmock.NewRows(column).AddRow(
		"1",
	)
	totalLeavesTakenQuery := `
								SELECT
									IFNULL\(SUM\(no_of_days\),0\)
								FROM lm_leave_application
								WHERE
									employee_id =\?
									AND leave_type_id=\?`
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			if test.isError == false {
				expected := 1
				mock.ExpectQuery(totalLeavesTakenQuery).WithArgs(test.employeeId, test.leaveTypeId).WillReturnRows(row)
				result, err := testDB.getTotalLeavesTaken(test.employeeId, test.leaveTypeId)
				if err != nil {
					t.Errorf("got error %v: want error: %v", err, false)
				}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("expected %v: got %v", expected, result)
				}
			} else if test.isError == true {
				mock.ExpectQuery(totalLeavesTakenQuery).WithArgs(test.employeeId, test.leaveTypeId).
					WillReturnError(errors.New("error"))
				_, err := testDB.getTotalLeavesTaken(test.employeeId, test.leaveTypeId)
				if err == nil {
					t.Errorf("got error %v: want error: %v", err, true)
				}
			}
		})
	}
}
func TestMySqlMock_ApplyLeave(t *testing.T) {
	testDB, mock := getTestMysqlDB(t)
	tests := []struct {
		description string
		request     *pb.ApplyLeaveRequest
		isError     string
	}{
		{
			description: "success",
			request: &pb.ApplyLeaveRequest{
				EmployeeId:  "1",
				LeaveTypeId: "1",
				FromDate:    "2022-04-20",
				ToDate:      "2022-04-21",
				Comment:     "Fever",
			},
			isError: "false",
		},
		{
			description: "execution failed",
			request: &pb.ApplyLeaveRequest{
				EmployeeId:  "1",
				LeaveTypeId: "1",
				FromDate:    "2022-04-20",
				ToDate:      "2022-04-21",
				Comment:     "Fever",
			},
			isError: "true",
		},
		{
			description: "invalid input",
			request: &pb.ApplyLeaveRequest{
				EmployeeId:  "",
				LeaveTypeId: "1",
				FromDate:    "2022-04-20",
				ToDate:      "2022-04-21",
				Comment:     "Fever",
			},
			isError: "true",
		},
		{
			description: "no leaves remaining",
			request: &pb.ApplyLeaveRequest{
				EmployeeId:  "1",
				LeaveTypeId: "1",
				FromDate:    "2022-04-20",
				ToDate:      "2022-04-26",
				Comment:     "Fever",
			},
			isError: "true",
		},
		{
			description: "Allowed Days",
			request: &pb.ApplyLeaveRequest{
				EmployeeId:  "1",
				LeaveTypeId: "1",
				FromDate:    "2022-04-20",
				ToDate:      "2022-04-21",
				Comment:     "Fever",
			},
			isError: "forAllowedDays",
		},
		{
			description: "invalid toDate",
			request: &pb.ApplyLeaveRequest{
				EmployeeId:  "1",
				LeaveTypeId: "1",
				FromDate:    "202-04-20",
				ToDate:      "2022-04-21",
				Comment:     "Fever",
			},
			isError: "true",
		},
		{
			description: "invalid from Date",
			request: &pb.ApplyLeaveRequest{
				EmployeeId:  "1",
				LeaveTypeId: "1",
				FromDate:    "2022-04-20",
				ToDate:      "202-04-21",
				Comment:     "Fever",
			},
			isError: "true",
		},
	}
	applyLeaveQuery := `
					INSERT INTO lm_leave_application \(
						employee_id, 
						leave_type_id, 
						date_of_application, 
						from_date, 
						to_date,
						no_of_days,
						leave_balance,
						comment\) 
					VALUES \(\?, \?, \?, \?, \?, \?, \?, \?\)`
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			if test.isError == "false" {
				column := []string{
					"number_of_days_allowed",
				}
				row := sqlmock.NewRows(column).AddRow(
					"3",
				)
				allowedDaysQuery := `SELECT number_of_days_allowed FROM lm_leave_type WHERE leave_type_id=\?`
				mock.ExpectQuery(allowedDaysQuery).WithArgs("1").WillReturnRows(row)
				column = []string{
					"no_of_days",
				}
				row = sqlmock.NewRows(column).AddRow(
					"1",
				)
				totalLeavesTakenQuery := `
									SELECT 
										IFNULL\(SUM\(no_of_days\),0\) 
									FROM lm_leave_application 
									WHERE 
										employee_id =\? 
										AND leave_type_id=\?`
				mock.ExpectQuery(totalLeavesTakenQuery).WithArgs("1", "1").WillReturnRows(row)
				mock.ExpectExec(applyLeaveQuery).
					WithArgs("1", 1, time.Now().Format(dateTimeFormat), "2022-04-20", "2022-04-21", 2, 0, "Fever").
					WillReturnResult(sqlmock.NewResult(1, 1))
				got := testDB.ApplyLeave(context.Background(), test.request)
				if got != nil {
					t.Errorf("got error %v: want error: %v", got, false)
				}
			} else if test.isError == "true" {
				column := []string{
					"number_of_days_allowed",
				}
				row := sqlmock.NewRows(column).AddRow(
					"3",
				)
				allowedDaysQuery := `SELECT number_of_days_allowed FROM lm_leave_type WHERE leave_type_id=\?`
				mock.ExpectQuery(allowedDaysQuery).WithArgs("1").WillReturnRows(row)
				column = []string{
					"no_of_days",
				}
				row = sqlmock.NewRows(column).AddRow(
					"1",
				)
				totalLeavesTakenQuery := `
									SELECT 
										IFNULL\(SUM\(no_of_days\),0\) 
									FROM lm_leave_application 
									WHERE 
										employee_id =\? 
										AND leave_type_id=\?`
				mock.ExpectQuery(totalLeavesTakenQuery).WithArgs("1", "1").WillReturnRows(row)
				mock.ExpectExec(applyLeaveQuery).
					WithArgs("1", 1, time.Now().Format(dateTimeFormat), "2022-04-20", "2022-04-21", 2, 0, "Fever").
					WillReturnError(errors.New("error"))
				got := testDB.ApplyLeave(context.Background(), test.request)
				if got == nil {
					t.Errorf("got error %v: want error: %v", got, true)
				}
			} else if test.isError == "forAllowedDays" {
				column := []string{
					"number_of_days_allowed",
				}
				row := sqlmock.NewRows(column).AddRow(
					"3",
				)
				allowedDaysQuery := `SELECT number_of_days_allowed FROM lm_leave_type WHERE leave_type_id=\?`
				mock.ExpectQuery(allowedDaysQuery).WithArgs("1").WillReturnError(errors.New("error"))
				column = []string{
					"no_of_days",
				}
				row = sqlmock.NewRows(column).AddRow(
					"1",
				)
				totalLeavesTakenQuery := `
									SELECT 
										IFNULL\(SUM\(no_of_days\),0\) 
									FROM lm_leave_application 
									WHERE 
										employee_id =\? 
										AND leave_type_id=\?`
				mock.ExpectQuery(totalLeavesTakenQuery).WithArgs("1", "1").WillReturnRows(row)
				mock.ExpectExec(applyLeaveQuery).
					WithArgs("1", "1", time.Now().Format(dateTimeFormat), "2022-04-20", "2022-04-21", 2, 0, "Fever").
					WillReturnResult(sqlmock.NewErrorResult(errors.New("error")))
				got := testDB.ApplyLeave(context.Background(), test.request)
				if got == nil {
					t.Errorf("got error %v: want error: %v", got, true)
				}
			}
		})
	}
}
func TestMySqlMock_designation(t *testing.T) {
	testDB, mock := getTestMysqlDB(t)
	tests := []struct {
		description string
		request     string
		response    string
		isError     bool
	}{
		{
			description: "correct case",
			request:     "4",
			response:    "2",
			isError:     false,
		},
		{
			description: "wrong case",
			request:     "4",
			response:    "",
			isError:     true,
		},
	}
	expectedSql := `SELECT designation_id FROM lm_employee where employee_id=\?`
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			if test.isError == false {
				columns := []string{
					"designation_id",
				}
				rows := sqlmock.NewRows(columns).AddRow(
					"2",
				)
				mock.ExpectQuery(expectedSql).WithArgs("4").WillReturnRows(rows)
				actual, err := testDB.getDesignationId(test.request)
				if err != nil {
					t.Errorf("got error %v: want error: %v", err, false)
				}
				if !reflect.DeepEqual(test.response, actual) {
					t.Errorf("expected %v: got %v", test.response, actual)
				}
			} else if test.isError == true {
				_, err := testDB.getDesignationId(test.request)
				if err == nil {
					t.Errorf("got error %v: want error: %v", err, true)
				}
			}
		})
	}
}
func TestMySqlMock_GetLeaveApplicationById(t *testing.T) {
	testDB, mock := getTestMysqlDB(t)
	tests := []struct {
		description string
		request     *pb.GetLeaveByIdRequest
		expected    *pb.GetLeaveByIdResponse
		isError     bool
	}{
		{
			description: "common case",
			request: &pb.GetLeaveByIdRequest{
				ApplicationId: "1",
			},
			expected: &pb.GetLeaveByIdResponse{
				ApplicationId:     "1",
				EmployeeId:        "5",
				LeaveTypeId:       "3",
				DateOfApplication: "2022-04-07T23:19:53+05:30",
				FromDate:          "2022-04-11T00:00:00+05:30",
				ToDate:            "2022-04-14T00:00:00+05:30",
				NoOfDays:          "4",
				LeaveStatus:       "2",
				LeaveBalance:      "5",
				Comment:           "Exams",
				FirstName:         "Saurabh",
				LastName:          "Jain",
				DateOfApproval:    "2022-04-11T00:00:00+05:30",
			},
			isError: false,
		},
		{
			description: "common case",
			request: &pb.GetLeaveByIdRequest{
				ApplicationId: "1",
			},
			expected: nil,
			isError:  true,
		},
	}
	columns := []string{
		"first_name",
		"last_name",
		"application_id",
		"employee_id",
		"leave_type_id",
		"date_of_application",
		"from_date",
		"to_date",
		"no_of_days",
		"leave_balance",
		"leave_status",
		"comment",
		"date_of_approval",
	}
	rows := sqlmock.NewRows(columns).AddRow(
		"Saurabh",
		"Jain",
		"1",
		"5",
		"3",
		"2022-04-07T23:19:53+05:30",
		"2022-04-11T00:00:00+05:30",
		"2022-04-14T00:00:00+05:30",
		"4",
		"5",
		"2",
		"Exams",
		"2022-04-11T00:00:00+05:30",
	)
	expectedSql := `
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
					IFNULL\(date_of_approval,"N/A"\) 
				FROM lm_leave_application 
				INNER JOIN lm_employee 
				USING \(employee_id\) 
				WHERE application_id=\?`
	mock.ExpectQuery(expectedSql).WithArgs("1").WillReturnRows(rows)
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			if test.isError == true {
				mock.ExpectQuery(expectedSql).WithArgs("1").WillReturnError(errors.New("error"))
				_, err := testDB.GetLeaveById(context.Background(), test.request)
				if err == nil {
					t.Errorf("got error %v: want error: %v", err, true)
				}
			} else if test.isError == false {
				actual, err := testDB.GetLeaveById(context.Background(), test.request)
				if err != nil {
					t.Errorf("got error %v: want error: %v", err, false)
				}
				if !reflect.DeepEqual(test.expected, actual) {
					t.Errorf("expected %v: got %v", test.expected, actual)
				}
			}
		})
	}
}
func TestMySqlMock_LeavesList(t *testing.T) {
	testDB, mock := getTestMysqlDB(t)
	tests := []struct {
		description string
		request     *pb.LeavesListRequest
		expected    *pb.LeavesListResponse
		isError     string
	}{
		{
			description: "common case",
			request: &pb.LeavesListRequest{
				EmployeeId:  "7",
				LeaveStatus: "2",
			},
			expected: &pb.LeavesListResponse{
				LeavesListResponse: []*pb.GetLeaveByIdResponse{
					{
						ApplicationId:     "1",
						EmployeeId:        "5",
						LeaveTypeId:       "3",
						DateOfApplication: "2022-04-07T23:19:53+05:30",
						FromDate:          "2022-04-11T00:00:00+05:30",
						ToDate:            "2022-04-14T00:00:00+05:30",
						NoOfDays:          "4",
						LeaveBalance:      "5",
						LeaveStatus:       "2",
						Comment:           "Exams",
						FirstName:         "Saurabh",
						LastName:          "Jain",
						DateOfApproval:    "2022-04-11T00:00:00+05:30",
					},
				},
			},
			isError: "false",
		},
		{
			description: "execution failed",
			request: &pb.LeavesListRequest{
				EmployeeId:  "4",
				LeaveStatus: "2",
			},
			expected: nil,
			isError:  "true",
		},
		{
			description: "access denied",
			request: &pb.LeavesListRequest{
				EmployeeId:  "7",
				LeaveStatus: "2",
			},
			expected: nil,
			isError:  "wrong designation",
		},
		{
			description: "execution failed",
			request: &pb.LeavesListRequest{
				EmployeeId:  "",
				LeaveStatus: "2",
			},
			expected: nil,
			isError:  "true",
		},
		{
			description: "invalid leave status",
			request: &pb.LeavesListRequest{
				EmployeeId:  "7",
				LeaveStatus: "3",
			},
			expected: nil,
			isError:  "true",
		},
	}
	columns := []string{
		"first_name",
		"last_name",
		"application_id",
		"employee_id",
		"leave_type_id",
		"date_of_application",
		"from_date",
		"to_date",
		"no_of_days",
		"leave_balance",
		"leave_status",
		"comment",
		"date_of_approval",
	}
	rows := sqlmock.NewRows(columns).AddRow(
		"Saurabh",
		"Jain",
		"1",
		"5",
		"3",
		"2022-04-07T23:19:53+05:30",
		"2022-04-11T00:00:00+05:30",
		"2022-04-14T00:00:00+05:30",
		"4",
		"5",
		"2",
		"Exams",
		"2022-04-11T00:00:00+05:30",
	)
	expectedSql := `
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
					IFNULL\(date_of_approval,"N/A"\)
				FROM lm_leave_application 
				INNER JOIN lm_employee 
				USING \(employee_id\)`
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			if test.isError == "false" {
				column := []string{
					"designation_id",
				}
				row := sqlmock.NewRows(column).AddRow(
					"2",
				)
				mock.ExpectQuery(`SELECT designation_id FROM lm_employee where employee_id=\?`).WithArgs("7").
					WillReturnRows(row)
				mock.ExpectQuery(expectedSql).WillReturnRows(rows)
				actual, err := testDB.LeavesList(context.Background(), test.request)
				if err != nil {
					t.Errorf("got error %v: want error: %v", err, false)
				}
				if !reflect.DeepEqual(test.expected, actual) {
					t.Errorf("expected %v: got %v", test.expected, actual)
				}
			} else if test.isError == "true" {
				column := []string{
					"designation_id",
				}
				row := sqlmock.NewRows(column).AddRow(
					"2",
				)
				mock.ExpectQuery(`SELECT designation_id FROM lm_employee where employee_id=\?`).WithArgs("4").
					WillReturnRows(row)
				mock.ExpectQuery(expectedSql).WillReturnError(errors.New("error"))
				_, err := testDB.LeavesList(context.Background(), test.request)
				if err == nil {
					t.Errorf("got error %v: want error: %v", err, true)
				}
			} else if test.isError == "wrong designation" {
				column := []string{
					"designation_id",
				}
				row := sqlmock.NewRows(column).AddRow(
					"1",
				)
				mock.ExpectQuery(`SELECT designation_id FROM lm_employee where employee_id=\?`).WithArgs("7").
					WillReturnRows(row)
				mock.ExpectQuery(expectedSql).WillReturnError(errors.New("error"))
				_, err := testDB.LeavesList(context.Background(), test.request)
				if err == nil {
					t.Errorf("got error %v: want error: %v", err, true)
				}
			}
		})
	}
}
func TestMySqlMock_DeleteLeave(t *testing.T) {
	testDB, mock := getTestMysqlDB(t)
	tests := []struct {
		description string
		request     *pb.DeleteLeaveRequest
		isError     string
	}{
		{
			description: "success",
			request: &pb.DeleteLeaveRequest{
				EmployeeId:    "7",
				ApplicationId: "2",
			},
			isError: "false",
		},
		{
			description: "Access denied",
			request: &pb.DeleteLeaveRequest{
				EmployeeId:    "1",
				ApplicationId: "2",
			},
			isError: "accessDenied",
		},
		{
			description: "validation error",
			request: &pb.DeleteLeaveRequest{
				EmployeeId:    "",
				ApplicationId: "2",
			},
			isError: "true",
		},
		{
			description: "common case",
			request: &pb.DeleteLeaveRequest{
				EmployeeId:    "7",
				ApplicationId: "2",
			},
			isError: "true",
		},
	}
	deleteQuery := `DELETE FROM lm_leave_application WHERE lm_leave_application.application_id=\?`
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			if test.isError == "false" {
				columns := []string{
					"designation_id",
				}
				rows := sqlmock.NewRows(columns).AddRow(
					"2",
				)
				designationIdQuery := `SELECT designation_id FROM lm_employee where employee_id=\?`
				mock.ExpectQuery(designationIdQuery).WithArgs("7").WillReturnRows(rows)
				mock.ExpectExec(deleteQuery).WithArgs("2").WillReturnResult(sqlmock.NewResult(1, 1))
				err := testDB.DeleteLeave(context.Background(), test.request)
				if err != nil {
					t.Errorf("got error %v: want error: %v", err, false)
				}
			} else if test.isError == "true" {
				columns := []string{
					"designation_id",
				}
				rows := sqlmock.NewRows(columns).AddRow(
					"2",
				)
				designationIdQuery := `SELECT designation_id FROM lm_employee where employee_id=\?`
				mock.ExpectQuery(designationIdQuery).WithArgs("1").WillReturnRows(rows)
				mock.ExpectExec(deleteQuery).WithArgs("2").WillReturnError(errors.New("error"))
				err := testDB.DeleteLeave(context.Background(), test.request)
				if err == nil {
					t.Errorf("got error %v: want error: %v", err, true)
				}
			} else if test.isError == "accessDenied" {
				columns := []string{
					"designation_id",
				}
				rows := sqlmock.NewRows(columns).AddRow(
					"1",
				)
				designationIdQuery := `SELECT designation_id FROM lm_employee where employee_id=\?`
				mock.ExpectQuery(designationIdQuery).WithArgs("1").WillReturnRows(rows)
				mock.ExpectExec(deleteQuery).WithArgs("2").WillReturnResult(sqlmock.NewResult(1, 1))
				err := testDB.DeleteLeave(context.Background(), test.request)
				if err == nil {
					t.Errorf("got error %v: want error: %v", err, true)
				}
			}
		})
	}
}
func TestMySqlMock_ChangeLeaveStatus(t *testing.T) {
	testDB, mock := getTestMysqlDB(t)
	tests := []struct {
		description string
		request     *pb.ChangeLeaveStatusRequest
		isError     string
	}{
		{
			description: "success",
			request: &pb.ChangeLeaveStatusRequest{
				EmployeeId:    "8",
				ApplicationId: "2",
				LeaveStatus:   "2",
			},
			isError: "false",
		},
		{
			description: "access denied",
			request: &pb.ChangeLeaveStatusRequest{
				EmployeeId:    "8",
				ApplicationId: "2",
				LeaveStatus:   "2",
			},
			isError: "accessDenied",
		},
		{
			description: "validation error",
			request: &pb.ChangeLeaveStatusRequest{
				EmployeeId:    "",
				ApplicationId: "2",
				LeaveStatus:   "2",
			},
			isError: "true",
		},
		{
			description: "common error",
			request: &pb.ChangeLeaveStatusRequest{
				EmployeeId:    "8",
				ApplicationId: "2",
				LeaveStatus:   "2",
			},
			isError: "true",
		},
	}
	updateQuery := `
				UPDATE lm_leave_application 
				SET 
					leave_status=\?, 
					date_of_approval=\? 
				WHERE lm_leave_application.application_id=\?`
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			if test.isError == "false" {
				columns := []string{
					"designation_id",
				}
				rows := sqlmock.NewRows(columns).AddRow(
					"3",
				)
				designationIdQuery := `SELECT designation_id FROM lm_employee where employee_id=\?`
				mock.ExpectQuery(designationIdQuery).WithArgs("8").WillReturnRows(rows)
				mock.ExpectExec(updateQuery).WithArgs("2", time.Now().Format(dateTimeFormat), "2").
					WillReturnResult(sqlmock.NewResult(1, 1))
				err := testDB.ChangeLeaveStatus(context.Background(), test.request)
				if err != nil {
					t.Errorf("got error %v: want error: %v", err, false)
				}
			}
			if test.isError == "true" {
				columns := []string{
					"designation_id",
				}
				rows := sqlmock.NewRows(columns).AddRow(
					"3",
				)
				designationIdQuery := `SELECT designation_id FROM lm_employee where employee_id=\?`
				mock.ExpectQuery(designationIdQuery).WithArgs("8").WillReturnRows(rows)
				mock.ExpectExec(updateQuery).WithArgs("2", time.Now().Format(dateTimeFormat)).
					WillReturnError(errors.New("error"))
				err := testDB.ChangeLeaveStatus(context.Background(), test.request)
				if err == nil {
					t.Errorf("got error %v: want error: %v", err, true)
				}
			}
			if test.isError == "accessDenied" {
				columns := []string{
					"designation_id",
				}
				rows := sqlmock.NewRows(columns).AddRow(
					"2",
				)
				designationIdQuery := `SELECT designation_id FROM lm_employee where employee_id=\?`
				mock.ExpectQuery(designationIdQuery).WithArgs("8").WillReturnRows(rows)
				mock.ExpectExec(updateQuery).WithArgs("2", time.Now().Format(dateTimeFormat)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				err := testDB.ChangeLeaveStatus(context.Background(), test.request)
				if err == nil {
					t.Errorf("got error %v: want error: %v", err, true)
				}
			}
		})
	}
}
func TestMySqlMock_UpdateLeave(t *testing.T) {
	testDB, mock := getTestMysqlDB(t)
	tests := []struct {
		description string
		request     *pb.UpdateLeaveRequest
		isError     bool
	}{
		{
			description: "success",
			request: &pb.UpdateLeaveRequest{
				ApplicationId: "1",
				EmployeeId:    "2",
				LeaveTypeId:   "3",
				Comment:       "fever",
				FromDate:      "2022-04-24",
				ToDate:        "2022-04-25",
			},
			isError: false,
		},
		{
			description: "common case",
			request: &pb.UpdateLeaveRequest{
				ApplicationId: "1",
				EmployeeId:    "2",
				LeaveTypeId:   "3",
				Comment:       "fever",
				FromDate:      "2022-04-24",
				ToDate:        "2022-04-25",
			},
			isError: true,
		},
		{
			description: "validation error",
			request: &pb.UpdateLeaveRequest{
				ApplicationId: "",
				EmployeeId:    "2",
				LeaveTypeId:   "3",
				Comment:       "fever",
				FromDate:      "2022-04-24",
				ToDate:        "2022-04-25",
			},
			isError: true,
		},
		{
			description: "invalid employeeId",
			request: &pb.UpdateLeaveRequest{
				ApplicationId: "1",
				EmployeeId:    "d",
				LeaveTypeId:   "3",
				Comment:       "fever",
				FromDate:      "2022-04-24",
				ToDate:        "2022-04-25",
			},
			isError: true,
		},
		{
			description: "access denied",
			request: &pb.UpdateLeaveRequest{
				ApplicationId: "2",
				EmployeeId:    "4",
				LeaveTypeId:   "3",
				Comment:       "fever",
				FromDate:      "2022-04-24",
				ToDate:        "2022-04-25",
			},
			isError: true,
		},
		{
			description: "invalid fromDate",
			request: &pb.UpdateLeaveRequest{
				ApplicationId: "2",
				EmployeeId:    "4",
				LeaveTypeId:   "3",
				Comment:       "fever",
				FromDate:      "202-04-24",
				ToDate:        "2022-04-25",
			},
			isError: true,
		},
		{
			description: "invalid toDate",
			request: &pb.UpdateLeaveRequest{
				ApplicationId: "2",
				EmployeeId:    "4",
				LeaveTypeId:   "3",
				Comment:       "fever",
				FromDate:      "2022-04-24",
				ToDate:        "202-04-25",
			},
			isError: true,
		},
	}
	updateQuery := `
			UPDATE lm_leave_application SET 
				leave_type_id=\?, 
				comment=\?, 
				from_date=\?, 
				to_date=\? 
			WHERE lm_leave_application.application_id=\?`
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			if test.isError == false {
				columns := []string{
					"employee_id",
				}
				rows := sqlmock.NewRows(columns).AddRow(
					"2",
				)
				getEmployeeIdQuery := `SELECT employee_id FROM lm_leave_application where application_id=\?`
				mock.ExpectQuery(getEmployeeIdQuery).WithArgs("1").WillReturnRows(rows)
				mock.ExpectExec(updateQuery).WithArgs("3", "fever", "2022-04-24", "2022-04-25", "1").
					WillReturnResult(sqlmock.NewResult(1, 1))
				err := testDB.UpdateLeave(context.Background(), test.request)
				if err != nil {
					t.Errorf("got error %v: want error: %v", err, false)
				}
			} else if test.isError == true {
				columns := []string{
					"employee_id",
				}
				rows := sqlmock.NewRows(columns).AddRow(
					"2",
				)
				getEmployeeIdQuery := `SELECT employee_id FROM lm_leave_application where application_id=\?`
				mock.ExpectQuery(getEmployeeIdQuery).WithArgs("1").WillReturnRows(rows)
				mock.ExpectExec(updateQuery).WithArgs("3", "fever", "2022-04-24", "2022-04-25", "1").
					WillReturnError(errors.New("error"))
				err := testDB.UpdateLeave(context.Background(), test.request)
				if err == nil {
					t.Errorf("got error %v: want error: %v", err, true)
				}
			}
		})
	}
}
