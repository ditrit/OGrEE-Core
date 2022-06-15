package logger

import (
	"log"
	"os"
)

var (
	WarningLogger         *log.Logger
	InfoLogger            *log.Logger
	ErrorLogger           *log.Logger
	ListenerInfoLogger    *log.Logger
	ListenerWarningLogger *log.Logger
	ListenerErrorLogger   *log.Logger
)

func InitLogs() {
	file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	f2, err2 := os.OpenFile("unitylog.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err2 != nil {
		log.Fatal(err2)
	}

	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	ListenerInfoLogger = log.New(f2, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ListenerWarningLogger = log.New(f2, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ListenerErrorLogger = log.New(f2, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	//InfoLogger.
}
