package mongo

import (
	"context"
	"time"

	"github.com/app-devper/um-api/servicekit/tenant"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Seeder prepares a tenant's database on first use.
type Seeder func(ctx context.Context, db *mongo.Database) error

// Client keeps one database per tenant through servicekit/tenant (um-api
// ADR-0007): <prefix>_<clientId>, seeded on first use and retried later if
// seeding fails.
type Client struct {
	client  *mongo.Client
	tenants *tenant.Registry[*mongo.Database]
}

func NewClient(uri, dbPrefix string, seeder Seeder) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	open := func(name string) *mongo.Database { return client.Database(name) }
	var initialise func(context.Context, string, *mongo.Database) error
	if seeder != nil {
		initialise = func(ctx context.Context, clientID string, db *mongo.Database) error {
			if err := seeder(ctx, db); err != nil {
				return err
			}
			logrus.Infof("Opened database %q for client %q", db.Name(), clientID)
			return nil
		}
	}
	return &Client{client: client, tenants: tenant.New(dbPrefix, open, initialise)}, nil
}

func (c *Client) MongoClient() *mongo.Client {
	return c.client
}

func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

func (c *Client) ForClient(clientID string) (*mongo.Database, error) {
	return c.tenants.For(clientID)
}

func (c *Client) CollectionFromCtx(ctx context.Context, name string) (*mongo.Collection, error) {
	clientID, err := ClientIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	db, err := c.ForClient(clientID)
	if err != nil {
		return nil, err
	}
	return db.Collection(name), nil
}

// ValidateClientID refuses an empty or malformed client id.
func ValidateClientID(clientID string) error { return tenant.ValidateClientID(clientID) }

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
	CollectionProductItems       = "product_items"
	CollectionStockLogs          = "stock_logs"
)
