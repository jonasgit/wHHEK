#!/bin/bash

export JETFILE=nojetdb.go

#  SET BUILDCMD=test -v
#  set TESTFILES=main_test.go personer_test.go konton_test.go transaktioner_test.go db_test.go

export BUILDCMD=build
export TESTFILES=

#  SET LINKCMD=-ldflags="-s -w"
export LINKCMD=

ECHO Hämtar beroenden...

go get github.com/mattn/go-sqlite3
go get golang.org/x/text/encoding/charmap
go get github.com/shopspring/decimal

ECHO Bygger...

export SOURCEFILES="main.go platser.go transaktioner.go fastatransaktioner.go personer.go konton.go budget.go"

go $BUILDCMD $LINKCMD -o wHHEK $SOURCEFILES $JETFILE $TESTFILES

ECHO Klar.

