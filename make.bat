@echo off
REM Första gången      : ./make.bat setup
REM Kör tester         : ./make.bat test
REM Kör tester pratsamt: ./make.bat test verbose
REM Bygg för användning: ./make.sh release
REM Bygg för utveckling: ./make.sh

SET ARG1=%1
SET ARG2=%2

FOR /F %%I IN ('go env GOOS') DO SET GOOS=%%I
FOR /F %%I IN ('go env GOARCH') DO SET GOARCH=%%I

IF NOT "%GOOS%"=="windows" (
  ECHO "Only intended for windows."
  EXIT
)

DEL wHHEK.exe


IF "%ARG1%"=="setup" (
  ECHO Hämtar beroenden...

  go get github.com/alexbrainman/odbc
  go get github.com/mattn/go-sqlite3
  go get golang.org/x/text/encoding/charmap
  go get github.com/shopspring/decimal
  go get github.com/extrame/xls
  go get github.com/xuri/excelize/v2
  ECHO Klar.
  EXIT
)

IF "%ARG1%"=="test" (
  IF "%ARG2%"=="verbose" (
    SET BUILDCMD=test -p 1 -v
  ) else (
    SET BUILDCMD=test -p 1
  )
) else (
  SET BUILDCMD=build -o wHHEK.exe
)

IF "%ARG1%"=="release" (
  SET LINKCMD=-ldflags="-s -w"
) else (
  SET LINKCMD=
)

ECHO Bygger...

@echo on
go %BUILDCMD% %LINKCMD%
@echo off

ECHO Klar.
