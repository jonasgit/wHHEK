//-*- coding: utf-8 -*-

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	
	_ "github.com/alexbrainman/odbc"    // BSD-3-Clause License 
	_ "github.com/mattn/go-sqlite3"     // MIT License
)

func printPersoner(w http.ResponseWriter, db *sql.DB) {
	res, err := db.Query("SELECT Namn,Född,Kön,Löpnr FROM Personer")

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	var namn []byte   // size 50
	var birth string  // size 4 (år, 0 för Gemensamt)
	var sex string    // size 10 (text: Gemensamt, Man, Kvinna)
	var nummer int    // autoinc Primary Key, index

	fmt.Fprintf(w, "<table style=\"width:100%%\"><tr><th>Namn</th><th>Födelsedag</th><th>Kön</th><th>Redigera</th><th>Radera</th>\n")
	for res.Next() {
		err = res.Scan(&namn, &birth, &sex, &nummer)

		fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td>", toUtf8(namn), birth, sex)
		fmt.Fprintf(w, "<td><form method=\"POST\" action=\"/personer\"><input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"editform\"><input type=\"submit\" value=\"Redigera\"></form></td>\n", nummer)
		fmt.Fprintf(w, "<td><form method=\"POST\" action=\"/personer\"><input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"radera\"><input type=\"checkbox\" id=\"OK\" name=\"OK\" required><label for=\"OK\">OK</label><input type=\"submit\" value=\"Radera\"></form></td></tr>\n", nummer)
	}
	fmt.Fprintf(w, "</table>\n")

	fmt.Fprintf(w, "<form method=\"POST\" action=\"/personer\"><input type=\"hidden\" id=\"action\" name=\"action\" value=\"addform\"><input type=\"submit\" value=\"Ny person\"></form>\n")
}

func printPersonerFooter(w http.ResponseWriter, db *sql.DB) {
	fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	fmt.Fprintf(w, "</body>\n")
	fmt.Fprintf(w, "</html>\n")
}

func raderaPerson(w http.ResponseWriter, lopnr int, db *sql.DB) {
	fmt.Println("raderaPerson lopnr: ", lopnr)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
`DELETE FROM Personer WHERE (Löpnr=?)`, lopnr)

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}
	fmt.Fprintf(w, "Person med löpnr %d raderad.<br>", lopnr);
}

func editformPerson(w http.ResponseWriter, lopnr int, db *sql.DB) {
	fmt.Println("editformPerson lopnr: ", lopnr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res1 := db.QueryRowContext(ctx,
`SELECT Namn,Född,Kön FROM Personer WHERE (Löpnr=?)`, lopnr)

	var namn []byte   // size 50
	var birth string  // size 4 (år, 0 för Gemensamt)
	var sex string    // size 10 (text: Gemensamt, Man, Kvinna)

	err := res1.Scan(&namn, &birth, &sex)
	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	fmt.Fprintf(w, "Redigera person<br>")
	fmt.Fprintf(w, "<form method=\"POST\" action=\"/personer\">")

	fmt.Fprintf(w,"<label for=\"namn\">Namn:</label>")
	fmt.Fprintf(w,"<input type=\"text\" id=\"namn\" name=\"namn\" value=\"%s\">", toUtf8(namn))
	fmt.Fprintf(w,"<label for=\"birth\">Födelsedag:</label>")
	fmt.Fprintf(w,"<input type=\"text\" id=\"birth\" name=\"birth\" value=\"%s\">", birth)
	fmt.Fprintf(w,"<label for=\"sex\">Kön:</label>")
	fmt.Fprintf(w,"<input type=\"text\" id=\"sex\" name=\"sex\" value=\"%s\">", sex)
	
	fmt.Fprintf(w, "<input type=\"hidden\" id=\"lopnr\" name=\"lopnr\" value=\"%d\">", lopnr)
	fmt.Fprintf(w, "<input type=\"hidden\" id=\"action\" name=\"action\" value=\"update\">")
	fmt.Fprintf(w, "<input type=\"submit\" value=\"Uppdatera\">")
	fmt.Fprintf(w, "</form>\n")
	fmt.Fprintf(w, "<p>\n");
}

func addformPerson(w http.ResponseWriter, db *sql.DB) {
	fmt.Println("addformPerson ")

	fmt.Fprintf(w, "Lägg till person<br>")
	fmt.Fprintf(w, "<form method=\"POST\" action=\"/personer\">")

	fmt.Fprintf(w,"<label for=\"namn\">Namn:</label>")
	fmt.Fprintf(w,"<input type=\"text\" id=\"namn\" name=\"namn\" value=\"%s\">", "")
	fmt.Fprintf(w,"<label for=\"birth\">Födelseår:</label>")
	fmt.Fprintf(w,"<input type=\"text\" id=\"birth\" name=\"birth\" value=\"%s\">", "")
	fmt.Fprintf(w,"  <label for=\"sex\">Kön:</label>")
	fmt.Fprintf(w,"  <select name=\"sex\" id=\"sex\" required>")
	fmt.Fprintf(w,"    <option value=\"%s\">%s</option>", "Kvinna", "Kvinna")
	fmt.Fprintf(w,"    <option value=\"%s\">%s</option>", "Man", "Man")
	fmt.Fprintf(w,"  </select>\n")

	
	fmt.Fprintf(w, "<input type=\"hidden\" id=\"action\" name=\"action\" value=\"add\">")
	fmt.Fprintf(w, "<input type=\"submit\" value=\"Ny person\">")
	fmt.Fprintf(w, "</form>\n")
	fmt.Fprintf(w, "<p>\n");
}

func addPerson(w http.ResponseWriter, namn string, birth string, sex string, db *sql.DB) {
	fmt.Println("addPerson namn: ", namn)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
`INSERT INTO Personer(Namn, Född, Kön) VALUES (?, ?, ?)`, namn, birth, sex)

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}
	fmt.Fprintf(w, "Person %s tillagd.<br>", namn);
}


