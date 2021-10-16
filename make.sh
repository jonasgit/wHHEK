#!/bin/bash

export BUILDCMD=build

#  SET LINKCMD=-ldflags="-s -w"
export LINKCMD=

echo $0 $1 $2
if [[ "$1" = "setup" ]]
then
    ECHO Hämtar beroenden...
    
    go get github.com/mattn/go-sqlite3
    go get golang.org/x/text/encoding/charmap
    go get github.com/shopspring/decimal
    go get github.com/extrame/xls
    go get github.com/xuri/excelize/v2
    
    exit 0
fi

if [[ "$1" = "test" ]]
then
    if [[ "$2" = "verbose" ]]
    then
	export BUILDCMD="test -p 1 -v"
    else
	export BUILDCMD="test -p 1"
    fi
else
    export BUILDCMD="build -o wHHEK"
fi

if [[ "$1" = "release" ]]
then
    export LINKCMD="-ldflags -s -ldflags -w"
else
    export LINKCMD=
fi

ECHO Bygger...

set -x
go $BUILDCMD $LINKCMD
set +x

echo Klar.
