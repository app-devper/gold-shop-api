package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// MongoTransactionManager implements repository.TransactionManager using MongoDB sessions
type MongoTransactionManager struct {
	client *mongo.Client
}

// NewTransactionManager creates a new MongoTransactionManager
func NewTransactionManager(client *Client) *MongoTransactionManager {
	return &MongoTransactionManager{client: client.MongoClient()}
}

// WithTransaction executes fn inside an atomic MongoDB transaction.
// If fn returns an error the transaction is aborted; otherwise it is committed.
func (m *MongoTransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	session, err := m.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	})
	return err
}
