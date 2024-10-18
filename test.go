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
// Convert every var arr [] to a pre allocated size
// Get the defer() func working for exiting
// Go over basic syntax
// Consistency with naming return values, etc.
// Decide how to order funcs within a file
// Make consistent * vs & in return types
// add seasons to everywhere
// add config file
// clean up db field names, and consider making them constants
// file organization, create a file for just constants, and another for helper funcs (right now they are one)
// move the db getter funcs to be one
// return if else with my ternary operator
// see if it is possible to make short functions into one line variables. i.e func1 := func(s string) string { return s + "test" }

// DONE FROM TODO:
// make the process names an enum,
// add a comment depicting the flow of the data with the DB names

import (
	"flag"
	"fmt"
	"nba/helpers"
)

func main() {

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
