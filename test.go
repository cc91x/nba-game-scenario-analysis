package main

/*
 - Go notes here
 - 3 distinct programs / tasks here

 - Odds lookup program
	- input a date as a program argument
	- query odds api with 3 UTC hours ...
		- look up mongoDb for date + utc hr combo, if exists, then don't query,
		- else do query and persist into DB

- Game cleaning program (depends on python game lookup script)
	- double check the raw odds format in Mongo, but once done:
	- clean the data and persist in mongo, as we have. Seems like a simple step

- Odds cleaning program (depends on the above 2)
	- input a date(don't need UTC, can do that on its own)
	- lookup in raw odds in Mongo
	- convert team names to IDs and do mongo lookup for cleaned games
	- for each game where we find a match for existing game, create a cleaned game object

go run test.go is working, so is debug

games to test -
0022300967, 0022300970, 0022300969



         lookup raw game data, save raw      --->    clean game data, save
		/																   \
date ---																	---> combine game data with odds data into CSVs ---> analyze csv data
		\ 															       /
 		 lookup odds from source, save raw response --->  clean odds, save

*/

// How to run this? Do: go run .    from the terminal
// let's try a switch statement and pass in the process as an argument
// go run main.go --date= --process=fetch_raw_games

// TODOS:
// Get the defer() func working for exiting - https://trstringer.com/golang-deferred-function-error-handling/
// Go over basic syntax
// Consistency with naming return values, etc.
// Decide how to order funcs within a file
// Make consistent * vs & in return types - should make all lists a pointer
// add seasons to everywhere. Add it to the python analysis file, then test with latest game data from 10-22-2024. Do this after we've finished everything else
// add config file ... could use something like viper, but how would we keep it isolated in the constants file, and not main?
//         -> I guess we can, using the init() function ... will have to try
//			-> now need to implement it everywhere
// More formal logging ... instead of fmt.println
// clean up db field names. Need to run aggregation pipelines on the fields I am naming myself with -. It's ok to leave the default fields in the format they come in

// DONE FROM TODO:
// make the process names an enum,
// add a comment depicting the flow of the data with the DB names
// Convert every var arr [] to a pre allocated size
// file organization, create a file for just constants, and another for helper funcs (right now they are one)
// return if else with my ternary operator
// move the db getter funcs to be one - but this is a bit of a stretch

import (
	"flag"
	"fmt"
	"nba/helpers"
)

func main() {

	cfg, err := helpers.ReadFile("go-config.yaml")
	if err != nil {
		panic(err)
	}
	fmt.Println(cfg.Database.Name)

	processName := flag.String("process", "", "Specify the process to run")
	date := flag.String("date", "", "Specify the game date to run")
	flag.Parse()
	fmt.Printf("Running process: %s for date: %s", *processName, *date)

	// "args": ["--process=clean_games", "--date=2024-03-18"] -> this seems to be working in launch.json
	processType, err := helpers.ValueOf(*processName)
	switch processType {
	case helpers.CleanAllGames:
		err = helpers.CleanGames(*date)
	case helpers.FetchRawOdds:
		err = helpers.OddsLookup(*date)
	case helpers.CleanRawOdds:
		err = helpers.ProcessRawOdds(*date)
	case helpers.CombineGameWithOdds:
		err = helpers.CombineGamesAndOddsToCsv(*date)
	default:
		fmt.Println("Incorrect process type parameter")
	}

	if err != nil {
		// Do something with the err here, for airflow
		panic(err)
	}

	// exit gracefully for airflow
	fmt.Println("Exiting program with success!")
}
