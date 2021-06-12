@echo off
SET ARG1=%1

FOR /F %%I IN ('go env GOOS') DO SET GOOS=%%I
FOR /F %%I IN ('go env GOARCH') DO SET GOARCH=%%I

IF NOT "%GOOS%"=="windows" (
  ECHO "Only intended for windows."
  EXIT
)

IF NOT "%GOARCH%"=="386" (
  ECHO "Only intended for 32-bit windows."
  EXIT
)

IF "%ARG1%"=="test" (
  SET BUILDCMD=test -v
  set TESTFILES=main_test.go personer_test.go konton_test.go transaktioner_test.go
) else (
  SET BUILDCMD=build
  set TESTFILES=
)

IF "%ARG1%"=="release" (
  SET LINKCMD=-ldflags="-s -w"
) else (
  SET LINKCMD=
)

ECHO Hämtar beroenden...

go get github.com/alexbrainman/odbc
go get github.com/mattn/go-sqlite3
go get golang.org/x/text/encoding/charmap
go get github.com/shopspring/decimal

ECHO Bygger...

set SOURCEFILES=main.go jetdb.go platser.go transaktioner.go fastatransaktioner.go personer.go konton.go budget.go

go %BUILDCMD% %LINKCMD% -o wHHEK.exe %SOURCEFILES% %TESTFILES%

ECHO Klar.
