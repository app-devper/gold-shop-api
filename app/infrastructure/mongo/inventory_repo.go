package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	"github.com/devper-gold/gold-shop-api/app/domain/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type inventoryTransferRepository struct {
	client *Client
}

func NewInventoryTransferRepository(client *Client) repository.InventoryTransferRepository {
	return &inventoryTransferRepository{client: client}
}

func (r *inventoryTransferRepository) coll(ctx context.Context) (*mongo.Collection, error) {
	return r.client.CollectionFromCtx(ctx, CollectionInventoryTransfers)
}

func (r *inventoryTransferRepository) Create(ctx context.Context, transfer *entity.InventoryTransfer) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	transfer.ID = primitive.NewObjectID()
	transfer.CreatedAt = time.Now()
	transfer.UpdatedAt = time.Now()

	_, err = coll.InsertOne(ctx, transfer)
	return err
}

func (r *inventoryTransferRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.InventoryTransfer, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	var transfer entity.InventoryTransfer
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&transfer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &transfer, nil
}

func (r *inventoryTransferRepository) GetByTransferNumber(ctx context.Context, transferNumber string) (*entity.InventoryTransfer, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	var transfer entity.InventoryTransfer
	err = coll.FindOne(ctx, bson.M{"transfer_number": transferNumber}).Decode(&transfer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &transfer, nil
}

func (r *inventoryTransferRepository) GetByFromBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.TransferStatus) ([]*entity.InventoryTransfer, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"from_branch_id": branchID}
	if len(status) > 0 {
		filter["status"] = bson.M{"$in": status}
	}

	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transfers []*entity.InventoryTransfer
	if err := cursor.All(ctx, &transfers); err != nil {
		return nil, err
	}
	return transfers, nil
}

func (r *inventoryTransferRepository) GetByToBranchID(ctx context.Context, branchID primitive.ObjectID, status []entity.TransferStatus) ([]*entity.InventoryTransfer, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"to_branch_id": branchID}
	if len(status) > 0 {
		filter["status"] = bson.M{"$in": status}
	}

	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transfers []*entity.InventoryTransfer
	if err := cursor.All(ctx, &transfers); err != nil {
		return nil, err
	}
	return transfers, nil
}

func (r *inventoryTransferRepository) Update(ctx context.Context, transfer *entity.InventoryTransfer) error {
	coll, err := r.coll(ctx)
	if err != nil {
		return err
	}
	transfer.UpdatedAt = time.Now()
	_, err = coll.UpdateOne(
		ctx,
		bson.M{"_id": transfer.ID},
		bson.M{"$set": transfer},
	)
	return err
}

func (r *inventoryTransferRepository) GenerateTransferNumber(ctx context.Context) (string, error) {
	coll, err := r.coll(ctx)
	if err != nil {
		return "", err
	}
	// Format: TR-YYYYMMDD-XXXX
	prefix := fmt.Sprintf("TR-%s", time.Now().Format("20060102"))

	count, err := coll.CountDocuments(ctx, bson.M{
		"transfer_number": bson.M{"$regex": "^" + prefix},
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%04d", prefix, count+1), nil
}
