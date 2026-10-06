# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

A one-shot Go CLI that imports ProQuest ETD (electronic thesis/dissertation) records from a CSV export into Libra's EasyStore. Each CSV row becomes one EasyStore object holding fields, serialized `librametadata.ETDWork` metadata, and the PDF (plus an optional supplemental file) as blobs. All code lives in `csv-importer/` as a single `main` package. There are no tests.

## Commands

Run these from `csv-importer/`:

```sh
make build      # darwin/amd64 binary with -race -> bin/libra-etd-proquest-import.darwin
make linux      # static linux/amd64 binary -> bin/libra-etd-proquest-import.linux
make vet        # go vet
make check      # staticcheck (all,-S1002,-ST1003) + shadow vet
make dep        # go get -u, go mod tidy, go mod verify
```

Typical run (do a dry run first; it skips the EasyStore connection and only builds and validates objects):

```sh
ESENDPOINT=<easystore proxy url> bin/libra-etd-proquest-import.darwin \
  -infile <drop-dir>/<file>.csv -namespace <namespace> \
  -dryrun -loglevel D -start 1 -limit 10
```

Flags: `-infile` and `-namespace` are required. The others are `-assets` (directory holding the PDFs; defaults to the CSV's directory), `-license CC0|ARR` (default ARR), `-nofiles`, `-start` (1-based data row), `-limit`, `-loglevel D|I|W|E`, and `-debug` (turns on EasyStore proxy logging). The `drop-*` directories hold the ProQuest data drops and are gitignored.

## Architecture

- `main.go`: parses flags, opens the EasyStore proxy (`uvaeasystore.NewEasyStoreProxy`, endpoint from the `ESENDPOINT` env var) unless `-dryrun` is set, then streams the CSV. It skips the header row and calls `es.ObjectCreate` for each row. A failure on one row is logged and counted, and the import moves on to the next row. Rows before `-start` are skipped and counted separately; `-limit` caps only the rows processed after `-start`.
- `etd.go`: the CSV-to-EasyStore mapping.
  - The `iota` const block names the CSV columns **by position**. The comment above it gives the expected header order. If ProQuest changes or reorders columns, update this block to match.
  - `libraEtdMetadata` builds the `ETDWork` metadata. It HTML-unescapes the title and abstract, maps degrees through `degreeTextLookup` (unmapped degrees pass through with a warning), splits keywords on `|`, and parses names as `"Last, First"`. Advisors and committee members are merged into `Advisors`, each with department `"unknown"`.
  - `libraEtdFields` sets the EasyStore fields: `source=ingest`, `source-id=proquest:<PUB NUMBER>`, `disposition=proquest`, the `*-sent=imported` notification flags, and a hardcoded `depositor`. CC0 records are `open`. Every other license gets `uva` visibility with a 70-year embargo that releases to `open`.
- `helpers.go`: blob loading (content type from `http.DetectContentType`), date formatting, and leveled logging (`logDebug`/`logInfo`/`logWarning`/`logError`/`logAlways`, gated by the global `logLevel`).

The `github.com/uvalib/easystore/uvaeasystore` and `github.com/uvalib/libra-metadata` modules are UVA Library packages. Look in them for the object, field, and metadata types.
