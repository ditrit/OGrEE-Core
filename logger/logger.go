package logger

import (
	"log"
	"os"
)

var (
	warningLogger         *log.Logger
	infoLogger            *log.Logger
	errorLogger           *log.Logger
	listenerInfoLogger    *log.Logger
	listenerWarningLogger *log.Logger
	listenerErrorLogger   *log.Logger
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

	infoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	warningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	listenerInfoLogger = log.New(f2, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	listenerWarningLogger = log.New(f2, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	listenerErrorLogger = log.New(f2, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func GetWarningLogger() *log.Logger {
	return warningLogger
}

func GetListenWarningLogger() *log.Logger {
	return listenerWarningLogger
}

func GetInfoLogger() *log.Logger {
	return infoLogger
}

func GetListenInfoLogger() *log.Logger {
	return listenerInfoLogger
}

func GetErrorLogger() *log.Logger {
	return errorLogger
}

func GetListenErrorLogger() *log.Logger {
	return listenerErrorLogger
}
