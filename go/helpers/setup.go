/*  */

package helpers

import (
	"errors"
	"flag"
	"io"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

/* Funcs for setting up globals, parsing command arguments */
func Setup() (processName string, date string, logFile *os.File) {

	processNameArg := flag.String("process", "", "Specify the process to run")
	dateArg := flag.String("date", "", "Specify the game date to run")
	configArg := flag.String("config", "", "Specify the absolute path of the config file")
	flag.Parse()

	cfg, err := readConfigFile(*configArg)
	if err != nil {
		ErrorWithFailure(err)
	}
	Config = cfg

	file, err := initializeLogger(logFilePath)
	if err != nil {
		ErrorWithFailure(err)
	}
	return *processNameArg, *dateArg, file
}

func initializeLogger(filePath string) (*os.File, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, errors.New("error initializing logger")
	}

	multiWriter := io.MultiWriter(os.Stdout, file)
	Logger = log.New(multiWriter, "", log.Ldate|log.Ltime|log.Lshortfile)
	return file, nil
}

func readConfigFile(configFileName string) (*NbaConfig, error) {
	f, err := os.Open(configFileName)

	if err != nil {
		return nil, errors.New("error reading config file")
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)

	var cfg NbaConfig
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, errors.New("error reading config file")
	}
	return &cfg, nil
}
