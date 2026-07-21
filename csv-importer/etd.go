//
//
//

package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/uvalib/easystore/uvaeasystore"
	"github.com/uvalib/libra-metadata"
)

// ID,AUTHORS,URI,TITLE,Virgo URL,Virgo Subjects,DEGREE,LICENSE,YEAR,SCHOOL NAME,DISS LANG,ISBN,PAGE COUNT,PQSUBJ,KEYWORD,DEPARTMENT,ABSTRACT,ADVISORS,ADVISOR URI
const (
	id = iota
	authors
	uri
	title
	virgoUrl
	virgoSubjects
	degree
	license
	year
	schoolName
	language
	isbn
	pageCount
	subject
	keywords
	department
	abstract
	advisors
	advisorsUri
	fieldCount
)

var licenseTextLookup = map[string]string{
	"CC0": "CC0 (permitting unconditional free use, with or without attribution)",
}

func makeEtdObject(namespace string, assetDir string, nofiles bool, record []string) (uvaeasystore.EasyStoreObject, error) {

	o := uvaeasystore.NewEasyStoreObject(namespace, "")

	// import domain metadata
	domainMetadata, err := libraEtdMetadata(record)
	if err != nil {
		return nil, err
	}

	// import fields
	fields, err := libraEtdFields(domainMetadata, record)
	if err != nil {
		return nil, err
	}

	// log as necessary
	logEtdFields(fields)
	logEtdMetadata(domainMetadata)

	// serialize domain metadata
	buf, err := domainMetadata.Payload()
	if err != nil {
		return nil, err
	}

	// create our store metadata object
	metadata := uvaeasystore.NewEasyStoreMetadata(domainMetadata.MimeType(), buf)

	// assign fields and serialized metadata
	o.SetFields(fields)
	o.SetMetadata(metadata)

	// do we import files?
	if nofiles == false {

		fname := filepath.Join(assetDir, record[id]) + ".pdf"

		// make sure the file exists
		if fileExists(fname) == false {
			logError(fmt.Sprintf("file [%s] does not exist", fname))
			return nil, uvaeasystore.ErrFileNotFound
		}

		// attempt to load the file
		buf, err := loadFile(fname)
		if err != nil {
			logError(fmt.Sprintf("loading file (%s)", err.Error()))
			return nil, err
		}

		// attempt to determine the content type
		mt := http.DetectContentType(buf)

		// create the blob
		blob := uvaeasystore.NewEasyStoreBlob(record[id], mt, buf)
		blobs := []uvaeasystore.EasyStoreBlob{blob}
		o.SetFiles(blobs)

		logInfo(fmt.Sprintf("loaded file [%s]", fname))
	}

	return o, nil
}

func libraEtdMetadata(record []string) (librametadata.ETDWork, error) {
	meta := librametadata.ETDWork{}

	//
	// default field mapping
	//

	meta.Program = record[department]
	meta.Degree = record[degree]
	meta.Title = record[title]
	meta.Abstract = record[abstract]
	meta.Language = record[language]

	//
	// specialized field processing
	//

	// see if we have appropriate text for this license
	v, ok := licenseTextLookup[record[license]]
	if ok {
		meta.License = v
	} else {
		meta.License = record[license]
		logWarning(fmt.Sprintf("no license text mapping for [%s]", record[license]))
	}

	// split with the specified field seperator
	meta.Keywords = strings.Split(record[keywords], "|")

	// we support only a single author
	authorSet := makeEtdNames(record[authors])
	if len(authorSet) != 0 {
		meta.Author = authorSet[0]
	}

	advisorSet := makeEtdNames(record[advisors])
	if len(advisorSet) != 0 {
		meta.Advisors = advisorSet
	}

	//logDebug(fmt.Sprintf("authors:    [%s]", record[authors]))
	//logDebug(fmt.Sprintf("title:      [%s]", record[title]))
	//logDebug(fmt.Sprintf("degree:     [%s]", record[degree]))
	//logDebug(fmt.Sprintf("license:    [%s]", record[license]))
	//logDebug(fmt.Sprintf("year:       [%s]", record[year]))
	//logDebug(fmt.Sprintf("language:   [%s]", record[language]))
	//logDebug(fmt.Sprintf("keywords:   [%s]", record[keywords]))
	//logDebug(fmt.Sprintf("department: [%s]", record[department]))
	//logDebug(fmt.Sprintf("abstract:   [%s]", record[abstract]))
	//logDebug(fmt.Sprintf("advisors:   [%s]", record[advisors]))

	return meta, nil
}

func libraEtdFields(meta librametadata.ETDWork, record []string) (uvaeasystore.EasyStoreObjectFields, error) {
	fields := uvaeasystore.DefaultEasyStoreFields()

	// all imported items get these
	fields["disposition"] = "proquest"

	// all published ETD's get these
	fields["draft"] = "false"
	fields["invitation-sent"] = "imported"
	fields["submitted-sent"] = "imported"
	fields["sis-sent"] = "imported"

	// try and make a clean publish date string
	str, err := makeDate(record[year], "2006")
	if err == nil {
		fields["publish-date"] = str
	}

	// maybe?
	fields["depositor"] = "dpg3k"
	fields["create-date"] = time.Now().Format("2006-01-02")

	// for now...
	fields["default-visibility"] = "open"

	return fields, nil
}

func makeEtdNames(names string) []librametadata.ContributorData {

	cdata := make([]librametadata.ContributorData, 0)
	if len(names) != 0 {
		for n := range strings.SplitSeq(names, "|") {
			bits := strings.Split(n, ",")
			if len(bits) >= 2 {
				fname := strings.TrimSpace(bits[1])
				sname := strings.TrimSpace(bits[0])
				cdata = append(cdata, librametadata.ContributorData{FirstName: fname, LastName: sname})
			}
		}
	}

	return cdata
}

func logEtdMetadata(meta librametadata.ETDWork) {
	b, _ := meta.Payload()
	logDebug(fmt.Sprintf("metadata: %s", string(b)))
}

func logEtdFields(fields uvaeasystore.EasyStoreObjectFields) {
	for k, v := range fields {
		logDebug(fmt.Sprintf("field: %s=%s", k, v))
	}
}

//
// end of file
//
