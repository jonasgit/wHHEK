//-*- coding: utf-8 -*-

# wHHEK
Personal finance package. Aimed at Swedish users.

I mitten av 90-talet fanns ett program som hette Hogia Hemekonomi. Det sparar data i en mdb-fil. Det finns en konverterare till sqlite3 på https://github.com/jonasgit/hhek2sqlite

Hogia Hemekonomi är ett 16-bitars windowsprogram som visserligen fungerar i Windows 10 om man kör 32-bitars varianten och slår på 16-bitars funktionen. Men det lär inte hålla i evighet. Så detta programmet är tänkt att blir en ersättare.

Detta programmet fungerar med både med mdb-filer och sqlite3.

Detta programmet går att köra i Windows så länge det går att köra 32-bitars program dvs Windows 10 64-bitars är ok under lång tid. Konverterar man till sqlite så bör det gå att köra även på Mac, i Linux och kravet på 32-bitars program i Windows försvinner, men inte testat.

Status: "i sin linda".

Installation: Bygg enligt kommentarer i början på filen main.go

Kör: Starta programmet i katalogen som har en mdb-fil.