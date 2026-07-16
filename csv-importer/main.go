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

// main entry point
func main() {

	var csvFile string
	var assetDir string
	var namespace string
	var dryRun bool
	var noFiles bool
	var limit int
	var debug bool
	var logger *log.Logger

	flag.StringVar(&csvFile, "csvFile", "", "input CSV file")
	flag.StringVar(&assetDir, "assetDir", "", "asset directory (default to input file location)")
	flag.StringVar(&namespace, "namespace", "", "namespace to import")
	flag.BoolVar(&dryRun, "dryRun", false, "dry run only")
	flag.BoolVar(&noFiles, "noFiles", false, "no files imported")
	flag.BoolVar(&debug, "debug", false, "log debug information")
	flag.IntVar(&limit, "limit", 0, "limit import count (default is no limit)")
	flag.StringVar(&logLevel, "loglevel", "E", "Logging level (D|I|W|E)")
	flag.Parse()

	// check the required values
	if len(csvFile) == 0 || len(namespace) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if logLevel != "D" && logLevel != "I" && logLevel != "W" && logLevel != "E" {
		logError("logging level must be D|I|W|E")
		os.Exit(1)
	}

	// if this was not provided, default to the same as the CSV file
	if len(assetDir) == 0 {
		assetDir = filepath.Dir(csvFile)
	}

	if debug == true {
		logger = log.Default()
	}

	// open the input file
	f, err := os.Open(csvFile)
	if err != nil {
		logError(fmt.Sprintf("opening %s (%s)", csvFile, err))
		os.Exit(1)
	}
	defer f.Close()

	var proxyConfig uvaeasystore.EasyStoreProxyConfig
	proxyConfig = uvaeasystore.ProxyConfigImpl{
		ServiceEndpoint: os.Getenv("ESENDPOINT"),
		Log:             logger,
	}
	es, err := uvaeasystore.NewEasyStoreProxy(proxyConfig)

	if err != nil {
		logError(fmt.Sprintf("creating easystore (%s)", err.Error()))
		os.Exit(1)
	}

	// important, cleanup properly
	defer es.Close()

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
		if limit != 0 && okCount+errCount >= limit {
			logAlways(fmt.Sprintf("terminating after %d record(s)", limit))
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

		// make the object to import
		eso, err := makeEtdObject(namespace, assetDir, noFiles, record)

		if err != nil {
			logError(fmt.Sprintf("creating object (%s), continuing", err.Error()))
			errCount++
			continue
		}

		// if we are configured to import
		if dryRun == false {
			_, err = es.ObjectCreate(eso)
			if err != nil {
				logError(fmt.Sprintf("importing ns/oid [%s/%s] (%s), continuing", eso.Namespace(), eso.Id(), err.Error()))
				errCount++
				continue
			}
		}

		okCount++
	}

	verb := "imported"
	if dryRun == true {
		verb = "processed"
	}

	logAlways(fmt.Sprintf("%s %d object(s) and encountered %d error(s)", verb, okCount, errCount))
}

//
// end of file
//
