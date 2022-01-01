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

type person struct {
	namn  string
	birth int
	sex   string
}

func printPersoner(w http.ResponseWriter, db *sql.DB) {
	res, err := db.Query("SELECT Namn,Född,Kön,Löpnr FROM Personer")

	if err != nil {
		log.Fatal(err)
	}

	var namn []byte  // size 50
	var birth string // size 4 (år, 0 för Gemensamt)
	var sex string   // size 10 (text: Gemensamt, Man, Kvinna)
	var nummer int   // autoinc Primary Key, index

	_, _ = fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Namn</th><th>Födelsedag</th><th>Kön</th><th>Redigera</th><th>Radera</th>\n")
	for res.Next() {
		err = res.Scan(&namn, &birth, &sex, &nummer)

		_, _ = fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td>", toUtf8(namn), birth, sex)
		_, _ = fmt.Fprintf(w, "<td><form method=\"POST\" action=\"/personer\"><input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"editform\"><input type=\"submit\" value=\"Redigera\"></form></td>\n", nummer)
		_, _ = fmt.Fprintf(w, "<td><form method=\"POST\" action=\"/personer\"><input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"radera\"><input type=\"checkbox\" id=\"OK\" name=\"OK\" required><label for=\"OK\">OK</label><input type=\"submit\" value=\"Radera\"></form></td></tr>\n", nummer)
	}
	_, _ = fmt.Fprintf(w, "</table>\n")

	_, _ = fmt.Fprintf(w, "<form method=\"POST\" action=\"/personer\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"addform\"><input type=\"submit\" value=\"Ny person\"></form>\n")
}

func printPersonerFooter(w http.ResponseWriter) {
	_, _ = fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	_, _ = fmt.Fprintf(w, "</body>\n")
	_, _ = fmt.Fprintf(w, "</html>\n")
}

func raderaPerson(w http.ResponseWriter, lopnr int, db *sql.DB) {
	fmt.Println("raderaPerson lopnr: ", lopnr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`DELETE FROM Personer WHERE (Löpnr=?)`, lopnr)

	if err != nil {
		log.Fatal(err)
	}
	_, _ = fmt.Fprintf(w, "Person med löpnr %d raderad.<br>", lopnr)
}

func editformPerson(w http.ResponseWriter, lopnr int, db *sql.DB) {
	fmt.Println("editformPerson lopnr: ", lopnr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res1 := db.QueryRowContext(ctx,
		`SELECT Namn,Född,Kön FROM Personer WHERE (Löpnr=?)`, lopnr)

	var namn []byte  // size 50
	var birth string // size 4 (år, 0 för Gemensamt)
	var sex string   // size 10 (text: Gemensamt, Man, Kvinna)

	err := res1.Scan(&namn, &birth, &sex)
	if err != nil {
		log.Fatal(err)
	}

	_, _ = fmt.Fprintf(w, "Redigera person<br>")
	_, _ = fmt.Fprintf(w, "<form method=\"POST\" action=\"/personer\">")

	_, _ = fmt.Fprintf(w, "<label for=\"namn\">Namn:</label>")
	_, _ = fmt.Fprintf(w, "<input type=\"text\" id=\"namn\" name=\"namn\" value=\"%s\">", toUtf8(namn))
	_, _ = fmt.Fprintf(w, "<label for=\"birth\">Födelsedag:</label>")
	_, _ = fmt.Fprintf(w, "<input type=\"text\" id=\"birth\" name=\"birth\" value=\"%s\">", birth)
	_, _ = fmt.Fprintf(w, "<label for=\"sex\">Kön:</label>")
	_, _ = fmt.Fprintf(w, "<input type=\"text\" id=\"sex\" name=\"sex\" value=\"%s\">", sex)

	_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\">", lopnr)
	_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"action\" name=\"action\" value=\"update\">")
	_, _ = fmt.Fprintf(w, "<input type=\"submit\" value=\"Uppdatera\">")
	_, _ = fmt.Fprintf(w, "</form>\n")
	_, _ = fmt.Fprintf(w, "<p>\n")
}

func addformPerson(w http.ResponseWriter) {
	fmt.Println("addformPerson ")

	_, _ = fmt.Fprintf(w, "Lägg till person<br>")
	_, _ = fmt.Fprintf(w, "<form method=\"POST\" action=\"/personer\">")

	_, _ = fmt.Fprintf(w, "<label for=\"namn\">Namn:</label>")
	_, _ = fmt.Fprintf(w, "<input type=\"text\" id=\"namn\" name=\"namn\" value=\"%s\">", "")
	_, _ = fmt.Fprintf(w, "<label for=\"birth\">Födelseår:</label>")
	_, _ = fmt.Fprintf(w, "<input type=\"text\" id=\"birth\" name=\"birth\" value=\"%s\">", "")
	_, _ = fmt.Fprintf(w, "  <label for=\"sex\">Kön:</label>")
	_, _ = fmt.Fprintf(w, "  <select name=\"sex\" id=\"sex\" required>")
	_, _ = fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", "Kvinna", "Kvinna")
	_, _ = fmt.Fprintf(w, "    <option value=\"%s\">%s</option>", "Man", "Man")
	_, _ = fmt.Fprintf(w, "  </select>\n")

	_, _ = fmt.Fprintf(w, "<input type=\"hidden\" id=\"action\" name=\"action\" value=\"add\">")
	_, _ = fmt.Fprintf(w, "<input type=\"submit\" value=\"Ny person\">")
	_, _ = fmt.Fprintf(w, "</form>\n")
	_, _ = fmt.Fprintf(w, "<p>\n")
}

