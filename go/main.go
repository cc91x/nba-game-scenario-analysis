/*  */

package main

/*

go build -o ../bin/nba_min .

         lookup raw game data, save raw      --->    clean game data, save
		/																   \
date ---																	---> clean odds, save ---> combine game data with odds data into CSVs ---> analyze csv data
		\ 															       /
 		 lookup odds from source, save raw response ----------------------
*/

// How to run this? Do: go run .    from the terminal
// let's try a switch statement and pass in the process as an argument
// go run main.go --date= --process=fetch_raw_games

// TODOS:
// Update airflow cmd w new directory structure - and try to move the file here
// include raw games fetch python script here
// Sanity check python script processing
// Upload to git, hiding right amount of data
// Go over basic syntax

// DONE FROM TODO:
// Add log statements everywhere...
// Consistency with naming return values, etc. -> should name most functions
// Decide how to order funcs within a file
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

import (
	"nba/helpers"
)

// "args": ["--process=clean_games", "--date=2024-03-18"] -> this seems to be working in launch.json
// /bin/bash -c '/Users/ericwhitehead/desktop/clag/nba-project-post-mv/nba_processing_bin --process=clean_games --date=2024-10-22 --config=/Users/ericwhitehead/desktop/clag/nba-project-post-mv/go_config.yaml'

// --process=clean_games --date=2024-10-22 --config=go_config.yaml
// /bin/bash -c 'cd /Users/ericwhitehead/Desktop/clag/nba-project-post-mv/ && nba_processing_bin --process=clean_games --date=2024-10-22 --config=/Users/ericwhitehead/desktop/clag/nba-project-post-mv/go_config.yaml'

func main() {
	processName, date, logFile := helpers.Setup()
	defer logFile.Close()

	helpers.Logger.Printf("Running process: %s for date: %s", processName, date)

	processType, err := helpers.ValueOf(processName)
	switch processType {
	case helpers.CleanAllGames:
		err = helpers.CleanGames(date)
	case helpers.FetchRawOdds:
		err = helpers.FetchOdds(date)
	case helpers.CleanRawOdds:
		err = helpers.CleanOdds(date)
	case helpers.CombineGameWithOdds:
		err = helpers.CombineGamesAndOddsToCsv(date)
	default:
		helpers.Logger.Println("Incorrect process type parameter")
	}

	if err != nil {
		helpers.ErrorWithFailure(err)
	}
	helpers.Logger.Println("No errors detected. Exiting with success")
}
