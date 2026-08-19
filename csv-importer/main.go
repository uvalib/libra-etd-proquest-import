package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/uvalib/easystore/uvaeasystore"
)

// global logging level
var logLevel string

type Config struct {
	InFile    string
	AssetDir  string
	Namespace string
	License   string
	Dryrun    bool
	NoFiles   bool
	StartRow  int
	Limit     int
	Debug     bool
}

// main entry point
func main() {

	var config Config
	var logger *log.Logger

	flag.StringVar(&config.InFile, "infile", "", "input file")
	flag.StringVar(&config.AssetDir, "assets", "", "asset directory (default to input file location)")
	flag.StringVar(&config.Namespace, "namespace", "", "namespace to import")
	flag.StringVar(&config.License, "license", "ARR", "CC0, ARR (all rights reserved)")
	flag.BoolVar(&config.Dryrun, "dryrun", false, "dry run only")
	flag.BoolVar(&config.NoFiles, "nofiles", false, "do not include files")
	flag.BoolVar(&config.Debug, "debug", false, "log debug information")
	flag.IntVar(&config.StartRow, "start", 1, "start row (default is 1, the first row)")
	flag.IntVar(&config.Limit, "limit", 0, "limit import count (default is no limit)")
	flag.StringVar(&logLevel, "loglevel", "E", "Logging level (D|I|W|E)")
	flag.Parse()

	// check the required values
	if len(config.InFile) == 0 || len(config.Namespace) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if config.Limit < 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if config.StartRow < 1 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if logLevel != "D" && logLevel != "I" && logLevel != "W" && logLevel != "E" {
		logError("logging level must be D|I|W|E")
		os.Exit(1)
	}

	// if this was not provided, default to the same as the CSV file
	if len(config.AssetDir) == 0 {
		config.AssetDir = filepath.Dir(config.InFile)
	}

	if config.Debug == true {
		logger = log.Default()
	}

	// open the input file
	f, err := os.Open(config.InFile)
	if err != nil {
		logError(fmt.Sprintf("opening %s (%s)", config.InFile, err))
		os.Exit(1)
	}
	defer f.Close()

	var proxyConfig uvaeasystore.EasyStoreProxyConfig
	var es uvaeasystore.EasyStore

	// don't need this if we are doing a dry run
	if config.Dryrun == false {
		proxyConfig = uvaeasystore.ProxyConfigImpl{
			ServiceEndpoint: os.Getenv("ESENDPOINT"),
			Log:             logger,
		}
		es, err = uvaeasystore.NewEasyStoreProxy(proxyConfig)

		if err != nil {
			logError(fmt.Sprintf("creating easystore (%s)", err.Error()))
			os.Exit(1)
		}

		// important, cleanup properly
		defer es.Close()
	}

	// new CSV reader
	reader := csv.NewReader(f)

	// read the header
	_, err = reader.Read()
	if err != nil {
		logError(fmt.Sprintf("reading header (%s)", err.Error()))
		os.Exit(1)
	}

	okCount := 0
	errCount := 0

	for {
		if config.Limit != 0 && okCount+errCount >= config.Limit {
			logAlways(fmt.Sprintf("terminating after %d record(s)", config.Limit))
			break
		}

		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			errCount++
			logError(fmt.Sprintf("reading row %d (%s), continuing", okCount+errCount, err.Error()))
			continue
		}

		// are we skipping any records?
		if okCount+errCount+1 >= config.StartRow {

			// make the object to import
			eso, err := makeEtdObject(config, record)

			if err != nil {
				logError(fmt.Sprintf("creating object (%s), continuing", err.Error()))
				errCount++
				continue
			}

			// if we are configured to import
			if config.Dryrun == false {
				_, err = es.ObjectCreate(eso)
				if err != nil {
					logError(fmt.Sprintf("importing ns/oid [%s/%s] (%s), continuing", eso.Namespace(), eso.Id(), err.Error()))
					errCount++
					continue
				}
			}

			okCount++
			logAlways(fmt.Sprintf("processed item %d ns/oid [%s/%s]", okCount+errCount, eso.Namespace(), eso.Id()))
		} else {
			okCount++
			logAlways(fmt.Sprintf("skipped item %d", okCount+errCount))
		}
	}

	verb := "imported"
	if config.Dryrun == true {
		verb = "processed"
	}

	logAlways(fmt.Sprintf("%s %d object(s) and encountered %d error(s)", verb, okCount, errCount))
}

//
// end of file
//
