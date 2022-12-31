//-*- coding: utf-8 -*-

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type Plats struct {
	Namn       string // size
	Gironummer string // size
	Typ        bool   // Oanvänt?
	RefKonto   string // size, != 0 betyder kontokortsföretag
}

func antalPlatser(db *sql.DB) int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var antal int

	err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM Platser`).Scan(&antal)
	if err != nil {
		log.Fatal(err)
	}

	return antal
}

func hämtaPlats(db *sql.DB, lopnr int) Plats {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var Namn []byte       // size 40
	var Gironummer []byte // size 20
	var Typ []byte        // size 2
	var RefKonto []byte   // size 40

	err := db.QueryRowContext(ctx,
		`SELECT Namn,Gironummer,Typ,RefKonto FROM Platser WHERE (Löpnr=?)`, lopnr).Scan(&Namn, &Gironummer, &Typ, &RefKonto)
	if err != nil {
		log.Fatal(err)
	}

	var retplats Plats

	retplats.Namn = toUtf8(Namn)
	retplats.Gironummer = toUtf8(Gironummer)
	if toUtf8(Typ) == "true" {
		retplats.Typ = true
	} else {
		retplats.Typ = false
	}
	retplats.RefKonto = toUtf8(RefKonto)

	return retplats
}

func printPlatser(w http.ResponseWriter, db *sql.DB) {
	res, err := db.Query("SELECT Namn,Gironummer,Typ,RefKonto,Löpnr FROM Platser")

	if err != nil {
		log.Fatal(err)
	}

	var Namn []byte       // size 40
	var Gironummer []byte // size 20
	var Typ []byte        // size 2
	var RefKonto []byte   // size 40
	var Löpnr []byte      // autoinc Primary Key, index

	_, _ = fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Namn</th><th>Gironummer</th><th>Typ</th><th>RefKonto</th><th>Redigera</th><th>Radera</th>\n")
	for res.Next() {
		err = res.Scan(&Namn, &Gironummer, &Typ, &RefKonto, &Löpnr)

		_, _ = fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td>", toUtf8(Namn), toUtf8(Gironummer), toUtf8(Typ), toUtf8(RefKonto))
		_, _ = fmt.Fprintf(w, "<td><form method=\"POST\" action=\"/platser\"><input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%s\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"editform\"><input type=\"submit\" value=\"Redigera\"></form></td>\n", Löpnr)
		_, _ = fmt.Fprintf(w, "<td><form method=\"POST\" action=\"/platser\"><input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%s\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"radera\"><input type=\"checkbox\" id=\"OK\" name=\"OK\" required><label for=\"OK\">OK</label><input type=\"submit\" value=\"Radera\"></form></td></tr>\n", Löpnr)
	}
	_, _ = fmt.Fprintf(w, "</table>\n")

	_, _ = fmt.Fprintf(w, "<form method=\"POST\" action=\"/platser\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"addform\"><input type=\"submit\" value=\"Ny plats\"></form>\n")
}

func printPlatserFooter(w http.ResponseWriter) {
	_, _ = fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	_, _ = fmt.Fprintf(w, "</body>\n")
	_, _ = fmt.Fprintf(w, "</html>\n")
}

func raderaPlats(w http.ResponseWriter, lopnr int, db *sql.DB) {
	fmt.Println("raderaPlats lopnr: ", lopnr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`DELETE FROM Platser WHERE (Löpnr=?)`, lopnr)

	if err != nil {
		log.Fatal(err)
	}
	_, _ = fmt.Fprintf(w, "Plats med löpnr %d raderad.<br>", lopnr)
}

func editformPlats(w http.ResponseWriter, lopnr int, db *sql.DB) {
	fmt.Println("editformPlats lopnr: ", lopnr)

	kontonamn := getAccNames()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var Namn []byte       // size 40
	var Gironummer []byte // size 20
	var Typ []byte        // size 2
	var RefKonto []byte   // size 40
	err := db.QueryRowContext(ctx,
		`SELECT Namn,Gironummer,Typ,RefKonto FROM Platser WHERE (Löpnr=?)`, lopnr).Scan(&Namn, &Gironummer, &Typ, &RefKonto)
	if err != nil {
		log.Fatal(err)
	}

	_, _ = fmt.Fprintf(w, "Redigera plats<br>")
	_, _ = fmt.Fprintf(w, "<form method=\"POST\" action=\"/platser\">")

	_, _ = fmt.Fprintf(w, "<label for=\"namn\">Namn:</label>")
	_, _ = fmt.Fprintf(w, "<input type=\"text\" id=\"namn\" name=\"namn\" value=\"%s\">", toUtf8(Namn))
	_, _ = fmt.Fprintf(w, "<label for=\"gironum\">Gironummer:</label>")
	_, _ = fmt.Fprintf(w, "<input type=\"text\" id=\"gironum\" name=\"gironum\" value=\"%s\">", toUtf8(Gironummer))
	_, _ = fmt.Fprintf(w, "<label for=\"type\">Typ:</label>")
	_, _ = fmt.Fprintf(w, "<input type=\"text\" id=\"type\" name=\"type\" value=\"%s\">", toUtf8(Typ))
	_, _ = fmt.Fprintf(w, "<label for=\"refacc\">RefKonto:</label>")
	_, _ = fmt.Fprintf(w, "<select id=\"refacc\" name=\"refacc\">")
	for _, s := range kontonamn {
		var selected = ""
		if s == toUtf8(RefKonto) {
			selected = "selected"
		}
		_, _ = fmt.Fprintf(w, "    <option value=\"%s\" %s>%s</option>", s, selected, s)
	}

	_, _ = fmt.Fprintf(w, "</select>\n")

	_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\">", lopnr)
	_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"action\" name=\"action\" value=\"update\">")
	_, _ = fmt.Fprintf(w, "<input type=\"submit\" value=\"Uppdatera\">")
	_, _ = fmt.Fprintf(w, "</form>\n")
	_, _ = fmt.Fprintf(w, "<p>\n")
}

func addformPlats(w http.ResponseWriter) {
	fmt.Println("addformPlats ")

	kontonamn := getAccNames()

	_, _ = fmt.Fprintf(w, "Lägg till plats<br>")
	_, _ = fmt.Fprintf(w, "<form method=\"POST\" action=\"/platser\">")

	_, _ = fmt.Fprintf(w, "<label for=\"namn\">Namn:</label>")
	_, _ = fmt.Fprintf(w, "<input type=\"text\" id=\"namn\" name=\"namn\" value=\"%s\">", "")
	_, _ = fmt.Fprintf(w, "<label for=\"gironum\">Gironummer:</label>")
	_, _ = fmt.Fprintf(w, "<input type=\"text\" id=\"gironum\" name=\"gironum\" value=\"%s\">", "")
	_, _ = fmt.Fprintf(w, "<label for=\"kontokort\">Kontokortsföretag:</label>")
	_, _ = fmt.Fprintf(w, "<input type=\"checkbox\" id=\"kontokort\" name=\"kontokort\">")
	_, _ = fmt.Fprintf(w, "<label for=\"refacc\">RefKonto:</label>")
	_, _ = fmt.Fprintf(w, "<select id=\"refacc\" name=\"refacc\">")
	for _, s := range kontonamn {
		var selected = ""
		_, _ = fmt.Fprintf(w, "    <option value=\"%s\" %s>%s</option>", s, selected, s)
	}

	_, _ = fmt.Fprintf(w, "</select>\n")

	_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"action\" name=\"action\" value=\"add\">")
	_, _ = fmt.Fprintf(w, "<input type=\"submit\" value=\"Ny plats\">")
	_, _ = fmt.Fprintf(w, "</form>\n")
	_, _ = fmt.Fprintf(w, "<p>\n")
}

func skapaPlats(db *sql.DB, namn string, gironum string, acctype bool, refacc string) error {
	if db == nil {
		log.Fatal("skapaPlats anropad med db=nil")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if len(gironum) < 1 {
		gironum = " "
	}

	if !acctype {
		refacc = " "
	}

	_, err := db.ExecContext(ctx,
		`INSERT INTO Platser (Namn,Gironummer,Typ,RefKonto) VALUES (?, ?, ?, ?)`, namn, gironum, "", refacc)

	if err != nil {
		log.Fatal(err)
	}

	return err
}

func addPlats(w http.ResponseWriter, namn string, gironum string, acctype bool, refacc string, db *sql.DB) {
	fmt.Println("addPlats namn: ", namn)

	_ = skapaPlats(db, namn, gironum, acctype, refacc)

	_, _ = fmt.Fprintf(w, "Plats %s tillagd.<br>", namn)
}

func updatePlats(w http.ResponseWriter, lopnr int, namn string, gironum string, acctype string, refacc string, db *sql.DB) {
	fmt.Println("updatePlats lopnr: ", lopnr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`UPDATE Platser SET Namn = ?, Gironummer = ?, Typ = ?, RefKonto = ? WHERE (Löpnr=?)`, namn, gironum, acctype, refacc, lopnr)

	if err != nil {
		log.Fatal(err)
	}
	_, _ = fmt.Fprintf(w, "Plats %s uppdaterad.<br>", namn)
}

func hanteraplatser(w http.ResponseWriter, req *http.Request) {
	_, _ = fmt.Fprintf(w, "<html>\n")
	_, _ = fmt.Fprintf(w, "<head>\n")
	_, _ = fmt.Fprintf(w, "<style>\n")
	_, _ = fmt.Fprintf(w, "table,th,td { border: 1px solid black }\n")
	_, _ = fmt.Fprintf(w, "</style>\n")
	_, _ = fmt.Fprintf(w, "</head>\n")
	_, _ = fmt.Fprintf(w, "<body>\n")

	_, _ = fmt.Fprintf(w, "<h1>%s</h1>\n", currentDatabase)
	_, _ = fmt.Fprintf(w, "<h2>Platser</h2>\n")

	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	formaction := req.FormValue("action")
	var lopnr = -1
	if len(req.FormValue("lopnr")) > 0 {
		lopnr, err = strconv.Atoi(req.FormValue("lopnr"))
	}

	switch formaction {
	case "radera":
		raderaPlats(w, lopnr, db)
	case "addform":
		addformPlats(w)
	case "add":
		var namn = ""
		if len(req.FormValue("namn")) > 0 {
			namn = req.FormValue("namn")
		}
		var gironum = ""
		if len(req.FormValue("gironum")) > 0 {
			gironum = req.FormValue("gironum")
		}
		var acctype = false
		fmt.Println("FormValue type: ", req.FormValue("kontokort"))
		if req.FormValue("kontokort") == "on" {
			acctype = true
		}
		var refacc = ""
		if len(req.FormValue("refacc")) > 0 {
			refacc = req.FormValue("refacc")
		}
		addPlats(w, namn, gironum, acctype, refacc, db)
	case "editform":
		editformPlats(w, lopnr, db)
	case "update":
		var namn = ""
		if len(req.FormValue("namn")) > 0 {
			namn = req.FormValue("namn")
		}
		var gironum = ""
		if len(req.FormValue("gironum")) > 0 {
			gironum = req.FormValue("gironum")
		}
		var acctype = ""
		if len(req.FormValue("type")) > 0 {
			acctype = req.FormValue("type")
		}
		var refacc = ""
		if len(req.FormValue("refacc")) > 0 {
			refacc = req.FormValue("refacc")
		}
		updatePlats(w, lopnr, namn, gironum, acctype, refacc, db)
	default:
		fmt.Println("Okänd action: ", formaction)
	}
	printPlatser(w, db)
	printPlatserFooter(w)
}

func getPlaceNames() []string {
	names := make([]string, 0)

	res, err := db.Query("SELECT Namn FROM Platser ORDER BY Namn")

	if err != nil {
		log.Fatal(err)
	}

	var Namn []byte // size 40, index
	for res.Next() {
		err = res.Scan(&Namn)
		names = append(names, toUtf8(Namn))
	}
	return names
}
