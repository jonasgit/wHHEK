#!/bin/bash
# Testar att MDAC, Jet 4.0 och ODBC-drivrutinen för Access fungerar i ett Wine-prefix,
# dvs det som wHHEK_x86.exe behöver för att öppna mdb-filer.
#
# Användning:
#   tools/jettest/run.sh [wine]       # wine från PATH om inget anges
#
# Utan WINEPREFIX: skapar ett nytt prefix i tools/jettest/obj/prefix, installerar
# MDAC/Jet med ../../wine-setup.sh och kör sedan JETTEST.EXE.
# Med WINEPREFIX satt: kör bara JETTEST.EXE i det prefixet (t.ex. ~/.wine-hhek).
#
# Kräver JETTEST.EXE (bygg med build.sh). Slutstatus 0 = godkänt.
cd "$(dirname "$0")"
WINE=$(command -v "${1:-wine}") || { echo "Hittar inte wine: ${1:-wine}" >&2; exit 2; }
WINE=$(realpath "$WINE")
export WINEDEBUG=${WINEDEBUG:--all}
[ -f JETTEST.EXE ] || { echo "JETTEST.EXE saknas, kör build.sh först" >&2; exit 2; }

if [ -z "$WINEPREFIX" ]; then
    export WINEPREFIX=$(pwd)/obj/prefix
    # wineserver bredvid wine (installerad Wine eller byggkatalog)
    for s in "$(dirname "$WINE")/wineserver" "$(dirname "$WINE")/server/wineserver"; do
        [ -x "$s" ] && export WINESERVER=$s && break
    done
    "${WINESERVER:-wineserver}" -k 2>/dev/null
    rm -rf "$WINEPREFIX"
    mkdir -p "$WINEPREFIX"
    echo "==================== wine-setup.sh i $WINEPREFIX"
    WINE="$WINE" ../../wine-setup.sh > obj/wine-setup.log 2>&1
    rc=$?
    grep -E '^==|^Klart|Saknas|misslyckades|Fel |saknas' obj/wine-setup.log | sed 's/^/   /'
    echo "   wine-setup.sh slutstatus: $rc (hela loggen: obj/wine-setup.log)"
    "${WINESERVER:-wineserver}" -w
    echo
fi

echo "==================== JETTEST.EXE i $WINEPREFIX"
out=$("$WINE" JETTEST.EXE /q 2>&1)
printf '%s\n' "$out" | grep -v 'libEGL warning' | grep -v '^$' | tr -d '\r'
case $out in
    *"Result: PASS"*) echo "==> GODKÄNT"; exit 0 ;;
    *"Result: FAIL"*) echo "==> UNDERKÄNT"; exit 1 ;;
    *) echo "==> FEL: inget resultat från JETTEST.EXE"; exit 2 ;;
esac
