package mongo

import (
	"context"
	"time"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const CollectionEmployees = "employees"

// EmployeeRepository implements repository.EmployeeRepository
type EmployeeRepository struct {
	client *Client
}

// NewEmployeeRepository creates a new EmployeeRepository
func NewEmployeeRepository(client *Client) *EmployeeRepository {
	return &EmployeeRepository{client: client}
}

func (r *EmployeeRepository) coll(ctx context.Context) (*mongo.Collection, error) {
	return r.client.CollectionFromCtx(ctx, CollectionEmployees)
}

// Create creates a new employee
func (r *EmployeeRepository) Create(ctx context.Context, emp *entity.Employee) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	emp.ID = primitive.NewObjectID()
	emp.CreatedDate = time.Now()
	emp.UpdatedDate = time.Now()
	_, err = coll.InsertOne(ctx, emp)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return entity.ErrDuplicateKey
		}
		return err
	}
	return nil
}

// GetByID retrieves an employee by ID
func (r *EmployeeRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Employee, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	var emp entity.Employee
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&emp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &emp, nil
}

// GetByUserID retrieves an employee by userId
func (r *EmployeeRepository) GetByUserID(ctx context.Context, userID string) (*entity.Employee, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	var emp entity.Employee
	err = coll.FindOne(ctx, bson.M{"userId": userID}).Decode(&emp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &emp, nil
}

// GetAll retrieves all employees
func (r *EmployeeRepository) GetAll(ctx context.Context) ([]*entity.Employee, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var employees []*entity.Employee
	if err := cursor.All(ctx, &employees); err != nil {
		return nil, err
	}
	if employees == nil {
		employees = []*entity.Employee{}
	}
	return employees, nil
}

// GetByBranchID retrieves employees by branch ID
func (r *EmployeeRepository) GetByBranchID(ctx context.Context, branchID primitive.ObjectID) ([]*entity.Employee, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	cursor, err := coll.Find(ctx, bson.M{"branchId": branchID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var employees []*entity.Employee
	if err := cursor.All(ctx, &employees); err != nil {
		return nil, err
	}
	if employees == nil {
		employees = []*entity.Employee{}
	}
	return employees, nil
}

// Update updates an employee
func (r *EmployeeRepository) Update(ctx context.Context, emp *entity.Employee) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	emp.UpdatedDate = time.Now()
	result, err := coll.ReplaceOne(ctx, bson.M{"_id": emp.ID}, emp)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return entity.ErrNotFound
	}
	return nil
}

// Delete deletes an employee
func (r *EmployeeRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	result, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return entity.ErrNotFound
	}
	return nil
}
