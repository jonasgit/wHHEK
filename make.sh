#!/bin/bash
# Första gången      : ./make.sh setup
# Kör tester         : ./make.sh test
# Kör tester pratsamt: ./make.sh test verbose
# Bygg för användning: ./make.sh release
# Bygg för utveckling: ./make.sh
# Städa              : ./make.sh clean

ARG1=$1
ARG2=$2

rm -f wHHEK wHHEK.exe wHHEK_x86.exe wHHEK_x64.exe

if [[ "$ARG1" = "setup" ]]
then
    echo Hämtar beroenden...
    go mod tidy
    echo Klar.
    exit 0
fi

if [[ "$ARG1" = "clean" ]]
then
    rm -f -- *~ html/*~ '#'*'#' .#* *.ldb got*.mdb
    exit 0
fi

if [[ "$ARG1" = "test" ]]
then
    if [[ "$ARG2" = "verbose" ]]
    then
        BUILDCMD=(test -v)
    else
        BUILDCMD=(test)
    fi
else
    BUILDCMD=(build)
fi

if [[ "$ARG1" = "release" ]]
then
    LINKCMD=(-ldflags="-s -w")
else
    LINKCMD=()
fi

echo Bygger...

# go-sqlite3 kräver cgo
if [[ "$ARG1" = "test" ]]
then
    set -x
    GOOS=linux GOARCH=amd64 CGO_ENABLED=1 CC=gcc go "${BUILDCMD[@]}" "${LINKCMD[@]}"
    { set +x; } 2>/dev/null
else
    set -x
    GOOS=linux GOARCH=amd64 CGO_ENABLED=1 CC=gcc go "${BUILDCMD[@]}" "${LINKCMD[@]}" -o wHHEK
    GOOS=windows GOARCH=386 CGO_ENABLED=1 CC=i686-w64-mingw32-gcc go "${BUILDCMD[@]}" "${LINKCMD[@]}" -o wHHEK_x86.exe
    GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go "${BUILDCMD[@]}" "${LINKCMD[@]}" -o wHHEK_x64.exe
    { set +x; } 2>/dev/null
fi

echo Klar.
