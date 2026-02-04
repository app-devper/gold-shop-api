package mongo

import (
	"context"
	"time"

	"github.com/devper-gold/gold-shop-api/internal/domain/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CustomerRepository implements repository.CustomerRepository
type CustomerRepository struct {
	collection *mongo.Collection
}

// NewCustomerRepository creates a new CustomerRepository
func NewCustomerRepository(client *Client) *CustomerRepository {
	return &CustomerRepository{
		collection: client.Collection(CollectionCustomers),
	}
}

// Create creates a new customer
func (r *CustomerRepository) Create(ctx context.Context, customer *entity.Customer) error {
	customer.CreatedAt = time.Now()
	customer.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, customer)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return entity.ErrDuplicateKey
		}
		return err
	}

	customer.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID retrieves a customer by ID
func (r *CustomerRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Customer, error) {
	var customer entity.Customer
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &customer, nil
}

// GetByMemberCode retrieves a customer by member code
func (r *CustomerRepository) GetByMemberCode(ctx context.Context, memberCode string) (*entity.Customer, error) {
	var customer entity.Customer
	err := r.collection.FindOne(ctx, bson.M{"member_code": memberCode}).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &customer, nil
}

// GetByRFID retrieves a customer by RFID card
func (r *CustomerRepository) GetByRFID(ctx context.Context, rfidCard string) (*entity.Customer, error) {
	var customer entity.Customer
	err := r.collection.FindOne(ctx, bson.M{"rfid_card": rfidCard}).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &customer, nil
}

// GetByPhone retrieves a customer by phone number
func (r *CustomerRepository) GetByPhone(ctx context.Context, phone string) (*entity.Customer, error) {
	var customer entity.Customer
	err := r.collection.FindOne(ctx, bson.M{"phone": phone}).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &customer, nil
}

// GetAll retrieves all customers with pagination
func (r *CustomerRepository) GetAll(ctx context.Context, limit, offset int) ([]*entity.Customer, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var customers []*entity.Customer
	if err := cursor.All(ctx, &customers); err != nil {
		return nil, err
	}
	return customers, nil
}

// Search searches customers by name or phone
func (r *CustomerRepository) Search(ctx context.Context, query string, limit int) ([]*entity.Customer, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"full_name": bson.M{"$regex": query, "$options": "i"}},
			{"phone": bson.M{"$regex": query, "$options": "i"}},
			{"member_code": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	opts := options.Find().SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var customers []*entity.Customer
	if err := cursor.All(ctx, &customers); err != nil {
		return nil, err
	}
	return customers, nil
}

// Update updates a customer
func (r *CustomerRepository) Update(ctx context.Context, customer *entity.Customer) error {
	customer.UpdatedAt = time.Now()

	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": customer.ID}, customer)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return entity.ErrNotFound
	}
	return nil
}

// Delete deletes a customer
func (r *CustomerRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return entity.ErrNotFound
	}
	return nil
}

// Count returns total customer count
func (r *CustomerRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}
