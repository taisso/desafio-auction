package auction

import (
	"context"
	"fmt"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func OpenConnection() (database *mongo.Database, close func()) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
		return
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mongo",
		Tag:        "latest",
	})
	if err != nil {
		log.Fatalf("Could not create mongo container: %s", err)
		return
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(
		fmt.Sprintf("mongodb://127.0.0.1:%s", resource.GetPort("27017/tcp"))))
	if err != nil {
		log.Println("Error trying to open connection")
		return
	}

	database = client.Database(os.Getenv("MONGODB_USER_DB"))
	close = func() {
		err := resource.Close()
		if err != nil {
			log.Println("Error trying to open connection")
			return
		}
	}

	return
}

func TestCreateAuction(t *testing.T) {
	os.Setenv("AUCTION_DURATION_SECONDS", "2")
	os.Setenv("MONGODB_USER_DB", "auction_test")

	database, close := OpenConnection()
	defer close()

	auctionRepository := NewAuctionRepository(database)

	auctionEntity, err := auction_entity.CreateAuction("auction 1", "category 1", "description 1", auction_entity.New)
	assert.Nil(t, err)
	err = auctionRepository.CreateAuction(context.TODO(), auctionEntity)

	time.Sleep(3 * time.Second)
	assert.Nil(t, err)

	var auctionEntityMongo AuctionEntityMongo
	database.Collection("auctions").FindOne(context.TODO(), bson.M{"_id": auctionEntity.Id}).Decode(&auctionEntityMongo)

	assert.Equal(t, auctionEntity.Id, auctionEntityMongo.Id)
	assert.Equal(t, auction_entity.Completed, auctionEntityMongo.Status)

}
