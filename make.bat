@echo off
REM F�rsta g�ngen      : ./make.bat setup
REM K�r tester         : ./make.bat test
REM K�r tester pratsamt: ./make.bat test verbose
REM Bygg f�r anv�ndning: ./make.bat release
REM Bygg f�r utveckling: ./make.bat

SET ARG1=%1
SET ARG2=%2

FOR /F %%I IN ('go env GOOS') DO SET GOOS=%%I
FOR /F %%I IN ('go env GOARCH') DO SET GOARCH=%%I

IF NOT "%GOOS%"=="windows" (
  ECHO "Only intended for windows."
  EXIT
)

DEL wHHEK.exe wHHEK_x86.exe wHHEK_x64.exe


IF "%ARG1%"=="setup" (
  ECHO H�mtar beroenden...
  go mod tidy
REM  go get github.com/alexbrainman/odbc
REM  go get github.com/mattn/go-sqlite3
REM  go get golang.org/x/text/encoding/charmap
REM  go get github.com/shopspring/decimal
REM  go get github.com/extrame/xls
REM  go get github.com/xuri/excelize/v2
REM  go get github.com/pkg/browser
  ECHO Klar.
  EXIT
)

IF "%ARG1%"=="clean" (
  del *~
  del html\*~
  del #*#
  del .#*
  del *.ldb
  del got*.mdb
  EXIT
)

IF "%ARG1%"=="test" (
  IF "%ARG2%"=="verbose" (
    SET BUILDCMD=test -p 1 -v
  ) else (
    SET BUILDCMD=test -p 1
  )
) else (
  SET BUILDCMD=build
)

IF "%ARG1%"=="release" (
  SET LINKCMD=-ldflags="-s -w"
) else (
  SET LINKCMD=
)

ECHO Bygger...

IF "%ARG1%"=="test" (
  @echo on
  go %BUILDCMD% %LINKCMD%
  @echo off
) else (
  @echo on
  SET GOARCH=386
  go %BUILDCMD% %LINKCMD% -o wHHEK_x86.exe
  SET GOARCH=amd64
  go %BUILDCMD% %LINKCMD% -o wHHEK_x64.exe
  @echo off
)

ECHO Klar.
