package helpers

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

/*
Should work on creating 2 csv's

	First:
		game-id,game-date,start-time,away-team-init,away-team-id,home-team-init,home-team-id,away-ml,away-spread,home-ml,home-spread,total,away-final-score,home-final-score
	Second:
		game_id,seconds_elapsed,away_score,home_score,underdog_score,favorite_score
*/

func CombineGamesAndOddsToCsv(date string) error {
	client, err := LoadMongoDbClient()
	if err != nil {
		return err
	}
	defer CloseMongoDBConnection(client, err)

	cleanedOddsCollection := client.Database(mongoDbName).Collection(cleanedOddsDbName)
	cleanedGamesCollection := client.Database(mongoDbName).Collection(cleanedGamesDbName)
	teamMetadataCollection := client.Database(mongoDbName).Collection(teamMetadataDbName)

	games, err1 := LookupGamesOnDate(date, cleanedGamesCollection)
	gameToOdds, err2 := lookupOddsForGames(games, cleanedOddsCollection)
	teamIdToAbbrev, err3 := fetchTeamIdsToAbbreviation(teamMetadataCollection)
	if err1 != nil || err2 != nil || err3 != nil {
		return HandleErrors(err1, err2)
	}

	gameCsv := make(map[string][]string)
	playsCsv := make(map[string][]string)
	for _, game := range games {
		gameRow := createGameCsv(game, gameToOdds[game.GameId], teamIdToAbbrev)
		gameCsv[gameCsvKeyFunc(gameRow)] = gameRow

		playsRow := createPlaysCsv(game, gameToOdds[game.GameId])
		for _, row := range playsRow {
			playsCsv[playsCsvKeyFunc(row)] = row
		}
	}

	err4 := upsertCsv(gamesCsvName, gameCsv, gameCsvKeyFunc)
	err5 := upsertCsv(playsCsvName, playsCsv, playsCsvKeyFunc)
	if err4 != nil || err5 != nil {
		return HandleErrors(err4, err5)
	}
	return nil
}

func gameCsvKeyFunc(row []string) string {
	return row[0]
}

func playsCsvKeyFunc(row []string) string {
	return row[0] + row[1]
}

func upsertCsv(csvName string, rowsToInsert map[string][]string, getKey func([]string) string) error {
	var numUpdatedRows int
	var numNewRows int

	readFile, err1 := os.Open(csvName)
	if err1 != nil {
		return err1
	}
	defer readFile.Close()

	reader := csv.NewReader(readFile)
	rows, err2 := reader.ReadAll()
	if err2 != nil {
		return err2
	}

	var newCsv [][]string
	for _, row := range rows {
		if val, exists := rowsToInsert[getKey(row)]; !exists {
			newCsv = append(newCsv, row)
		} else {
			newCsv = append(newCsv, val)
			numUpdatedRows += 1
			delete(rowsToInsert, getKey(row))
		}
	}
	for _, row := range rowsToInsert {
		newCsv = append(newCsv, row)
		numNewRows += 1
	}

	writeFile, err3 := os.Create(csvName)
	if err3 != nil {
		return err3
	}

	writer := csv.NewWriter(writeFile)
	defer writer.Flush()
	for _, row := range newCsv {
		if err4 := writer.Write(row); err4 != nil {
			return err4
		}
	}
	fmt.Printf("For csv %s inserted %d new rows and updated %d rows", csvName, numNewRows, numUpdatedRows)
	return nil
}

func cleanedGamesQueryFilter(games []CleanedGame) bson.M {
	// consider making this a generic map() call (as in stream map filter) and inline
	var gameIds []string
	for _, game := range games {
		gameIds = append(gameIds, game.GameId)
	}
	return bson.M{"gameid": bson.M{
		"$in": gameIds,
	}}
}

func lookupOddsForGames(games []CleanedGame, dbCollection *mongo.Collection) (res map[string]CleanedOdds, err error) {
	var odds []CleanedOdds
	cursor, err1 := dbCollection.Find(context.TODO(), cleanedGamesQueryFilter(games))
	err2 := cursor.All(context.TODO(), &odds)
	if err1 != nil || err2 != nil {
		return nil, HandleErrors(err1, err2)
	}

	res = make(map[string]CleanedOdds)
	for _, odd := range odds {
		res[odd.GameId] = odd
	}
	return
}

func createGameCsv(game CleanedGame, odds CleanedOdds, teamIdToAbbrev map[string]string) []string {
	awayScore, homeScore := extractFinalScore(game)

	return []string{
		game.GameId,
		game.Date,
		game.StartTime,
		teamIdToAbbrev[game.AwayTeamId],
		game.AwayTeamId,
		teamIdToAbbrev[game.HomeTeamId],
		game.HomeTeamId,
		strconv.FormatFloat(float64(odds.MoneyLine.AwayPrice), 'f', -2, 32),
		strconv.FormatFloat(float64(odds.MoneyLine.HomePrice), 'f', -2, 32),
		strconv.FormatFloat(float64(odds.PointSpread.AwaySpread), 'f', -1, 32),
		strconv.FormatFloat(float64(odds.PointSpread.HomeSpread), 'f', -1, 32),
		strconv.Itoa(awayScore),
		strconv.Itoa(homeScore),
	}
}

func extractFinalScore(game CleanedGame) (awayScore int, homeScore int) {
	lastPlay := game.PlayByPlay[len(game.PlayByPlay)-1]
	return lastPlay.AwayScore, lastPlay.HomeScore
}

func createPlaysCsv(game CleanedGame, odds CleanedOdds) (result [][]string) {
	awayIsFavored := odds.PointSpread.AwaySpread <= 0
	for _, play := range game.PlayByPlay {
		underdogScore := TernaryOperator(awayIsFavored, play.HomeScore, play.AwayScore)
		favoriteScore := TernaryOperator(!awayIsFavored, play.HomeScore, play.AwayScore)
		result = append(result, []string{
			game.GameId,
			strconv.Itoa(int(play.SecondsElapsed)),
			strconv.Itoa(play.AwayScore),
			strconv.Itoa(play.HomeScore),
			strconv.Itoa(underdogScore),
			strconv.Itoa(favoriteScore),
			strconv.Itoa(favoriteScore - underdogScore),
		})
	}
	return
}

func fetchTeamIdsToAbbreviation(dbCollection *mongo.Collection) (res map[string]string, err error) {
	results, err := FetchTeamMetadata(dbCollection)
	if err != nil {
		return nil, err
	}
	res = make(map[string]string)
	for _, result := range results {
		res[strconv.Itoa(result.TeamId)] = result.TeamAbbreviaton
	}
	return
}
