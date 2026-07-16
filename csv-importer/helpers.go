//
//
//

package main

import (
	"errors"
	"log"
	"os"
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
