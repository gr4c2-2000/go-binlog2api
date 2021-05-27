package scripts

import (
	"flag"
	"os"

	"github.com/google/logger"
)

//Logger its logger.Logger struct with configuration
var Logger *logger.Logger
var lf *os.File

const logPath = "../logs/logfile.log"

// LoggerInit responsible for createing a logger object
func LoggerInit() error {
	flag.Parse()
	lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	Logger = logger.Init("LoggerExample", false, false, lf)
	Logger.Info("logger...")
	return nil
}

//LoggerClose connection and file
func LoggerClose() {

	logger.Close()
	lf.Close()
}
