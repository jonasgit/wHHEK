<!--  //-*- coding: utf-8 -*-  -->

# wHHEK
Personal finance package. Aimed at Swedish users.

I mitten av 90-talet fanns ett program som hette Hogia Hemekonomi. Det sparar data i en mdb-fil. Det är ett program för att hålla lite ordning på sin ekonomi, glorifierad kassabok och budget ungefär.

Hogia Hemekonomi är ett 16-bitars windowsprogram som visserligen fungerar i Windows 10 om man kör 32-bitars varianten och slår på 16-bitars funktionen. Men det lär inte hålla i evighet (hösten 2025 är prognosen). I Windows 11 krävs https://github.com/otya128/winevdm
Så detta programmet är tänkt att vara en ersättare.

Detta programmet fungerar med både med mdb-filer och sqlite3-filer. Det finns en konverterare till sqlite3 på https://github.com/jonasgit/hhek2sqlite

Gränssnittet är webbaserat, dvs programmet körs i bakgrunden och användare interagerar via webläsaren. Allt körs lokalt på egna maskinen.

Detta programmet går att köra i Windows så länge det går att köra 32-bitars program dvs Windows 10 64-bitars är ok och även Windows 11 fungerar.

Konverterar man datafilen till sqlite så bör det gå att köra även på Mac, i Linux och kravet på 32-bitars program i Windows försvinner, men inte testat.

Windows-versionen av det här går att köra i Linux också med hjälp av Wine. Men då krävs:
`winetricks mdac28 jet40`

**Projektmål:**
1. Använd samma filformat som Hogia Hemekonomi (MDB/Jet/Access). För att man ska kunna växla fritt mellan detta program och Hemekonomi. Endast på Windows.
2. Och ett mer framtidssäkert filformat (Sqlite3) som fungerar på andra OS. Det ska gå att konvertera mellan båda.
3. Motsvarande funktioner med att hantera transaktioner, budget, m.m.
4. Ytterligare funktioner som fungerar inom filformatets begränsningar.

**Status:** Mycket grundfunktioner finns. Windows-varianten är fokus tills vidare.

**Installation:** Bygg enligt kommentarer i filen make.bat. Eller se om det finns en färdig
exe-fil under Releases/version/Assets/.

**Kör:** Starta programmet i katalogen som har en mdb-fil. Eller där du vill skapa en ny databas. Standard webläsare öppnas automatiskt med startsidan för programmet.
