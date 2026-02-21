package mongo

import (
	"context"
	"time"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CollectionEmployees = "employees"

// EmployeeRepository implements repository.EmployeeRepository
type EmployeeRepository struct {
	collection *mongo.Collection
}

// NewEmployeeRepository creates a new EmployeeRepository
func NewEmployeeRepository(client *Client) *EmployeeRepository {
	repo := &EmployeeRepository{
		collection: client.Collection(CollectionEmployees),
	}
	repo.ensureIndexes()
	return repo
}

func (r *EmployeeRepository) ensureIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "branchId", Value: 1}},
	})
	r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "userId", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
}

// Create creates a new employee
func (r *EmployeeRepository) Create(ctx context.Context, emp *entity.Employee) error {
	emp.ID = primitive.NewObjectID()
	emp.CreatedDate = time.Now()
	emp.UpdatedDate = time.Now()
	_, err := r.collection.InsertOne(ctx, emp)
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
	var emp entity.Employee
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&emp)
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
	var emp entity.Employee
	err := r.collection.FindOne(ctx, bson.M{"userId": userID}).Decode(&emp)
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
	cursor, err := r.collection.Find(ctx, bson.M{})
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
	cursor, err := r.collection.Find(ctx, bson.M{"branchId": branchID})
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
	emp.UpdatedDate = time.Now()
	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": emp.ID}, emp)
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
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return entity.ErrNotFound
	}
	return nil
}
