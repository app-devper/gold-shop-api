package app

import (
	"context"
	"time"

	"github.com/devper-gold/gold-shop-api/app/domain/entity"
	gold_price_app "github.com/devper-gold/gold-shop-api/app/feature/gold_price"
	mongoinfra "github.com/devper-gold/gold-shop-api/app/infrastructure/mongo"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newTenantSeeder(goldAPIClient gold_price_app.ExternalGoldPriceAPI) mongoinfra.Seeder {
	return func(ctx context.Context, db *mongo.Database) error {
		if err := ensureEmployeeIndexes(ctx, db); err != nil {
			logrus.Warnf("seed: ensure employee indexes: %v", err)
		}
		if err := seedHQBranch(ctx, db); err != nil {
			return err
		}
		return seedInitialGoldPrice(ctx, db, goldAPIClient)
	}
}

func ensureEmployeeIndexes(ctx context.Context, db *mongo.Database) error {
	coll := db.Collection(mongoinfra.CollectionEmployees)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "branchId", Value: 1}}},
		{
			Keys:    bson.D{{Key: "userId", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	return err
}

func seedHQBranch(ctx context.Context, db *mongo.Database) error {
	coll := db.Collection(mongoinfra.CollectionBranches)
	count, err := coll.CountDocuments(ctx, bson.M{"code": "HQ"})
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	branch := entity.NewBranch("HQ", "สำนักงานใหญ่", "กรุงเทพมหานคร", "02-000-0000")
	if _, err := coll.InsertOne(ctx, branch); err != nil {
		return err
	}
	logrus.Infof("seed: created default branch %q (%s) in %q", branch.Code, branch.Name, db.Name())
	return nil
}

func seedInitialGoldPrice(ctx context.Context, db *mongo.Database, apiClient gold_price_app.ExternalGoldPriceAPI) error {
	coll := db.Collection(mongoinfra.CollectionGoldPrices)
	count, err := coll.CountDocuments(ctx, bson.M{"is_active": true})
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	priceData := &gold_price_app.GoldPriceData{
		GoldBarBuy:       42350.00,
		GoldBarSell:      42450.00,
		GoldOrnamentBuy:  41850.00,
		GoldOrnamentSell: 42950.00,
	}
	if apiClient != nil {
		apiCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		fetched, err := apiClient.GetCurrentPrice(apiCtx)
		cancel()
		if err != nil {
			logrus.Warnf("seed: fetch gold price for %q failed, using defaults: %v", db.Name(), err)
		} else {
			priceData = fetched
		}
	}

	goldPrice := entity.NewGoldPrice(
		priceData.GoldBarBuy,
		priceData.GoldBarSell,
		priceData.GoldOrnamentBuy,
		priceData.GoldOrnamentSell,
		"api",
	)
	if _, err := coll.InsertOne(ctx, goldPrice); err != nil {
		return err
	}
	logrus.Infof("seed: created gold price for %q Bar(B:%.2f/S:%.2f) Ornament(B:%.2f/S:%.2f)",
		db.Name(), priceData.GoldBarBuy, priceData.GoldBarSell, priceData.GoldOrnamentBuy, priceData.GoldOrnamentSell)
	return nil
}
