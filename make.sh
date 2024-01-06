#!/bin/bash

export BUILDCMD=build

#  SET LINKCMD=-ldflags="-s -w"
export LINKCMD=

echo $0 $1 $2
if [[ "$1" = "setup" ]]
then
    echo Hämtar beroenden...
    
    go get github.com/alexbrainman/odbc
    go get github.com/mattn/go-sqlite3
    go get golang.org/x/text/encoding/charmap
    go get github.com/shopspring/decimal
    go get github.com/extrame/xls
    go get github.com/xuri/excelize/v2
    go get github.com/pkg/browser
    
    exit 0
fi


rm wHHEK wHHEK.exe wHHEK_x86.exe wHHEK_x64.exe

if [[ "$1" = "clean" ]]
then
  rm *~
  rm html\*~
  rm #*#
  rm .#*
  rm *.ldb
  rm got*.mdb
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
    export BUILDCMD="build"
fi

if [[ "$1" = "release" ]]
then
    export LINKCMD="-ldflags -s -ldflags -w"
else
    export LINKCMD=
fi

echo Bygger...

if [[ "$1" = "test" ]]
then
    set -x
    go $BUILDCMD $LINKCMD
    set +x
else
    echo Build native
    set -x
    go $BUILDCMD $LINKCMD
    set +x
    echo Build win32
    export GOOS="windows"
    export GOARCH="386"
    set -x
    go $BUILDCMD $LINKCMD -o wHHEK_x86.exe
    set +x
fi

echo Klar.
