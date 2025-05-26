package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

func init() {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatal("Failed to create logs directory:", err)
	}

	// Create or open log file
	logFile, err := os.OpenFile(
		fmt.Sprintf("logs/app-%s.log", time.Now().Format("2006-01-02")),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}

	// Initialize loggers
	InfoLogger = log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// LogInfo logs information messages
func LogInfo(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	InfoLogger.Println(message)
	fmt.Printf("INFO: %s\n", message)
}

// LogError logs error messages
func LogError(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	ErrorLogger.Println(message)
	fmt.Printf("ERROR: %s\n", message)
}
