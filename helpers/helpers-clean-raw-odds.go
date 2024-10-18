package helpers

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var timezoneEstName string = "America/New_York"
var timezoneUtcName string = "UTC"
var amSuffix string = ":00 AM"

var moneylineKey string = "h2h"
var spreadKey string = "spreads"
var totalKey string = "total"
var overOutcome string = "Over"

func ProcessRawOdds(date string) error {
	fmt.Println("In clean odds")

	client, err := LoadMongoDbClient()
	if err != nil {
		return err
	}
	defer CloseMongoDBConnection(client, err)

	// TODO: Move loacl to local-nba-project
	rawOddsCollection := client.Database(mongoDbName).Collection(historicalOddsDbName)
	cleanedOddsCollection := client.Database(mongoDbName).Collection(cleanedOddsDbName)
	cleanedGamesCollection := client.Database(mongoDbName).Collection(cleanedGamesDbName)
	teamMetadataCollection := client.Database(mongoDbName).Collection(teamMetadataDbName)

	// Need a better plan for how to go about this
	/*
	   1. Find games on the date
	   2. Find the odds at the latest time pre game
	   	- Need to convert the EST start time to UTC, then do lookup for game odds
	   	- Can try a mongo query with date, utc hour and team names (team names owuld be 1 level deep,e lse pre process)
	   3. Now we have a game and the odds to process
	   	- Follow ordering of fanduel, draftkings, etc. to populate odds fields
	   4. Persist
	*/

	teamIdsToNamesMap, err1 := fetchTeamNameToIds(teamMetadataCollection)
	gamesOnDate, err2 := LookupGamesOnDate(date, cleanedGamesCollection)
	if err1 != nil || err2 != nil {
		return HandleErrors(err1, err2)
	}

	var cleanedOdds = make([]CleanedOdds, 0, len(gamesOnDate))
	for _, game := range gamesOnDate {
		utcHour, err3 := determineLatestHourBeforeGame(game)
		rawOdds, err4 := lookupRawOdds(utcHour, game, teamIdsToNamesMap, rawOddsCollection)
		cleanedOdd, err5 := cleanRawOdds(rawOdds, game)

		if err3 != nil || err4 != nil || err5 != nil {
			return HandleErrors(err3, err4, err5)
		}
		cleanedOdds = append(cleanedOdds, *cleanedOdd)
	}

	return upsertGameOdds(cleanedOdds, cleanedOddsCollection)
}

func determineLatestHourBeforeGame(game CleanedGame) (result int, err error) {
	gameStartTime, err1 := convertDateTimeToStandard(game.StartTime, game.Date, timezoneEstName)
	for _, utcHour := range utcHoursForLookup {
		clockTime := strconv.Itoa(utcHour) + amSuffix
		utcStartTime, err2 := convertDateTimeToStandard(clockTime, game.Date, timezoneUtcName)

		if err1 != nil || err2 != nil {
			return 0, HandleErrors(err1, err2)
		}
		if gameStartTime.Sub(*utcStartTime).Minutes() > 0 {
			result = utcHour
		}
	}
	return
}

func isValidClockTimeString(clocktime string) bool {
	return len(clocktime) == 8 && clocktime[2] == ':' && (clocktime[5:] == " AM" || clocktime[5:] == " PM")
}

func isValidDateString(date string) bool {
	return len(date) == 10 && date[4] == '-' && date[7] == '-'
}

func extractTimeUnits(clockTime string) (minute int, hour int, err error) {
	if len(clockTime) == 7 {
		clockTime = "0" + clockTime
	}
	if !isValidClockTimeString(clockTime) {
		return 0, 0, errors.New("invalid clocktime string")
	}

	minute, err1 := strconv.Atoi(clockTime[3:5])
	hour, err2 := strconv.Atoi(clockTime[:2])
	if err1 != nil || err2 != nil {
		err = HandleErrors(err1, err2)
	}
	if clockTime[6:] == "PM" {
		hour += 12
	}
	return
}

func extractDateUnits(date string) (year int, month int, day int, err error) {
	if !isValidDateString(date) {
		return 0, 0, 0, errors.New("invalid date string")
	}

	vals := strings.Split(date, "-")
	year, err1 := strconv.Atoi(vals[0])
	month, err2 := strconv.Atoi(vals[1])
	day, err3 := strconv.Atoi(vals[2])
	if err1 != nil || err2 != nil || err3 != nil {
		err = HandleErrors(err1, err2, err3)
	}
	return
}

func convertDateTimeToStandard(clockTime string, date string, timezone string) (*time.Time, error) {
	minute, hour, err1 := extractTimeUnits(clockTime)
	year, month, day, err2 := extractDateUnits(date)
	loc, err3 := time.LoadLocation(timezone)

	if err1 != nil || err2 != nil || err3 != nil {
		return nil, HandleErrors(err1, err2, err3)
	}
	standardTime := time.Date(year, time.Month(month), day, hour, minute, 0, 0, loc)
	return &standardTime, nil
}

