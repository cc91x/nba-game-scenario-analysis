package helpers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO: Hide this from github so we protect data source...?
func OddsLookup(date string) (err error) {
	var (
		existingData  RawOddsResponse
		oddsResponses []RawOddsResponse
		rawOdds       *RawOddsResponse
	)

	client, err := LoadMongoDbClient()
	if err != nil {
		return
	}
	rawOddsCollection := client.Database(mongoDbName).Collection(historicalOddsDbName)

	for _, val := range utcHoursForLookup {
		err = rawOddsCollection.FindOne(context.TODO(), uniqueOddsFilter(date, val)).Decode(&existingData)
		if err == mongo.ErrNoDocuments {
			rawOdds, err = fetchOdds(date, val)
			oddsResponses = append(oddsResponses, *rawOdds)
		}
		if err != nil {
			return
		}
	}
	_, err = upsertRawOddsRows(oddsResponses, rawOddsCollection)
	return
}

// TODO: Do I need to do something like &oddsResponse...?
func fetchOdds(date string, utcHour int) (oddsResponse *RawOddsResponse, err error) {
	urlString := buildOddsSourceUrl(oddsSourceBaseUrl, oddsSourceApiKey, date, strconv.Itoa(utcHour))
	response, err1 := http.Get(urlString)
	responseData, err2 := io.ReadAll(response.Body)
	err3 := json.Unmarshal(responseData, &oddsResponse)

	if err1 != nil || err2 != nil || err3 != nil {
		return nil, HandleErrors(err1, err2, err3)
	}
	oddsResponse.Date = date
	oddsResponse.UtcHour = utcHour
	return
}

func upsertRawOddsRows(oddsResponse []RawOddsResponse, dbCollection *mongo.Collection) (*mongo.BulkWriteResult, error) {
	var operations = make([]mongo.WriteModel, 0, len(oddsResponse))
	for _, doc := range oddsResponse {
		operations = append(operations, mongo.NewUpdateOneModel().
			SetFilter(uniqueOddsFilter(doc.Date, doc.UtcHour)).
			SetUpdate(bson.M{"$set": doc}).
			SetUpsert(true))
	}
	return UpsertItemsGeneric(operations, dbCollection)
}

// TODO: Hide this from git
func buildOddsSourceUrl(baseUrl string, apiKey string, date string, utcHour string) string {
	// urlString := fmt.Sprintf("https://api.the-odds-api.com/v4/historical/sports/basketball_nba/odds/?apiKey=%s&markets=spreads,totals,h2h&regions=us&date=%sT%d:00:00Z", apiKey, date, utcHour)
	return baseUrl + "?apiKey=" + apiKey + "&markets=spreads,totals,h2h&regions=us&date=" + date + "T" + utcHour + ":00:00Z"
}

func uniqueOddsFilter(date string, hour int) bson.M {
	return bson.M{
		"date-str": date,
		"utc-hour": hour,
	}
}
