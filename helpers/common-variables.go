package helpers

import "errors"

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
