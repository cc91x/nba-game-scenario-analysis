package helpers

import (
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

var oddsSourceApiKey = "c528d650a3ba67786937ad9a771224cf"
var oddsSourceBaseUrl = "https://api.the-odds-api.com/v4/historical/sports/basketball_nba/odds/"

// TODO: Alphabetize
var mongoDbName = "local-nba-project"

var gamesCsvName string = "go-game-data.csv"
var playsCsvName string = "go-play-by-play-data.csv"

var utcHoursForLookup = []int{16, 21, 23}

var bookmakersPriority map[string]int = map[string]int{
	"fanduel":        1,
	"draftkings":     2,
	"williamhill_us": 3,
	"betmgm":         4,
}

type ProcessType string

const (
	FetchRawOdds        ProcessType = "fetch_raw_odds"
	CleanRawOdds        ProcessType = "clean_raw_odds"
	CleanAllGames       ProcessType = "clean_games"
	CombineGameWithOdds ProcessType = "combine_game_and_odds"
)

func ValueOf(processName string) (ProcessType, error) {
	switch processName {
	case "fetch_raw_odds":
		return FetchRawOdds, nil
	case "clean_raw_odds":
		return CleanRawOdds, nil
	case "clean_games":
		return CleanAllGames, nil
	case "combine_game_and_odds":
		return CombineGameWithOdds, nil
	default:
		return "", errors.New("found unknown process type")
	}
}

var cleanedGamesCollectionName = "nba-cleaned-game-data"
var cleanedOddsCollectionName = "nba-cleaned-odds"
var historicalOddsCollectionName = "raw_historical_odds"
var rawGamesCollectionName = "nba-raw-games"
var teamMetadataCollectionName = "nba-team-id-mapping"

func getCleanedGamesCollection(client *mongo.Client) *mongo.Collection {
	return client.Database(mongoDbName).Collection(cleanedGamesCollectionName)
}

func getCleanedOddsCollection(client *mongo.Client) *mongo.Collection {
	return client.Database(mongoDbName).Collection(cleanedOddsCollectionName)
}

func getHistoricalOddscollection(client *mongo.Client) *mongo.Collection {
	return client.Database(mongoDbName).Collection(historicalOddsCollectionName)
}

func getRawGamesCollection(client *mongo.Client) *mongo.Collection {
	return client.Database(mongoDbName).Collection(rawGamesCollectionName)
}

func getTeamMetadataCollection(client *mongo.Client) *mongo.Collection {
	return client.Database(mongoDbName).Collection(teamMetadataCollectionName)
}