func skapaPerson(db *sql.DB, namn string, birth int, sex string) error {
	if db == nil {
		log.Fatal("skapaPerson anropad med db=nil")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`INSERT INTO Personer(Namn, Född, Kön) VALUES (?, ?, ?)`, namn, birth, sex)

	if err != nil {
		log.Fatal(err)
	}
	return err
}

func addPerson(w http.ResponseWriter, namn string, birth string, sex string, db *sql.DB) {
	fmt.Println("addPerson namn: ", namn)

	birthint, err := strconv.Atoi(birth)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Person ej tillagd, felaktigt födelseår.<br>")
		log.Fatal(err)
	}

	err = skapaPerson(db, namn, birthint, sex)

	if err != nil {
		log.Fatal(err)
	}
	_, _ = fmt.Fprintf(w, "Person %s tillagd.<br>", namn)
}

func updatePerson(w http.ResponseWriter, lopnr int, namn string, birth string, sex string, db *sql.DB) {
	fmt.Println("updatePerson lopnr: ", lopnr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`UPDATE Personer SET Namn = ?, Född = ?, Kön = ? WHERE (Löpnr=?)`, namn, birth, sex, lopnr)

	if err != nil {
		log.Fatal(err)
	}
	_, _ = fmt.Fprintf(w, "Person %s uppdaterad.<br>", namn)
}

func hanterapersoner(w http.ResponseWriter, req *http.Request) {
	_, _ = fmt.Fprintf(w, "<html>\n")
	_, _ = fmt.Fprintf(w, "<head>\n")
	_, _ = fmt.Fprintf(w, "<style>\n")
	_, _ = fmt.Fprintf(w, "table,th,td { border: 1px solid black }\n")
	_, _ = fmt.Fprintf(w, "</style>\n")
	_, _ = fmt.Fprintf(w, "</head>\n")
	_, _ = fmt.Fprintf(w, "<body>\n")

	_, _ = fmt.Fprintf(w, "<h1>%s</h1>\n", currentDatabase)
	_, _ = fmt.Fprintf(w, "<h2>Personer</h2>\n")

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
		raderaPerson(w, lopnr, db)
	case "addform":
		addformPerson(w)
	case "add":
		var namn = ""
		if len(req.FormValue("namn")) > 0 {
			namn = req.FormValue("namn")
		}
		var birth = ""
		if len(req.FormValue("birth")) > 0 {
			birth = req.FormValue("birth")
		}
		var sex = ""
		if len(req.FormValue("sex")) > 0 {
			sex = req.FormValue("sex")
		}
		addPerson(w, namn, birth, sex, db)
	case "editform":
		editformPerson(w, lopnr, db)
	case "update":
		var namn = ""
		if len(req.FormValue("namn")) > 0 {
			namn = req.FormValue("namn")
		}
		var birth = ""
		if len(req.FormValue("birth")) > 0 {
			birth = req.FormValue("birth")
		}
		var sex = ""
		if len(req.FormValue("sex")) > 0 {
			sex = req.FormValue("sex")
		}
		updatePerson(w, lopnr, namn, birth, sex, db)
	default:
		fmt.Println("Okänd action: ", formaction)
	}
	printPersoner(w, db)
	printPersonerFooter(w)
}

func getPersonNames() []string {
	names := make([]string, 0)

	res, err := db.Query("SELECT Namn FROM Personer ORDER BY Namn")

	if err != nil {
		log.Fatal(err)
	}

	var Namn []byte // size 50, index
	for res.Next() {
		err = res.Scan(&Namn)
		names = append(names, toUtf8(Namn))
	}
	return names
}

func antalPersoner(db *sql.DB) int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res1 := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM Personer`)

	var antal int

	err := res1.Scan(&antal)
	if err != nil {
		log.Fatal(err)
	}

	return antal
}

func hämtaPerson(lopnr int) person {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res1 := db.QueryRowContext(ctx,
		`SELECT Namn,Född,Kön FROM Personer WHERE (Löpnr=?)`, lopnr)

	var namn []byte  // size 50
	var birth string // size 4 (år, 0 för Gemensamt)
	var sex string   // size 10 (text: Gemensamt, Man, Kvinna)

	err := res1.Scan(&namn, &birth, &sex)
	if err != nil {
		log.Fatal(err)
	}

	var retperson person

	retperson.namn = toUtf8(namn)
	retperson.birth, err = strconv.Atoi(birth)
	retperson.sex = sex

	return retperson
}
