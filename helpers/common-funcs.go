package helpers

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// load from config
/*
var oddsSourceApiKey = "c528d650a3ba67786937ad9a771224cf"
var oddsSourceBaseUrl = "https://api.the-odds-api.com/v4/historical/sports/basketball_nba/odds/"

// TODO: Alphabetize
var mongoDbName = "local-nba-project"
var historicalOddsDbName = "raw_historical_odds"
var rawGamesDbName = "nba-raw-games"
var cleanedOddsDbName = "nba-cleaned-odds"
var cleanedGamesDbName = "nba-cleaned-game-data"
var teamMetadataDbName = "nba-team-id-mapping"

var gamesCsvName string = "go-game-data.csv"
var playsCsvName string = "go-play-by-play-data.csv"

var utcHoursForLookup = []int{16, 21, 23}

var bookmakersPriority map[string]int = map[string]int{
	"fanduel":        1,
	"draftkings":     2,
	"williamhill_us": 3,
	"betmgm":         4,
}
*/

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

func LoadMongoDbClient() (client *mongo.Client, err error) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, errors.New("error connecting to mongoDb client")
	}

	return
}

func FetchTeamMetadata(teamInfoCollection *mongo.Collection) (results []TeamIdMapping, err error) {
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

func TernaryOperator(condition bool, val1 int, val2 int) int {
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
