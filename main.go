package main

/*
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
// rename csv field names
// Add log statements everywhere...
// Go over basic syntax
// Consistency with naming return values, etc. -> should name most functions
// Decide how to order funcs within a file
// Make the project structure have an internal, cmd, and pkg parts to it? https://appliedgo.com/blog/go-project-layout

// DONE FROM TODO:
// make the process names an enum,
// add a comment depicting the flow of the data with the DB names
// Convert every var arr [] to a pre allocated size
// file organization, create a file for just constants, and another for helper funcs (right now they are one)
// return if else with my ternary operator
// move the db getter funcs to be one - but this is a bit of a stretch
// add config file ... could use something like viper, but how would we keep it isolated in the constants file, and not main?
//         -> I guess we can, using the init() function ... will have to try
//			-> now need to implement it everywhere
// Confirmed everything is passed by value when possible and reference when needed
// More formal logging ... instead of fmt.println
// Get the defer() func working for exiting - https://trstringer.com/golang-deferred-function-error-handling/
// rename files
// See if its worth making any of structs combined into one, like I do with the config struct... I know there are plenty of nested objects
// 		-> decided against it, but I can reformat as I wish
//

import (
	"flag"
	"fmt"
	"nba/helpers"
)

func main() {
	helpers.Logger.Println("test log statement")

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
		helpers.ErrorWithFailure(err)
	}

	// exit gracefully for airflow
	fmt.Println("Exiting program with success!")
}
