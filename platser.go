//-*- coding: utf-8 -*-

package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type Plats struct {
	Lopnr      int    // autoinc Primary Key, index
	Namn       string // size 40
	Gironummer string // size 20
	Typ        bool   // size 2
	RefKonto   string // size 40
}

type PlatserTemplateData struct {
	DatabaseName string
	Message      string
	Platser      []Plats
	Kontonamn    []string
	ShowEditForm bool
	ShowAddForm  bool
	EditPlats    Plats
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
	retplats.Lopnr = lopnr
	retplats.Namn = toUtf8(Namn)
	retplats.Gironummer = toUtf8(Gironummer)
	retplats.Typ = toUtf8(Typ) == "true"
	retplats.RefKonto = toUtf8(RefKonto)

	return retplats
}

func getPlatser(db *sql.DB) []Plats {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rows, err := db.QueryContext(ctx, "SELECT Löpnr, Namn, Gironummer, Typ, RefKonto FROM Platser")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var platser []Plats
	for rows.Next() {
		var p Plats
		var Namn []byte
		var Gironummer []byte
		var Typ []byte
		var RefKonto []byte
		err := rows.Scan(&p.Lopnr, &Namn, &Gironummer, &Typ, &RefKonto)
		if err != nil {
			log.Fatal(err)
		}
		p.Namn = toUtf8(Namn)
		p.Gironummer = toUtf8(Gironummer)
		p.Typ = toUtf8(Typ) == "true"
		p.RefKonto = toUtf8(RefKonto)
		platser = append(platser, p)
	}
	return platser
}

func raderaPlats(db *sql.DB, lopnr int) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`DELETE FROM Platser WHERE (Löpnr=?)`, lopnr)
	return err
}

func skapaPlats(db *sql.DB, namn string, gironum string, acctype bool, refacc string) error {
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
	return err
}

func updatePlats(db *sql.DB, lopnr int, namn string, gironum string, acctype string, refacc string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`UPDATE Platser SET Namn = ?, Gironummer = ?, Typ = ?, RefKonto = ? WHERE (Löpnr=?)`, namn, gironum, acctype, refacc, lopnr)
	return err
}

func hanteraplatser(w http.ResponseWriter, req *http.Request) {
	// First try to load from file system (for development)
	tmpl, err := template.ParseFiles("templates/platser.html")
	if err != nil {
		// Fall back to embedded template
		tmpl, err = template.New("platser.html").ParseFS(htmlTemplates, "html/platser.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	data := PlatserTemplateData{
		DatabaseName: currentDatabase,
		Platser:      getPlatser(db),
		Kontonamn:    getAccNames(),
	}

	err = req.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	formaction := req.FormValue("action")
	var lopnr = -1
	if len(req.FormValue("lopnr")) > 0 {
		lopnr, _ = strconv.Atoi(req.FormValue("lopnr"))
	}

	switch formaction {
	case "radera":
		if err := raderaPlats(db, lopnr); err != nil {
			data.Message = fmt.Sprintf("Fel vid radering av plats: %v", err)
		} else {
			data.Message = fmt.Sprintf("Plats med löpnr %d raderad.", lopnr)
		}
		data.Platser = getPlatser(db)
	case "addform":
		data.ShowAddForm = true
	case "add":
		namn := req.FormValue("namn")
		gironum := req.FormValue("gironum")
		acctype := req.FormValue("kontokort") == "on"
		refacc := req.FormValue("refacc")

		if err := skapaPlats(db, namn, gironum, acctype, refacc); err != nil {
			data.Message = fmt.Sprintf("Fel vid tillägg av plats: %v", err)
		} else {
			data.Message = fmt.Sprintf("Plats %s tillagd.", namn)
		}
		data.Platser = getPlatser(db)
	case "editform":
		data.ShowEditForm = true
		data.EditPlats = hämtaPlats(db, lopnr)
	case "update":
		namn := req.FormValue("namn")
		gironum := req.FormValue("gironum")
		acctype := req.FormValue("type")
		refacc := req.FormValue("refacc")

		if err := updatePlats(db, lopnr, namn, gironum, acctype, refacc); err != nil {
			data.Message = fmt.Sprintf("Fel vid uppdatering av plats: %v", err)
		} else {
			data.Message = fmt.Sprintf("Plats %s uppdaterad.", namn)
		}
		data.Platser = getPlatser(db)
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Hämtar alla platser, både från tabellen Platser och från tabellen Transaktioner
func getPlaceNames() []string {
	names := make([]string, 0)

	res, err := db.Query("select DISTINCT Namn from Platser union select DISTINCT TillKonto from Transaktioner where TillKonto <> '---' and (Typ = 'Inköp' or Typ = 'Fast Utgift') order by Namn")

	if err != nil {
		log.Fatal(err)
	}

	var Namn []byte
	for res.Next() {
		_ = res.Scan(&Namn)
		names = append(names, toUtf8(Namn))
	}
	res.Close()
	return names
}
