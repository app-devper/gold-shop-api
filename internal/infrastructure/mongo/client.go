package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client wraps MongoDB client
type Client struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewClient creates a new MongoDB client
func NewClient(uri, databaseName string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &Client{
		client:   client,
		database: client.Database(databaseName),
	}, nil
}

// Database returns the MongoDB database
func (c *Client) Database() *mongo.Database {
	return c.database
}

// Collection returns a MongoDB collection
func (c *Client) Collection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

// Close closes the MongoDB connection
func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

// Collection names
const (
	CollectionBranches           = "branches"
	CollectionUsers              = "users"
	CollectionCustomers          = "customers"
	CollectionProductCategories  = "product_categories"
	CollectionProducts           = "products"
	CollectionGoldPrices         = "gold_prices"
	CollectionSales              = "sales"
	CollectionPawns              = "pawns"
	CollectionGoldSavings        = "gold_savings"
	CollectionExpenseCategories  = "expense_categories"
	CollectionExpenses           = "expenses"
	CollectionInventoryTransfers = "inventory_transfers"
	CollectionRewards            = "rewards"
	CollectionRewardRedemptions  = "reward_redemptions"
)
