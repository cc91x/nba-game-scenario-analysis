package helpers

import (
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
)

/* Globals */
var Logger *log.Logger
var Config *NbaConfig

/* Config specific variables */
var logFilePath string = "logs/nba-game-data-analysis.txt"
var configFileName string = "go_config.yaml"
var oddsSourceApiPath string = "/v4/historical/sports/basketball_nba/odds"

/* CSV generation specifics */
var csvDirectory string = "csvs"
var gamesCsvName string = "games_summary_data.csv"
var playsCsvName string = "game_play_by_play_data.csv"

/* Odds sourcing specifics */
var utcHoursForLookup = []int{16, 21, 23}

var bookmakersPriority map[string]int = map[string]int{
	"fanduel":        1,
	"draftkings":     2,
	"williamhill_us": 3,
	"betmgm":         4,
}

/* Process type parameter makeshift enum */
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

/* Database related constants */
// TODO: Rename DB
var mongoDbName = "local-nba-project"

var cleanedGamesCollectionName = "cleanedGameData"
var cleanedOddsCollectionName = "cleanedOdds"
var historicalOddsCollectionName = "rawHistoricalOdds"
var rawGamesCollectionName = "rawGames"
var teamMetadataCollectionName = "teamMetadata"

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