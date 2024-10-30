package helpers

import (
	"context"
	"errors"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v3"
)

// load from config

func UpsertItemsGeneric(operations []mongo.WriteModel, dbCollection *mongo.Collection) (result *mongo.BulkWriteResult, err error) {
	if len(operations) == 0 {
		fmt.Println("Found 0 rows to upsert")
	} else {
		result, err := dbCollection.BulkWrite(context.TODO(), operations)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Upserted %v documents", result.UpsertedCount)
	}
	return
}

func LoadMongoDbClient(config *NbaConfig) (client *mongo.Client, err error) {
	clientUrl := config.Database.Name + "://" + config.Database.Name + ":" + config.Database.Port
	// clientUrl := "mongodb://localhost:27017"

	clientOptions := options.Client().ApplyURI(clientUrl)
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, errors.New("error connecting to mongoDb client")
	}

	return
}

func FetchTeamMetadata(teamInfoCollection *mongo.Collection) (results *[]TeamIdMapping, err error) {
	cursor, err1 := teamInfoCollection.Find(context.TODO(), bson.M{})
	err2 := cursor.All(context.TODO(), &results)

	if err1 != nil || err2 != nil {
		return nil, errors.New("error doing db lookup for team metadata")
	}
	return
}

func HandleErrors(errors ...error) error {
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}

func CloseMongoDBConnection(client *mongo.Client, err error) error {
	if err != nil {
		return err
	}
	if err = client.Disconnect(context.TODO()); err != nil {
		return err
	}
	return nil
}

func DateFieldStringFilter(date string) bson.M {
	return bson.M{"date": date}
}

// TODO: ensure this still works
func TernaryOperator[T any](condition bool, val1 T, val2 T) T {
	if condition {
		return val1
	} else {
		return val2
	}
}

func LookupGamesOnDate(date string, dbCollection *mongo.Collection) (games []CleanedGame, err error) {
	cursor, err1 := dbCollection.Find(context.TODO(), DateFieldStringFilter(date))
	err2 := cursor.All(context.TODO(), &games)
	if err1 != nil || err2 != nil {
		return nil, HandleErrors(err1, err2)
	}
	return
}

// TODO: from this article https://dev.to/ilyakaznacheev/a-clean-way-to-pass-configs-in-a-go-application-1g64
func ReadFile(configFileName string) (*NbaConfig, error) {
	f, err := os.Open(configFileName)

	if err != nil {
		return nil, errors.New("error reading config file")
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)

	var cfg NbaConfig
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, errors.New("error reading config file")
	}
	return &cfg, nil
}