func updatePerson(w http.ResponseWriter, lopnr int, namn string, birth string, sex string, db *sql.DB) {
	fmt.Println("updatePerson lopnr: ", lopnr)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
`UPDATE Personer SET Namn = ?, Född = ?, Kön = ? WHERE (Löpnr=?)`, namn, birth, sex, lopnr)

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}
	fmt.Fprintf(w, "Person %s uppdaterad.<br>", namn);
}

func hanterapersoner(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "<html>\n")
	fmt.Fprintf(w, "<head>\n")
	fmt.Fprintf(w, "<style>\n")
	fmt.Fprintf(w, "table,th,td { border: 1px solid black }\n")
	fmt.Fprintf(w, "</style>\n")
	fmt.Fprintf(w, "</head>\n")
	fmt.Fprintf(w, "<body>\n")

	fmt.Fprintf(w, "<h1>%s</h1>\n", currentDatabase)
	fmt.Fprintf(w, "<h2>Personer</h2>\n")

	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	formaction := req.FormValue("action")
	var lopnr int=-1;
	if len(req.FormValue("lopnr"))>0 {
		lopnr,err = strconv.Atoi(req.FormValue("lopnr"))
	}

	switch formaction {
	case "radera" : raderaPerson(w, lopnr, db)
	case "addform" : addformPerson(w, db)
	case "add" :
		var namn string="";
		if len(req.FormValue("namn"))>0 {
			namn = req.FormValue("namn")
		}
		var birth string="";
		if len(req.FormValue("birth"))>0 {
			birth = req.FormValue("birth")
		}
		var sex string="";
		if len(req.FormValue("sex"))>0 {
			sex = req.FormValue("sex")
		}
		addPerson(w, namn, birth, sex, db)
	case "editform" : editformPerson(w, lopnr, db)
	case "update" :
		var namn string="";
		if len(req.FormValue("namn"))>0 {
			namn = req.FormValue("namn")
		}
		var birth string="";
		if len(req.FormValue("birth"))>0 {
			birth = req.FormValue("birth")
		}
		var sex string="";
		if len(req.FormValue("sex"))>0 {
			sex = req.FormValue("sex")
		}
		updatePerson(w, lopnr, namn, birth, sex, db)
	default :
		fmt.Println("Okänd action: %s\n", formaction)
	}
	printPersoner(w, db)
	printPersonerFooter(w, db)
}

func getPersonNames() []string {
	names := make([]string, 0)

	res, err := db.Query("SELECT Namn FROM Personer ORDER BY Namn")

	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}

	var Namn  []byte  // size 50, index
	for res.Next() {
		err = res.Scan(&Namn)
		names = append(names, toUtf8(Namn))
	}
	return names
}
