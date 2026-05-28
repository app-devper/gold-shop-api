package mongo

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var validClientID = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9_-]{0,48}[a-zA-Z0-9])?$`)

type Seeder func(ctx context.Context, db *mongo.Database) error

type tenantDB struct {
	db     *mongo.Database
	seedMu sync.Mutex
	seeded atomic.Bool
}

type Client struct {
	client   *mongo.Client
	dbPrefix string
	seeder   Seeder
	cache    sync.Map
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

	return &Client{
		client:   client,
		dbPrefix: dbPrefix,
		seeder:   seeder,
	}, nil
}

func (c *Client) MongoClient() *mongo.Client {
	return c.client
}

func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

func (c *Client) ForClient(clientID string) (*mongo.Database, error) {
	if err := ValidateClientID(clientID); err != nil {
		return nil, err
	}
	if v, ok := c.cache.Load(clientID); ok {
		entry := v.(*tenantDB)
		c.ensureSeeded(entry, clientID)
		return entry.db, nil
	}
	entry := &tenantDB{db: c.client.Database(c.dbName(clientID))}
	actual, _ := c.cache.LoadOrStore(clientID, entry)
	entry = actual.(*tenantDB)
	c.ensureSeeded(entry, clientID)
	return entry.db, nil
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

func (c *Client) dbName(clientID string) string {
	if clientID == "000" {
		return c.dbPrefix
	}
	return fmt.Sprintf("%s_%s", c.dbPrefix, clientID)
}

func (c *Client) ensureSeeded(t *tenantDB, clientID string) {
	if c.seeder == nil {
		return
	}
	if t.seeded.Load() {
		return
	}
	t.seedMu.Lock()
	defer t.seedMu.Unlock()
	if t.seeded.Load() {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	logrus.Infof("Opened database %q for client %q", t.db.Name(), clientID)
	if err := c.seeder(ctx, t.db); err != nil {
		logrus.Warnf("tenant %q: seeder failed (will retry on next request): %v", clientID, err)
		return
	}
	t.seeded.Store(true)
}

func ValidateClientID(clientID string) error {
	if clientID == "" {
		return errors.New("clientId is required")
	}
	if !validClientID.MatchString(clientID) {
		return fmt.Errorf("invalid clientId %q", clientID)
	}
	return nil
}

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
