package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UMRole represents roles from um-api JWT claims
type UMRole string

const (
	UMRoleSuper UMRole = "SUPER"
	UMRoleAdmin UMRole = "ADMIN"
	UMRoleUser  UMRole = "USER"
)

// EmployeeRole represents local branch roles
type EmployeeRole string

const (
	EmployeeRoleAdmin   EmployeeRole = "ADMIN"
	EmployeeRoleManager EmployeeRole = "MANAGER"
	EmployeeRoleStaff   EmployeeRole = "STAFF"
)

// Employee maps a um-api userId to a branch and local role
type Employee struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	BranchID    primitive.ObjectID `json:"branchId" bson:"branchId"`
	UserID      string             `json:"userId" bson:"userId"`
	Role        string             `json:"role" bson:"role"`
	CreatedBy   string             `json:"createdBy" bson:"createdBy"`
	CreatedDate time.Time          `json:"createdDate" bson:"createdDate"`
	UpdatedBy   string             `json:"updatedBy" bson:"updatedBy"`
	UpdatedDate time.Time          `json:"updatedDate" bson:"updatedDate"`
}
