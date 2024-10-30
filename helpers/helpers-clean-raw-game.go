package helpers

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func gameIdFilter(gameId string) bson.M {
	return bson.M{"gameId": gameId}
}

func CleanGames(date string, config *NbaConfig) (err error) {

	client, err := LoadMongoDbClient(config)
	if err != nil {
		return err
	}
	// this needs work - with handling the error
	defer CloseMongoDBConnection(client, err)

	rawGamesCollection := getRawGamesCollection(client)
	rawGames, err := lookupGames(date, rawGamesCollection)
	if err != nil {
		return err
	}

	teamAbbrevIdMap, err := buildTeamIdMap(getTeamMetadataCollection(client))
	if err != nil {
		return err
	}

	// From ChatGPT - pre allocating the slice length
	var cleanedGames = make([]CleanedGame, 0, len(*rawGames))
	for _, rawGame := range *rawGames {
		cleanedGame, err := cleanGame(rawGame, teamAbbrevIdMap)
		if err != nil {
			return err
		}
		cleanedGames = append(cleanedGames, *cleanedGame)
	}
	return upsertItems(cleanedGames, getCleanedGamesCollection(client))
}

func lookupGames(date string, dbCollection *mongo.Collection) (rawGames *[]RawNbaGame, err error) {
	cursor, err1 := dbCollection.Find(context.TODO(), DateFieldStringFilter(date))
	err2 := cursor.All(context.TODO(), &rawGames)
	if err1 != nil || err2 != nil {
		return nil, errors.New("error fetching from raw games DB")
	}
	return
}

func buildTeamIdMap(dbCollection *mongo.Collection) (*map[string]string, error) {
	teamMetadata, err := FetchTeamMetadata(dbCollection)
	if err != nil {
		return nil, err
	}
	teamIdMap := make(map[string]string)
	for _, result := range *teamMetadata {
		teamIdMap[result.TeamAbbreviaton] = strconv.Itoa(result.TeamId)
	}
	return &teamIdMap, nil
}

func cleanGame(game RawNbaGame, teamIds *map[string]string) (cleanedGame *CleanedGame, err error) {
	awayTeam, homeTeam, err2 := extractTeamsFromMatchup(game.Matchup, teamIds)
	startTime, processedPlayByPlay, err1 := processPlayByPlay(game)
	if err1 != nil || err2 != nil {
		return nil, HandleErrors(err1, err2)
	}

	return &CleanedGame{
		GameId:     game.Parameters.GameId,
		Date:       game.Date,
		StartTime:  startTime,
		AwayTeamId: awayTeam,
		HomeTeamId: homeTeam,
		PlayByPlay: *processedPlayByPlay,
		SeasonId:   game.SeasonId,
	}, nil
}

func upsertItems(cleanedGames []CleanedGame, dbCollection *mongo.Collection) error {
	var operations []mongo.WriteModel
	for _, doc := range cleanedGames {
		operations = append(operations, mongo.NewUpdateOneModel().
			SetFilter(gameIdFilter(doc.GameId)).
			SetUpdate(bson.M{"$set": doc}).
			SetUpsert(true))
	}
	_, err := UpsertItemsGeneric(operations, dbCollection)
	return err
}

func processPlayByPlay(game RawNbaGame) (startTime string, playByPlay *[]PlayByPlay, err error) {
	var plays bson.A = game.PlayByPlayRows
	var err1 error
	prevTimeInterval, awayScore, homeScore := int32(-30), 0, 0

	for i, element := range plays {
		rawPlay, err2 := extractRawPlayFields(element)
		if err2 != nil {
			return "", nil, err
		}

		if i == 0 {
			startTime = rawPlay.EstTime
		}

		elapsed, err3 := elapsedFromGameClock(rawPlay.GameClockTime, rawPlay.Quarter)
		if rawPlay.Score != "" {
			awayScore, homeScore, err1 = parseScores(rawPlay.Score)
		}

		if err1 != nil || err3 != nil {
			return "", nil, HandleErrors(err1, err3)
		}

		// TODO: confirm this works, bc cool if so, does this work?
		playByPlay := make([]PlayByPlay, 0)
		for prevTimeInterval+30 <= elapsed {
			playByPlay = append(playByPlay, PlayByPlay{
				SecondsElapsed: int32(prevTimeInterval + 30),
				AwayScore:      awayScore,
				HomeScore:      homeScore,
			})
			prevTimeInterval += 30
		}
	}
	return
}

func parseScores(rawScore string) (awayScore int, homeScore int, err error) {
	split := strings.Split(rawScore, " - ")

	awayScore, err1 := strconv.Atoi(split[0])
	homeScore, err2 := strconv.Atoi(split[1])

	if err1 != nil || err2 != nil {
		return 0, 0, errors.New("error trying to parse game score. Expected \"# - #\"")
	}
	return
}

func elapsedFromGameClock(clockTime string, quarter int32) (secondsElapsed int32, err error) {
	digits := strings.Split(clockTime, ":")
	mins, err1 := strconv.Atoi(digits[0])
	seconds, err2 := strconv.Atoi(digits[1])

	if err1 != nil || err2 != nil {
		return 0, HandleErrors(err1, err2)
	}

	elapsedInQuarter := int32(((11 - mins) * 60) + (60 - seconds))
	prevQuarters := (quarter - 1) * (60 * 12)
	return elapsedInQuarter + prevQuarters, nil
}

func extractRawPlayFields(element interface{}) (*RawPlay, error) {
	switch play := element.(type) {
	case primitive.A:
		quarter, ok1 := play[4].(int32)
		startTime, ok2 := play[5].(string)
		clock, ok3 := play[6].(string)
		score, _ := play[10].(string)

		if !ok1 || !ok2 || !ok3 {
			return nil, errors.New("error parsing play by play field")
		}

		return &RawPlay{
			EstTime:       startTime,
			GameClockTime: clock,
			Quarter:       quarter,
			Score:         score,
		}, nil
	default:
		return nil, errors.New("error processing game play by play data")
	}

}

// Matchup string will be in the format: 'ATL vs. BOS' or 'ATL @ BOS'
func extractTeamsFromMatchup(matchup string, teamIds *map[string]string) (awayTeam string, homeTeam string, err error) {
	err = errors.New("error processing team ids from listed matchup")

	dicedMatchup := strings.Split(matchup, " ")
	if len(dicedMatchup) != 3 {
		return
	}

	var awayAbbrev, homeAbbrev string
	matchupSymbol := dicedMatchup[1]
	if matchupSymbol == "@" {
		awayAbbrev, homeAbbrev = dicedMatchup[0], dicedMatchup[2]
	} else {
		awayAbbrev, homeAbbrev = dicedMatchup[2], dicedMatchup[0]
	}

	awayTeam, ok1 := *teamIds[awayAbbrev]
	homeTeam, ok2 := *teamIds[homeAbbrev]
	if !ok1 || !ok2 {
		return
	}

	return teamIds[awayAbbrev], teamIds[homeAbbrev], nil
}
