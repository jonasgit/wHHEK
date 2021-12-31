//-*- coding: utf-8 -*-

# wHHEK
Personal finance package. Aimed at Swedish users.

I mitten av 90-talet fanns ett program som hette Hogia Hemekonomi. Det sparar data i en mdb-fil. Det är ett program för att hålla lite ordning på sin ekonomi, glorifierad kassabok ungefär.

Hogia Hemekonomi är ett 16-bitars windowsprogram som visserligen fungerar i Windows 10 om man kör 32-bitars varianten och slår på 16-bitars funktionen. Men det lär inte hålla i evighet. Så detta programmet är tänkt att blir en ersättare.

Detta programmet fungerar med både med mdb-filer och sqlite3. Det finns en konverterare till sqlite3 på https://github.com/jonasgit/hhek2sqlite

Gränssnittet är webbaserat, dvs programmet körs i bakgrunden och användare interagerar via webläsaren.

Detta programmet går att köra i Windows så länge det går att köra 32-bitars program dvs Windows 10 64-bitars är ok och även Windows 11 fungerar.

Konverterar man till sqlite så bör det gå att köra även på Mac, i Linux och kravet på 32-bitars program i Windows
försvinner, men inte testat.

**Status:** "i sin linda".

**Installation:** Bygg enligt kommentarer i början på filen main.go eller filen make.bat. Eller se om det finns en färdig
exe-fil under Releases.

**Kör:** Starta programmet i katalogen som har en mdb-fil.

*Under 2022 sponsras detta projektet av JetBrains med en Open Source Development licens för GoLand.*
