package mongo

import (
	"context"
	"regexp"
	"time"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CustomerRepository implements repository.CustomerRepository
type CustomerRepository struct {
	client *Client
}

// NewCustomerRepository creates a new CustomerRepository
func NewCustomerRepository(client *Client) *CustomerRepository {
	return &CustomerRepository{client: client}
}

func (r *CustomerRepository) coll(ctx context.Context) (*mongo.Collection, error) {
	return r.client.CollectionFromCtx(ctx, CollectionCustomers)
}

// Create creates a new customer
func (r *CustomerRepository) Create(ctx context.Context, customer *entity.Customer) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	customer.CreatedAt = time.Now()
	customer.UpdatedAt = time.Now()

	result, err := coll.InsertOne(ctx, customer)
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
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	var customer entity.Customer
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&customer)
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
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	var customer entity.Customer
	err = coll.FindOne(ctx, bson.M{"member_code": memberCode}).Decode(&customer)
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
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	var customer entity.Customer
	err = coll.FindOne(ctx, bson.M{"rfid_card": rfidCard}).Decode(&customer)
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
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	var customer entity.Customer
	err = coll.FindOne(ctx, bson.M{"phone": phone}).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) GetNamesByIDs(ctx context.Context, ids []primitive.ObjectID) (map[primitive.ObjectID]string, error) {
	names := make(map[primitive.ObjectID]string, len(ids))
	if len(ids) == 0 {
		return names, nil
	}
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	opts := options.Find().SetProjection(bson.M{"full_name": 1})
	cursor, err := coll.Find(ctx, bson.M{"_id": bson.M{"$in": ids}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var customers []*entity.Customer
	if err := cursor.All(ctx, &customers); err != nil {
		return nil, err
	}
	for _, c := range customers {
		names[c.ID] = c.FullName
	}
	return names, nil
}

// GetAll retrieves all customers with pagination
func (r *CustomerRepository) GetAll(ctx context.Context, limit, offset int) ([]*entity.Customer, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := coll.Find(ctx, bson.M{}, opts)
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
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	safeQuery := regexp.QuoteMeta(query)
	filter := bson.M{
		"$or": []bson.M{
			{"full_name": bson.M{"$regex": safeQuery, "$options": "i"}},
			{"phone": bson.M{"$regex": safeQuery, "$options": "i"}},
			{"member_code": bson.M{"$regex": safeQuery, "$options": "i"}},
		},
	}

	opts := options.Find().SetLimit(int64(limit))
	cursor, err := coll.Find(ctx, filter, opts)
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
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	customer.UpdatedAt = time.Now()

	result, err := coll.ReplaceOne(ctx, bson.M{"_id": customer.ID}, customer)
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

// Count returns total customer count
func (r *CustomerRepository) Count(ctx context.Context) (int64, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return 0, err
	}
	return coll.CountDocuments(ctx, bson.M{})
}