func lookupRawOdds(utcHour int, game CleanedGame, teamIdsToNamesMap map[string]string, dbCollection *mongo.Collection) (oddsData OddsData, err error) {
	awayTeamName, ok1 := teamIdsToNamesMap[game.AwayTeamId]
	homeTeamName, ok2 := teamIdsToNamesMap[game.HomeTeamId]

	var rawOdds RawOddsResponse
	err1 := dbCollection.FindOne(context.TODO(), rawOddsDbFilter(game.Date, utcHour)).Decode(&rawOdds)
	if err1 != nil || !ok1 || !ok2 {
		return OddsData{}, errors.New("could not find any games for this date and time")
	}

	for _, rawOddsGame := range rawOdds.Data {
		if rawOddsGame.AwayTeam == awayTeamName && rawOddsGame.HomeTeam == homeTeamName {
			return rawOddsGame, nil
		}
	}
	return OddsData{}, errors.New("could not find odds for this game")
}

func filterAndOrderBookmakers(allBooks []Bookmaker, bookmakersPriority map[string]int) []Bookmaker {
	var validBooks = make([]Bookmaker, 0, len(bookmakersPriority))
	for _, book := range allBooks {
		if _, ok := bookmakersPriority[book.Key]; ok {
			validBooks = append(validBooks, book)
		}
	}
	sort.Slice(validBooks, func(i, j int) bool {
		return bookmakersPriority[validBooks[i].Key] < bookmakersPriority[validBooks[j].Key]
	})
	return validBooks
}

func cleanRawOdds(odds OddsData, game CleanedGame) (cleanedOdds *CleanedOdds, err error) {
	validBooks := filterAndOrderBookmakers(odds.Bookmakers, bookmakersPriority)
	for _, bookmaker := range validBooks {
		if len(bookmaker.Markets) == 3 {
			ml, spread, total := extractOdds(bookmaker, odds.AwayTeam)
			return &CleanedOdds{
				GameId:      game.GameId,
				Bookmaker:   bookmaker.Key,
				MoneyLine:   ml,
				PointSpread: spread,
				Total:       total,
			}, nil
		}
	}
	return nil, errors.New("could not find valid bookmaker")
}

func extractOdds(bookmaker Bookmaker, awayTeam string) (ml MoneyLine, spread PointSpread, total Total) {
	for _, market := range bookmaker.Markets {
		switch market.Key {
		case moneylineKey:
			ml = createMoneyLine(market, awayTeam)
		case spreadKey:
			spread = createSpread(market, awayTeam)
		case totalKey:
			total = createTotal(market)
		default:
			fmt.Printf("Unknown market found: %s. Skipping", market.Key)
		}
	}
	return
}

func createMoneyLine(market Market, awayTeam string) (ml MoneyLine) {
	for _, outcome := range market.Outcome {
		if outcome.Name == awayTeam {
			ml.AwayPrice = float32(outcome.Price)
		} else {
			ml.HomePrice = float32(outcome.Price)
		}
	}
	return
}

func createSpread(market Market, awayTeam string) (spread PointSpread) {
	for _, outcome := range market.Outcome {
		if outcome.Name == awayTeam {
			spread.AwayPrice = float32(outcome.Price)
			spread.AwaySpread = float32(outcome.Point)
		} else {
			spread.HomePrice = float32(outcome.Price)
			spread.HomeSpread = float32(outcome.Price)
		}
	}
	return
}

func createTotal(market Market) (total Total) {
	for _, outcome := range market.Outcome {
		if outcome.Name == overOutcome {
			total.OverPrice = float32(outcome.Price)
		} else {
			total.UnderPrice = float32(outcome.Price)
		}
		total.Total = float32(outcome.Point)
	}
	return
}

func cleanedOddsGameFilter(odds CleanedOdds) bson.M {
	return bson.M{"game": odds.GameId}
}

func upsertGameOdds(odds []CleanedOdds, dbCollection *mongo.Collection) (err error) {
	var operations = make([]mongo.WriteModel, 0, len(odds))
	for _, game := range odds {
		operations = append(operations, mongo.NewUpdateOneModel().
			SetFilter(cleanedOddsGameFilter(game)).
			SetUpdate(bson.M{"$set": game}).
			SetUpsert(true))
	}
	_, err = UpsertItemsGeneric(operations, dbCollection)
	return
}

func fetchTeamNameToIds(dbCollection *mongo.Collection) (res map[string]string, err error) {
	results, err := FetchTeamMetadata(dbCollection)
	if err != nil {
		return nil, err
	}
	res = make(map[string]string)
	for _, result := range results {
		res[strconv.Itoa(result.TeamId)] = result.TeamName
	}
	return
}

func rawOddsDbFilter(date string, utcHour int) bson.M {
	return bson.M{"date-str": date, "utc-hour": utcHour}
}
