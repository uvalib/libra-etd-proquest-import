//
//
//

package main

import (
	"fmt"
	"html"
	"path/filepath"
	"strings"
	"time"

	"github.com/uvalib/easystore/uvaeasystore"
	"github.com/uvalib/libra-metadata"
)

// PUB NUMBER,PDF FILENAME,AUTHOR,TITLE,CALL NUMBER,DEPARTMENT,WIKIDATA URI,ORCID URI,VIAF URI,VIRGO URL,YEAR,SCHOOL NAME,PUB DATE,DISS LANG,LANG CODE,DEGREE,DEGREE DESC,ISBN,PAGE COUNT,KEYWORD,ADVISORS 1,ADVISOR 1 URI,ADVISORS 2,ADVISOR 2 URI,ADVISORS 3,ADVISOR 3 URI,ADVISORS 4,ADVISOR 4 URI,COMMITTEE MEMBERS 1,COMMITTEE 1 URI,COMMITTEE MEMBERS 2,COMMITTEE 2 URI,COMMITTEE MEMBERS 3,COMMITTEE 3 URI,COMMITTEE MEMBERS 4,COMMITTEE 4 URI,COMMITTEE MEMBERS 5,COMMITTEE 5 URI,ABSTRACT,ABSTRACT LANG,PUBLIC NOTE,SUPPLEMENTAL FILE NAMES,PROQUEST AUTHOR,Author Match,PROQUEST TITLE
const (
	id = iota
	etd_filename
	author
	title
	call_number
	department
	wikidata_uri
	orcid_uri
	viaf_uri
	virgo_url
	year
	school_name
	pub_date
	language
	language_code
	degree
	degree_desc
	isbn
	page_count
	keywords
	advisor_1
	advisor_1_uri
	advisor_2
	advisor_2_uri
	advisor_3
	advisor_3_uri
	advisor_4
	advisor_4_uri
	committee_1
	committee_1_uri
	committee_2
	committee_2_uri
	committee_3
	committee_3_uri
	committee_4
	committee_4_uri
	committee_5
	committee_5_uri
	abstract
	abstract_lang
	public_note
	suplemental_filename
	proquest_author
	author_match
	proquest_title
	fieldCount
)

var licenseTextLookup = map[string]string{
	"CC0": "CC0 (permitting unconditional free use, with or without attribution)",
	"ARR": "All rights reserved by the author (no additional license for public reuse)",
}

var degreeTextLookup = map[string]string{
	//"":       "BA (Bachelor of Arts)",
	//"":       "BARH (Bachelor of Architectural History)",
	//"":       "BS (Bachelor of Science)",
	//"":       "BSC (Bachelor of Science in Commerce)",
	//"":       "BUEP (Bachelor of Urban and Environmental Planning)",
	"D.N.P.": "DNP (Doctor of Nursing Practice)",
	"Ed.D.":  "EDD (Doctor of Education)",
	//"":       "EDS (Education Specialist)",
	"M.A.": "MA (Master of Arts)",
	//"":       "MAPE (Master of Arts in Physics Education)",
	//"":       "MAR (Master of Architecture)",
	//"":       "MARH (Master of Architectural History)",
	//"":       "MCS (Master of Computer Science)",
	//"":       "ME (Master of Engineering)",
	"M.Ed.": "MED (Master of Education)",
	//"":       "MEP (Master of Engineering Physics)",
	//"":       "MFA (Master of Fine Arts)",
	//"":       "MLA (Master of Landscape Architecture)",
	//"":       "MMSE (Master of Materials Science and Engineering)",
	//"":       "MPP (Master of Public Policy)",
	"M.S.": "MS (Master of Science)",
	//"":       "MSDS (Master of Science in Data Science)",
	//"":       "MUEP (Master of Urban and Environmental Planning)",
	"Ph.D.": "PHD (Doctor of Philosophy)",
	//"":       "SJD (Doctor of Juridical Science)",
}

func makeEtdObject(config Config, record []string) (uvaeasystore.EasyStoreObject, error) {

	o := uvaeasystore.NewEasyStoreObject(config.Namespace, "")

	// import domain metadata
	domainMetadata, err := libraEtdMetadata(config.License, record)
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
	if config.NoFiles == false {

		// load the etd blob
		etdBlob, err := loadBlob(filepath.Join(config.AssetDir, record[etd_filename]), record[etd_filename])
		if err != nil {
			return nil, err
		}

		blobs := make([]uvaeasystore.EasyStoreBlob, 0)
		blobs = append(blobs, etdBlob)

		if len(record[suplemental_filename]) != 0 {
			supBlob, err := loadBlob(filepath.Join(config.AssetDir, record[suplemental_filename]), record[suplemental_filename])
			if err != nil {
				return nil, err
			}
			blobs = append(blobs, supBlob)
		}

		o.SetFiles(blobs)
	}

	return o, nil
}

