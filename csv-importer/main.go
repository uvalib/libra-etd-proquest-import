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

	var inFile string
	var assets string
	var namespace string
	var dryrun bool
	var nofiles bool
	var limit int
	var debug bool
	var logger *log.Logger

	flag.StringVar(&inFile, "infile", "", "input file")
	flag.StringVar(&assets, "assets", "", "asset directory (default to input file location)")
	flag.StringVar(&namespace, "namespace", "", "namespace to import")
	flag.BoolVar(&dryrun, "dryrun", false, "dry run only")
	flag.BoolVar(&nofiles, "nofiles", false, "do not include files")
	flag.BoolVar(&debug, "debug", false, "log debug information")
	flag.IntVar(&limit, "limit", 0, "limit import count (default is no limit)")
	flag.StringVar(&logLevel, "loglevel", "E", "Logging level (D|I|W|E)")
	flag.Parse()

	// check the required values
	if len(inFile) == 0 || len(namespace) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if logLevel != "D" && logLevel != "I" && logLevel != "W" && logLevel != "E" {
		logError("logging level must be D|I|W|E")
		os.Exit(1)
	}

	// if this was not provided, default to the same as the CSV file
	if len(assets) == 0 {
		assets = filepath.Dir(inFile)
	}

	if debug == true {
		logger = log.Default()
	}

	// open the input file
	f, err := os.Open(inFile)
	if err != nil {
		logError(fmt.Sprintf("opening %s (%s)", inFile, err))
		os.Exit(1)
	}
	defer f.Close()

	var proxyConfig uvaeasystore.EasyStoreProxyConfig
	var es uvaeasystore.EasyStore

	// dont need this if we are doing a dry run
	if dryrun == false {
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
		eso, err := makeEtdObject(namespace, assets, nofiles, record)

		if err != nil {
			logError(fmt.Sprintf("creating object (%s), continuing", err.Error()))
			errCount++
			continue
		}

		// if we are configured to import
		if dryrun == false {
			_, err = es.ObjectCreate(eso)
			if err != nil {
				logError(fmt.Sprintf("importing ns/oid [%s/%s] (%s), continuing", eso.Namespace(), eso.Id(), err.Error()))
				errCount++
				continue
			}

			logInfo(fmt.Sprintf("imported ns/oid [%s/%s]", eso.Namespace(), eso.Id()))
		}

		okCount++
	}

	verb := "imported"
	if dryrun == true {
		verb = "processed"
	}

	logAlways(fmt.Sprintf("%s %d object(s) and encountered %d error(s)", verb, okCount, errCount))
}

//
// end of file
//
