//
//
//

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"
)

func loadFile(filename string) ([]byte, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || errors.Is(err, os.ErrNotExist) == false
}

func makeDate(date string, format string) (string, error) {
	tm, err := time.Parse(format, date)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02dZ",
		tm.Year(), tm.Month(), tm.Day(), tm.Hour(), tm.Minute(), tm.Second()), nil
}

func logDebug(msg string) {
	if logLevel == "D" {
		log.Printf("DEBUG: %s", msg)
	}
}

func logInfo(msg string) {
	if logLevel == "D" || logLevel == "I" {
		log.Printf("INFO: %s", msg)
	}
}

func logWarning(msg string) {
	if logLevel == "D" || logLevel == "I" || logLevel == "W" {
		log.Printf("WARNING: %s", msg)
	}
}

func logError(msg string) {
	log.Printf("ERROR: %s", msg)
}

func logAlways(msg string) {
	log.Printf("INFO: %s", msg)
}

//
// end of file
//
