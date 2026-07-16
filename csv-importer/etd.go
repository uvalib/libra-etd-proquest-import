//
//
//

package main

import (
	"github.com/uvalib/easystore/uvaeasystore"
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
	keywords
	department
	abstract
	advisors
	advisorsUri
	fieldCount
)

func makeEtdObject(namespace string, assetDir string, noFiles bool, record []string) (uvaeasystore.EasyStoreObject, error) {

	o := uvaeasystore.NewEasyStoreObject(namespace, "123")
	return o, nil
}

//
// end of file
//
