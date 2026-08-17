//
//
//

package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/uvalib/easystore/uvaeasystore"
	librametadata "github.com/uvalib/libra-metadata"
)

func unknownIfEmpty(value string) string {
	return defaultIfEmpty(value, "unknown")
}

func defaultIfEmpty(value string, defaultValue string) string {
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func loadBlob(filename string, blobname string) (uvaeasystore.EasyStoreBlob, error) {

	// make sure the file exists
	if fileExists(filename) == false {
		logError(fmt.Sprintf("file [%s] does not exist", filename))
		return nil, uvaeasystore.ErrFileNotFound
	}

	// attempt to load the file
	buf, err := loadFile(filename)
	if err != nil {
		logError(fmt.Sprintf("loading file (%s)", err.Error()))
		return nil, err
	}

	// attempt to determine the content type
	mt := http.DetectContentType(buf)

	logInfo(fmt.Sprintf("loaded file [%s]", filename))

	// create the blob
	return uvaeasystore.NewEasyStoreBlob(blobname, mt, buf), nil
}

func addContributor(cset []librametadata.ContributorData, cdata *librametadata.ContributorData) []librametadata.ContributorData {

	if cdata != nil {
		return append(cset, *cdata)
	}
	return cset
}

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

func embargoRelease(date string) (string, error) {
	tm, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return "", err
	}
	tm = tm.AddDate(70, 0, 0)
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
