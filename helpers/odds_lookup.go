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

// TODO: Is there a way to use init() and preload the config file, so I'm not passing it everywhere?
// TODO: Hide this from github so we protect data source...?
func OddsLookup(date string) (err error) {
	client, err := LoadMongoDbClient(*Config)
	if err != nil {
		return
	}
	rawOddsCollection := getHistoricalOddscollection(client)

	var (
		existingData  RawOddsResponse
		oddsResponses []RawOddsResponse
		rawOdds       *RawOddsResponse
	)

	for _, val := range utcHoursForLookup {
		err = rawOddsCollection.FindOne(context.TODO(), rawOddsDbFilter(date, val)).Decode(&existingData)
		if err == mongo.ErrNoDocuments {
			rawOdds, err = fetchOdds(date, val)
			oddsResponses = append(oddsResponses, *rawOdds)
		}
		if err != nil {
			return err
		}
	}
	_, err = upsertRawOddsRows(oddsResponses, rawOddsCollection)
	return err
}

// "https://api.the-odds-api.comv4/historical/sports/basketball_nba/odds?apiKey=c528d650a3ba67786937ad9a771224cf&markets=spreads,totals,h2h&regions=us&date=2024-10-22T16:00:00Z"
// TODO: Do I need to do something like &oddsResponse...?
func fetchOdds(date string, utcHour int) (oddsResponse *RawOddsResponse, err error) {
	urlString := buildOddsSourceUrl(date, strconv.Itoa(utcHour))
	response, err1 := http.Get(urlString)
	responseData, err2 := io.ReadAll(response.Body)
	err3 := json.Unmarshal(responseData, &oddsResponse)

	if err1 != nil || err2 != nil || err3 != nil {
		return nil, HandleErrors(err1, err2, err3)
	}
	oddsResponse.Date = date
	oddsResponse.UtcHour = utcHour
	return oddsResponse, nil
}

func upsertRawOddsRows(oddsResponse []RawOddsResponse, dbCollection *mongo.Collection) (*mongo.BulkWriteResult, error) {
	var operations = make([]mongo.WriteModel, 0, len(oddsResponse))
	for _, doc := range oddsResponse {
		operations = append(operations, mongo.NewUpdateOneModel().
			SetFilter(rawOddsDbFilter(doc.Date, doc.UtcHour)).
			SetUpdate(bson.M{"$set": doc}).
			SetUpsert(true))
	}
	return UpsertItemsGeneric(operations, dbCollection)
}

// TODO: Hide this from git
// Make the odds source path in the config?
func buildOddsSourceUrl(date string, utcHour string) string {
	return Config.OddsApi.BaseUrl + oddsSourceApiPath + "?apiKey=" + Config.OddsApi.Key + "&markets=spreads,totals,h2h&regions=us&date=" + date + "T" + utcHour + ":00:00Z"
}
