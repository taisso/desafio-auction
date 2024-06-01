package auction

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}

type AuctionRepository struct {
	Collection      *mongo.Collection
	auctionInterval time.Duration
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	return &AuctionRepository{
		Collection:      database.Collection("auctions"),
		auctionInterval: getAuctionDuration(),
	}
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}
	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	go ar.auctionTimer(ctx, auctionEntity.Id)

	return nil
}

func (ar *AuctionRepository) auctionTimer(ctx context.Context, id string) {
	timer := time.NewTimer(ar.auctionInterval)
	for range timer.C {
		filter := bson.M{"_id": id, "status": auction_entity.Active}
		update := bson.M{"$set": bson.M{"status": auction_entity.Completed}}

		err := ar.Collection.FindOne(ctx, filter).Err()
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return
			}
			logger.Error("Error when trying to process auction closing", err)
			return
		}

		_, err = ar.Collection.UpdateOne(ctx, filter, update)
		if err != nil {
			logger.Error("Error when trying to process auction closing", err)
		}

	}
}

func getAuctionDuration() time.Duration {
	auctionSecond := os.Getenv("AUCTION_DURATION_SECONDS")
	auctionParse, err := strconv.Atoi(auctionSecond)

	if err != nil {
		return 5 * time.Second
	}

	return time.Duration(auctionParse) * time.Second
}
