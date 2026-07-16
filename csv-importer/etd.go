//
//
//

package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

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

	// default field mapping
	meta.Program = record[department]
	meta.Degree = record[degree]
	meta.Title = record[title]
	meta.Abstract = record[abstract]
	//meta.License = record[license]
	meta.Language = record[language]

	meta.Keywords = strings.Split(record[keywords], ",")

	fmt.Printf("Authors:  [%s]\n", record[authors])
	fmt.Printf("Advisors: [%s]\n", record[advisors])
	fmt.Printf("Year:     [%s]\n", record[year])
	fmt.Printf("License:  [%s]\n", record[license])

	return meta, nil
}

func libraEtdFields(meta librametadata.ETDWork, record []string) (uvaeasystore.EasyStoreObjectFields, error) {
	fields := uvaeasystore.DefaultEasyStoreFields()

	// all imported items get these
	fields["disposition"] = "imported"
	fields["invitation-sent"] = "imported"

	// all published ETD's get these
	fields["draft"] = "false"
	fields["submitted-sent"] = "imported"
	fields["sis-sent"] = "imported"

	return fields, nil
}

//
// end of file
//