func libraEtdMetadata(license string, record []string) (librametadata.ETDWork, error) {
	meta := librametadata.ETDWork{}

	//
	// standard field mapping
	//

	meta.Program = unknownIfEmpty(record[department])
	meta.RelatedURLs = splitAndTrim(record[virgo_url], "")
	meta.Language = record[language]
	meta.Notes = record[public_note]

	//
	// bit of character mapping
	//

	meta.Title = html.UnescapeString(record[title])
	meta.Abstract = html.UnescapeString(record[abstract])

	// DEBUG ONLY
	//if meta.Title != record[title] {
	//	logAlways(fmt.Sprintf("Title transform for ID: %s", record[id]))
	//	logAlways(fmt.Sprintf("Before [%s]", record[title]))
	//	logAlways(fmt.Sprintf("After  [%s]", meta.Title))
	//}
	//if meta.Abstract != record[abstract] {
	//	logAlways(fmt.Sprintf("Abstract transform for ID: %s", record[id]))
	//	logAlways(fmt.Sprintf("Before [%s]", record[abstract]))
	//	logAlways(fmt.Sprintf("After  [%s]", meta.Abstract))
	//}

	//
	// specialized field processing
	//

	v, ok := degreeTextLookup[record[degree]]
	if ok {
		meta.Degree = v
	} else {
		meta.Degree = record[degree]
		logWarning(fmt.Sprintf("no degree text mapping for [%s]", record[degree]))
	}

	// see if we have appropriate text for this license
	v, ok = licenseTextLookup[license]
	if ok {
		meta.License = v
	} else {
		meta.License = license
		logWarning(fmt.Sprintf("no license text mapping for [%s]", license))
	}

	// split with the specified field seperator
	meta.Keywords = splitAndTrim(record[keywords], "|")

	// we support only a single author, and it is required
	author1, err := makeEtdName(record[author], record[school_name])
	if err != nil {
		return meta, fmt.Errorf("author: %w", err)
	}
	if author1 == nil {
		return meta, fmt.Errorf("author name is empty")
	}
	author1.Department = record[department]
	author1.ORCID = record[orcid_uri]
	meta.Author = *author1

	// we support multiple advisors/contributors (advisors followed by committee members)
	advisorSet := make([]librametadata.ContributorData, 0)
	var cdata *librametadata.ContributorData
	for _, col := range []int{advisor_1, advisor_2, advisor_3, advisor_4,
		committee_1, committee_2, committee_3, committee_4, committee_5} {
		cdata, err = makeEtdName(record[col], record[school_name])
		if err != nil {
			return meta, fmt.Errorf("advisor/committee member: %w", err)
		}
		if cdata != nil {
			advisorSet = append(advisorSet, *cdata)
		}
	}

	meta.Advisors = advisorSet

	// DEBUG ONLY
	//if len(meta.Advisors) != 0 {
	//	logAlways(fmt.Sprintf("Adding %d advisor(s) for ID: %s", len(meta.Advisors), record[id]))
	//}

	return meta, nil
}

func libraEtdFields(meta librametadata.ETDWork, record []string) (uvaeasystore.EasyStoreObjectFields, error) {
	fields := uvaeasystore.DefaultEasyStoreFields()
	var err error

	// all imported items get these
	fields["disposition"] = "proquest"
	fields["source"] = "ingest"
	fields["source-id"] = fmt.Sprintf("proquest:%s", record[id])

	// all published ETD's get these
	fields["draft"] = "false"
	fields["invitation-sent"] = "imported"
	fields["submitted-sent"] = "imported"
	fields["sis-sent"] = "imported"

	fields["publish-date"], err = makeDate(record[pub_date], "2006-01-02")
	if err != nil {
		return fields, err
	}

	// set visibility depending on the license
	if strings.HasPrefix(meta.License, "CC0") == true {
		fields["default-visibility"] = "open"
	} else {
		fields["default-visibility"] = "uva"
		fields["embargo-release"], _ = embargoRelease(fields["publish-date"])
		fields["embargo-release-visibility"] = "open"
	}

	// maybe?
	fields["depositor"] = "dpg3k"
	fields["create-date"] = time.Now().UTC().Format(time.RFC3339)

	return fields, nil
}

// parse a name of the form "Last, First" or "Last, First, Suffix" (the suffix is appended to the
// last name). Returns nil for an empty name and an error for a name that cannot be parsed
func makeEtdName(name string, institution string) (*librametadata.ContributorData, error) {

	if len(strings.TrimSpace(name)) == 0 {
		return nil, nil
	}

	bits := strings.Split(name, ",")
	for i := range bits {
		bits[i] = strings.TrimSpace(bits[i])
	}

	if len(bits) < 2 || len(bits[0]) == 0 || len(bits[1]) == 0 {
		return nil, fmt.Errorf("cannot parse name [%s], expected \"Last, First\"", name)
	}

	// append any suffix (Jr., III, etc) to the last name
	sname := bits[0]
	for _, suffix := range bits[2:] {
		if len(suffix) != 0 {
			sname += " " + suffix
		}
	}

	cdata := librametadata.ContributorData{
		FirstName:   bits[1],
		LastName:    sname,
		Department:  "unknown",
		Institution: institution,
	}
	return &cdata, nil
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
