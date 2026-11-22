#!/bin/bash
# Installerar MDAC 2.8, MDAC 2.7, Jet 4.0 SP8 och ODBC-drivrutinen för Access i ett
# Wine-prefix så att wHHEK_x86.exe kan öppna mdb-filer i Linux.
#
# För Wine i ren WoW64-mode (64-bitars prefix, t.ex. Wine 11) där winetricks vägrar
# installera mdac28/jet40. Testat med Wine 11.22.
#
# Användning:
#   ./wine-setup.sh                         # prefix ~/.wine-hhek, wine från PATH
#   WINEPREFIX=~/.mittprefix ./wine-setup.sh
#   WINE=/väg/till/wine ./wine-setup.sh     # egenbyggd Wine
#
# Kräver: winetricks, curl, cabextract, sha256sum
# Kör sedan: WINEPREFIX=... wine wHHEK_x86.exe  (i katalogen med mdb-filen)

set -euo pipefail

export WINEPREFIX="${WINEPREFIX:-$HOME/.wine-hhek}"
export WINE="${WINE:-wine}"
export WINEDEBUG="${WINEDEBUG:--all}"
CACHE="${XDG_CACHE_HOME:-$HOME/.cache}/winetricks"

MDAC27_URL='https://web.archive.org/web/20060718123742/http://ftp.gunadarma.ac.id/pub/driver/itegno/USB%20Software/MDAC/MDAC_TYP.EXE'
MDAC27_SHA=36d2a3099e6286ae3fab181a502a95fbd825fa5ddb30bf09b345abc7f1f620b4
JET40_URL='https://web.archive.org/web/20210225171713/http://download.microsoft.com/download/4/3/9/4393c9ac-e69e-458d-9f6d-2fe191c51469/Jet40SP8_9xNT.exe'
JET40_SHA=b060246cd499085a31f15873689d5fa7df817e407c8261a5c71fa6b9f7042560

for cmd in winetricks curl cabextract sha256sum "$WINE"; do
    command -v "$cmd" >/dev/null || { echo "Saknas: $cmd" >&2; exit 1; }
done

echo "Wine-prefix: $WINEPREFIX"

# Laddar ner $1 till $2 om filen saknas och kontrollerar sha256 ($3).
download() {
    local url=$1 file=$2 sha=$3
    if [ ! -f "$file" ]; then
        mkdir -p "$(dirname "$file")"
        curl -fL -o "$file.part" "$url"
        mv "$file.part" "$file"
    fi
    echo "$sha  $file" | sha256sum -c --quiet - || { echo "Fel checksumma: $file" >&2; exit 1; }
}

# Packar upp ett MDAC_TYP.EXE (IExpress) och kör setup för hand.
install_mdac() {
    local name=$1 dir="$WINEPREFIX/drive_c/$1"
    echo "== Installerar $name"
    rm -rf "$dir"
    mkdir -p "$dir"
    (cd "$CACHE/$name" && "$WINE" MDAC_TYP.EXE /C "/T:C:\\$name" /Q)
    [ -f "$dir/setup.exe" ] || { echo "Uppackning av $name misslyckades (fungerar skärmen/DISPLAY?)" >&2; exit 1; }
    (cd "$dir" && "$WINE" setup.exe /qnt)
    rm -rf "$dir"
}

# Laddar ner MDAC 2.8 och sätter DLL-overrides. Själva installationen i winetricks
# misslyckas (status 2) i 64-bitars prefix, därför körs den för hand nedan.
echo "== winetricks mdac28 (förväntas misslyckas med status 2)"
winetricks -q --force mdac28 || true
[ -f "$CACHE/mdac28/MDAC_TYP.EXE" ] || { echo "winetricks laddade inte ner mdac28" >&2; exit 1; }

# winetricks jet40 avbryter innan nedladdning, så MDAC 2.7 och Jet hämtas direkt.
download "$MDAC27_URL" "$CACHE/mdac27/MDAC_TYP.EXE" "$MDAC27_SHA"
download "$JET40_URL" "$CACHE/jet40/Jet40SP8_9xNT.exe" "$JET40_SHA"

# MDAC-installationen kräver Windows NT 4.0
winetricks -q --force nt40

install_mdac mdac28
install_mdac mdac27

echo "== Installerar Windows Script Host och Jet 4.0 SP8"
winetricks -q wsh57
(cd "$CACHE/jet40" && "$WINE" Jet40SP8_9xNT.exe /q)

winetricks -q win10   # återställ Windows-versionen

# Jet-installationen registrerar inte ODBC-drivrutinen för Access, som wHHEK använder.
# Setup måste vara odbcjt32.dll (inte odbcji32.dll), annars blir felet
# "{01S00} Invalid connection string attribute".
echo "== Registrerar ODBC-drivrutinen för Access"
REG="$WINEPREFIX/drive_c/access-odbc.reg"
cat > "$REG" <<'EOF'
REGEDIT4

[HKEY_LOCAL_MACHINE\Software\ODBC\ODBCINST.INI\ODBC Drivers]
"Microsoft Access Driver (*.mdb)"="Installed"

[HKEY_LOCAL_MACHINE\Software\ODBC\ODBCINST.INI\Microsoft Access Driver (*.mdb)]
"Driver"="C:\\windows\\system32\\odbcjt32.dll"
"Setup"="C:\\windows\\system32\\odbcjt32.dll"
"APILevel"="1"
"ConnectFunctions"="YYN"
"DriverODBCVer"="02.50"
"FileExtns"="*.mdb"
"FileUsage"="2"
"SQLLevel"="0"
"UsageCount"=dword:00000001
EOF
"$WINE" 'C:\windows\syswow64\regedit.exe' /S 'C:\access-odbc.reg'
rm -f "$REG"

[ -f "$WINEPREFIX/drive_c/windows/syswow64/msjet40.dll" ] || { echo "msjet40.dll saknas, något gick fel" >&2; exit 1; }
echo "Klart. Starta med: WINEPREFIX=$WINEPREFIX $WINE wHHEK_x86.exe"
