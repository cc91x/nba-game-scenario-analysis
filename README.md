Add a nice README 

Using historical data

Background:
This project is designed to provide insight into the likelihood of end game results based on an in game scenario. If a heavily favored team is down in the 1st, 2nd or 3rd quarter, what is the chance they come back and win? If the favorite has a 30 point lead in the 3rd, do they expand on the lead or pull their starters and let the margin close? If the two teams combine for only 90 points in the first half, how likely is it they will hit the over? These are the types of scenarios the project seeks to address.

Method: 
The NBA makes their statistics accessible via APIs. Using the python project here swarpatel, we can lookup game logs at the play by play. We can then combine this pregame historical odds data - sourced from a different provider. Note that this project doesn't look at in game odds for a scenario, only pregame.

Setup:
Config file: 
The properties in /go/go_configl.yaml are needed for the go tasks. Input a MongoDB connection and Odds API key. 

MongoDB section
MongoDB is used to store the raw and processed data. Connect to a mongoDB instance that has these five collections defined:
cleanedGameData
cleanedOdds
rawGames
rawHistoricalOdds
teamMetadata
Note: the teamMetadata collection needs to be populated before running anything. See the /mongodb/teamMetadata.json for sample data. Specify the host, port and dbName in /go/go_config.yaml

Golang section - 
The golang portion of the project needs to be compiled. 
From the /go directory, run go build -o ../bin/nba_main .

-> should I move the DB name local-nba-project to the go_config file? probably



We use airflow to handle the data sourcing process. The process runs once, nightly, and collects data for the T+2 date.
Getting Started with Airflow: 
Install airflow if not already done - https://airflow.apache.org/docs/apache-airflow/stable/start.html
run: export AIRFLOW_HOME={PROJECT_HOME}/airflow - where PROJECT_HOME is this root directory
run: airflow users create --username admin --firstname test --lastname admin --role Admin --email test@test.com
run: did airflow db init 
in one terminal window: airflow scheduler 
in another: airflow webserver -p 8080

Update nba_project_dag.py, set variables PYTHON_PATH and PROJECT_HOME
Heading to http://localhost:8080/home should bring up the airflow UI, where we can trigger the nba_project_dag 

Analyzing data: 
The python script HistoricalAnalysis.py, found under /python, is where we answer the questions propsed under background. We can configure pregame and midgame parameters under AnalysisConfig.py, and this script will then print out the historical probabilities of the end game result. 


Running Tasks individually -

Python:

Python section - 
Add a requirements.txt file by pip freeze, and provide requirements.txt
Use Pip install for requirements

There are two relevant python scripts
    1. RawGameDataSourcing.py - given a date as command line argument, fetches play by play data from the NBA API for games played on (date - 2)
    2. HistoricalAnalysis.py - used to ... can tune parameters in AnalysisConfig.py 
These can be found and ran under the /python directory.  

Golang - add building logic:
There are four relevant golang processes
1.
2.
3.
4.

for example: go run 
 


