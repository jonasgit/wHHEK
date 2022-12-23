//-*- coding: utf-8 -*-

# wHHEK
Personal finance package. Aimed at Swedish users.

I mitten av 90-talet fanns ett program som hette Hogia Hemekonomi. Det sparar data i en mdb-fil. Det är ett program för att hålla lite ordning på sin ekonomi, glorifierad kassabok och budget ungefär.

Hogia Hemekonomi är ett 16-bitars windowsprogram som visserligen fungerar i Windows 10 om man kör 32-bitars varianten och slår på 16-bitars funktionen. Men det lär inte hålla i evighet (ca 2025 är prognosen). Så detta programmet är tänkt att blir en ersättare.

Detta programmet fungerar med både med mdb-filer och sqlite3. Det finns en konverterare till sqlite3 på https://github.com/jonasgit/hhek2sqlite

Gränssnittet är webbaserat, dvs programmet körs i bakgrunden och användare interagerar via webläsaren. Allt körs lokalt på egna maskinen.

Detta programmet går att köra i Windows så länge det går att köra 32-bitars program dvs Windows 10 64-bitars är ok och även Windows 11 fungerar.

Konverterar man till sqlite så bör det gå att köra även på Mac, i Linux och kravet på 32-bitars program i Windows försvinner, men inte testat.

Windows-versionen av det här går att köra i Linux också med hjälp av Wine. Men då krävs:
`winetricks mdac28 jet40`

**Status:** "i sin linda". Lite grundfunktioner med att lägga in transaktioner och visa finns.

**Installation:** Bygg enligt kommentarer i filen make.bat. Eller se om det finns en färdig
exe-fil under Releases/version/Assets/.

**Kör:** Starta programmet i katalogen som har en mdb-fil. Eller där du vill skapa en ny databas.

*Under 2022-2023 sponsras detta projektet av JetBrains med en Open Source Development licens för GoLand.*
