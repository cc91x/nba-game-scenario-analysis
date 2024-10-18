package helpers

import "go.mongodb.org/mongo-driver/bson"

type CleanedGame struct {
	GameId     string
	Date       string
	StartTime  string
	AwayTeamId string
	HomeTeamId string
	PlayByPlay []PlayByPlay
	SeasonId   string
}

type RawPlay struct {
	EstTime       string
	GameClockTime string
	Quarter       int32
	Score         string
}

type PlayByPlay struct {
	SecondsElapsed int32
	AwayScore      int
	HomeScore      int
}

// These types will work for loading the game data
type RawNbaGame struct {
	Resource       string     `bson:"resource"`
	Parameters     Parameters `bson:"parameters"`
	PlayByPlayRows bson.A     `bson:"raw-play-by-play"` // should this just be a bson? ... or
	Date           string     `bson:"date"`
	Matchup        string     `bson:"matchup"` // this should be "Bos vs Mia", copied from the season data script
}

type Parameters struct {
	GameId     string `bson:"GameID"`
	StarPeriod int32  `bson:"StartPeriod"`
	EndPeriod  int32  `bson:"EndPeriod"`
}

type TeamIdMapping struct {
	TeamId          int    `bson:"team-id"`
	TeamName        string `bson:"team-name"`
	TeamAbbreviaton string `bson:"team-abbreviation"`
}

// hmmm this should be in the shape of the data returned, not mongo response
// looks like used here - https://tutorialedge.net/golang/consuming-restful-api-with-go/
/*
type oddsResponse struct {
}
*/
type RawOddsResponse struct {
	Timestamp         string     `json:"timestamp" bson:"timestamp"`
	PreviousTimestamp string     `json:"previous_timestamp" bson:"previous_timestamp"`
	NextTimestamp     string     `json:"next_timestamp" bson:"next_timestamp"`
	Data              []OddsData `json:"data" bson:"data"`
	Date              string     `json:"date-str" bson:"date-str"`
	UtcHour           int        `json:"utc-hour" bson:"utc-hour"`
}

type OddsData struct {
	Id           string      `json:"id" bson:"id"`
	SportKey     string      `json:"sport_key" bson:"sport_key"`
	SportTitle   string      `json:"sport_title" bson:"sport_title"`
	CommenceTime string      `json:"commence_time" bson:"commence_time"`
	HomeTeam     string      `json:"home_team" bson:"home_team"`
	AwayTeam     string      `json:"away_team" bson:"away_team"`
	Bookmakers   []Bookmaker `json:"bookmakers" bson:"bookmakers"`
}

type Bookmaker struct {
	Key        string   `json:"key" bson:"key"`
	Title      string   `json:"title" bson:"title"`
	LastUpdate string   `json:"last_update" bson:"last_update"`
	Markets    []Market `json:"markets" bson:"markets"`
}

type Market struct {
	Key        string    `json:"key" bson:"key" `
	LastUpdate string    `json:"last_update" bson:"last_update" `
	Outcome    []Outcome `json:"outcomes" bson:"outcomes" `
}

type Outcome struct {
	Name  string  `json:"name" bson:"name"`
	Price float64 `json:"price" bson:"price"`
	Point float64 `json:"point" bson:"point"`
}

type CleanedOdds struct {
	GameId string `bson:"gameid"`
	// Date      string - this is on the game id
	// StartTime string - this is on the game id
	// seasonId string - this is on the game id
	Bookmaker   string      `bson:"bookmaker"`
	MoneyLine   MoneyLine   `bson:"moneyline"`
	PointSpread PointSpread `bson:"pointspread"`
	Total       Total       `bson:"total"`
}

type Total struct {
	Total      float32 `bson:"total"`
	OverPrice  float32 `bson:"overPrice"`
	UnderPrice float32 `bson:"underPrice"`
}

type MoneyLine struct {
	AwayPrice float32 `bson:"awayPrice"`
	HomePrice float32 `bson:"homePrice"`
}

type PointSpread struct {
	AwaySpread float32 `bson:"awaySpread"`
	HomeSpread float32 `bson:"homeSpread"`
	AwayPrice  float32 `bson:"awayPrice"`
	HomePrice  float32 `bson:"homePrice"`
}

// game-id,game-date,start-time,away-team-init,away-team-id,home-team-init,home-team-id,away-ml,away-spread,home-ml,home-spread,total,away-final-score,home-final-score

type GameCsv struct {
	GameId               string
	Date                 string
	StartTime            string
	AwayTeamAbbreviation string
	AwayTeamId           string
	HomeTeamAbbreviation string
	HomeTeamId           string
	AwayMl               string // float32
	HomeMl               string // float32
	AwaySpread           string // float32
	HomeSpread           string // float32
	AwayFinalScore       string // float32
	HomeFinalScore       string // float32
}

// game_id,seconds_elapsed,away_score,home_score,underdog_score,favorite_score
type PlayByPlayCsv struct {
	GameId         string
	SecondsElapsed string // int
	AwayScore      string // int
	HomeScore      string // int
	UnderdogScore  string // int
	FavoriteScore  string // int
}
