@echo off
SET ARG1=%1

FOR /F %%I IN ('go env GOOS') DO SET GOOS=%%I
FOR /F %%I IN ('go env GOARCH') DO SET GOARCH=%%I

IF NOT "%GOOS%"=="windows" (
  ECHO "Only intended for windows."
  EXIT
)

IF NOT "%GOARCH%"=="386" (
  ECHO Found 64-bit Go-compiler for Windows, disable JetDB/MDB.
  SET JETFILE=nojetdb.go
) else (
  ECHO Found 32-bit Go-compiler for Windows, enable JetDB/MDB.
  SET JETFILE=jetdb.go
)

IF "%ARG1%"=="setup" (
  ECHO Hämtar beroenden...

  go get github.com/alexbrainman/odbc
  go get github.com/mattn/go-sqlite3
  go get golang.org/x/text/encoding/charmap
  go get github.com/shopspring/decimal
  ECHO Klar.
  EXIT
)

IF "%ARG1%"=="test" (
  SET BUILDCMD=test -v -p 1
  set TESTFILES=main_test.go personer_test.go konton_test.go transaktioner_test.go db_test.go
) else (
  SET BUILDCMD=build
  set TESTFILES=
)

IF "%ARG1%"=="release" (
  SET LINKCMD=-ldflags="-s -w"
) else (
  SET LINKCMD=
)

ECHO Bygger...

set SOURCEFILES=main.go platser.go transaktioner.go fastatransaktioner.go personer.go konton.go budget.go

@echo on
go %BUILDCMD% %LINKCMD% -o wHHEK.exe %SOURCEFILES% %JETFILE% %TESTFILES%
@echo off

ECHO Klar.
