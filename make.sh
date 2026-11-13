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
	export BUILDCMD="test -v"
    else
	export BUILDCMD="test "
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

# go-sqlite3 kräver cgo
if [[ "$1" = "test" ]]
then
    set -x
    GOOS=linux GOARCH=amd64 CGO_ENABLED=1 CC=gcc go $BUILDCMD $LINKCMD
    set +x
else
    echo Build linux amd64
    set -x
    GOOS=linux GOARCH=amd64 CGO_ENABLED=1 CC=gcc go $BUILDCMD $LINKCMD -o wHHEK
    set +x
    echo Build win64
    set -x
    GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go $BUILDCMD $LINKCMD -o wHHEK_x64.exe
    set +x
    echo Build win32
    set -x
    GOOS=windows GOARCH=386 CGO_ENABLED=1 CC=i686-w64-mingw32-gcc go $BUILDCMD $LINKCMD -o wHHEK_x86.exe
    set +x
fi

echo Klar.
