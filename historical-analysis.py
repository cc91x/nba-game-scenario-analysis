
"""
python3 historical-analysis.py 

This should open the gameCsv and PlayByPlay csv and put together a report  

- Should take the following arguments

Format - game-id,game-date,start-time,away-team-init,away-team-id,home-team-init,home-team-id,away-ml,away-spread,home-ml,home-spread,total,away-final-score,home-final-score
For gameCsv:
    - pregame odds range 
    - favored team ids 
    - underdog team ids 
    - months 
    - seasons

Format - game_id,seconds_elapsed,away_score,home_score,underdog_score,favorite_score
For play by play csv:
    - pregame odds range 
    - in game time range 
    - in game [underdog - favorite] margin
    - favored team ids  
    - underdog team ids 
    - months 
    - seasons 
    - favorite back to back (bonus) 
    - underdog back to back (bonus)

"""

import AnalysisConfig as cfg
from enum import Enum
import csv
import numpy as np 

CSV_DIRECTORY = '/Users/ericwhitehead/Desktop/clag/nba-project-post-mv'
GAME_SUMMARY_CSV = 'games_summary_data.csv'
PLAY_BY_PLAY_CSV = 'game_play_by_play_data.csv'

class AnalysisType(Enum):
    INGAME = 1
    PREGAME = 2

class FilterType(Enum):
    EQUALITY = 1
    IN_RANGE = 2

class FilterField(Enum):
    def __init__(self, csvIndex, value, filterType):
        self._csvIndex = csvIndex
        self._filterType = filterType
        self._value = value
        
    @property
    def csvIndex(self):
        return self._csvIndex

    @property 
    def filterType(self):
        return self._filterType

    @property
    def value(self):
        return self._value 
    

# Game Summary CSV row lookups 
def getUnderdogId(row):
    return row[4] if float(row[9]) > 0 else row[6]

def getFavoriteId(row):
    return row[4] if float(row[9]) <= 0 else row[6]

def getFavoriteSpread(row):
    return abs(float(row[9]))

def getFavoriteMoneyline(row):
    return float(row[7]) if float(row[9]) <= 0 else float(row[8])

# Play by Play CSV row lookups
def getFavoriteMargin(playsRow):
    return int(playsRow[4]) - int(playsRow[5])

def getTotal(playsRow):
    return int(playsRow[4]) + int(playsRow[5])


class GameFilterFields(FilterField):
    E_AWAY_TEAM_IDS = (lambda row: row[4], cfg.AWAY_TEAM_IDS, FilterType.EQUALITY)
    E_HOME_TEAM_IDS = (lambda row: row[6], cfg.HOME_TEAM_IDS, FilterType.EQUALITY) 
    E_UNDERDOG_TEAM_IDS = (getUnderdogId, cfg.UNDERDOG_TEAM_IDS, FilterType.EQUALITY)
    E_FAVORITE_TEAM_IDS = (getFavoriteId, cfg.FAVORITE_TEAM_IDS, FilterType.EQUALITY)
    E_FAVORITE_SPREAD_RANGE = (getFavoriteSpread, cfg.PREGAME_FAVORITE_SPREAD_RANGE, FilterType.IN_RANGE)
    E_FAVORITE_ML_RANGE = (getFavoriteMoneyline, cfg.PREGAME_FAVORITE_ML_RANGE, FilterType.IN_RANGE)
    E_MONTHS = (lambda row: int(row[1][2:4]), cfg.MONTHS, FilterType.EQUALITY)
    # E_SEASONS = (lambda row: row[1], SEASONS, FilterType.EQUALITY)

class PlayByPlayFilterFields(FilterField): 
    E_IN_GAME_ELAPSED_SECONDS_RANGE = (lambda row: int(row[1]), cfg.IN_GAME_ELAPSED_SECONDS_RANGE, FilterType.IN_RANGE)
    E_IN_GAME_FAVORITE_MARGIN_RANGE = (getFavoriteMargin, cfg.IN_GAME_FAVORITE_MARGIN_RANGE, FilterType.IN_RANGE)
    E_IN_GAME_TOTAL_RANGE = (getTotal, cfg.IN_GAME_TOTAL_RANGE, FilterType.IN_RANGE)
    

def checkCriteria(row, valueGetter, filterValues, filterType):
    if filterValues == []:
        return True
    
    val = valueGetter(row)
    if filterType == FilterType.EQUALITY:
        return val in filterValues
    else:
        lo, hi = filterValues
        return val >= lo and val <= hi
        
        
def loadCsvAndFilter(csvName, filterEnum):
    filteredRows = []

    with open(f'{CSV_DIRECTORY}/{csvName}', mode ='r') as file:
        rows = csv.reader(file)
        
        next(rows)
        for row in rows:
            if all(checkCriteria(row, f.csvIndex, f.value, f.filterType) for f in filterEnum):
                filteredRows.append(row)

    return filteredRows

def printStatistics(data):
    print(f'Sample size: {len(data)} games \n')
    print(f'Average: {np.average(data)}')
    print('Deciles -  ')
    for dec in [10, 20, 30, 40, 50, 60, 70, 80, 90]:
        print(f'{dec}% {round(np.percentile(data, dec), 2)} pts')

def processResults(gameCsvRows):
    totals = []
    favoriteMargins = []

    for game in gameCsvRows:
        favScore = int(game[11]) if float(game[9]) <= 0 else int(game[12])
        dogScore = int(game[12]) if float(game[9]) < 0 else int(game[11])
        total = favScore + dogScore
        totals.append(total)
        favoriteMargins.append(favScore - dogScore)
 
    print('ENDGAME TOTALS ANALYSIS') 
    printStatistics(totals)

    print('\n ENDGAME FAVORITE MARGINS ANALYSIS')
    printStatistics(favoriteMargins)


if __name__ == '__main__':

    analysisType = AnalysisType.INGAME
    gameCsvRows = loadCsvAndFilter(GAME_SUMMARY_CSV, GameFilterFields)

    if analysisType == AnalysisType.INGAME:
        playsCsvGameIds = set(map(lambda row: row[0], loadCsvAndFilter(PLAY_BY_PLAY_CSV, PlayByPlayFilterFields))) 
        gameCsvRows = list(filter(lambda row: row[0] in playsCsvGameIds, gameCsvRows))

    processResults(gameCsvRows)

