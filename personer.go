//-*- coding: utf-8 -*-

package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

// PersonRow holds the data for a single person record displayed in the table.
type PersonRow struct {
	Namn  string
	Birth string
	Sex   string
	Lopnr int
}

type PageData struct {
	People []PersonRow
}

// personerTemplates innehåller alla mallar för personsidan.
var personerTemplates = template.Must(template.ParseFS(htmlTemplates,
	"html/personer.html", "html/person_list.html", "html/personer_footer.html"))

// renderPersoner skriver ut mallen name från personerTemplates.
func renderPersoner(w http.ResponseWriter, name string, data any) {
	if err := personerTemplates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("Error executing template %s: %v", name, err)
	}
}

// personerMsg skriver ut ett meddelande med radbrytning efter.
func personerMsg(w http.ResponseWriter, format string, args ...any) {
	renderPersoner(w, "personer_msg", fmt.Sprintf(format, args...))
}

type person struct {
	namn  string
	birth int
	sex   string
}

func printPersoner(w http.ResponseWriter, db *sql.DB) {
	// 1. Fetch all data into memory structure
	rows := make([]PersonRow, 0)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res, err := db.QueryContext(ctx, "SELECT Namn,Född,Kön,Löpnr FROM Personer")
	if err != nil {
		log.Printf("Error querying person data: %v", err)
		// Instead of log.Fatal, write an error message to the user and proceed with empty table structure if possible.
		http.Error(w, "Could not load person records.", http.StatusInternalServerError)
		return
	}
	defer res.Close()

	var namn []byte  // size 50
	var birth string // size 4 (år, 0 för Gemensamt)
	var sex string   // size 10 (text: Gemensamt, Man, Kvinna)
	var nummer int   // autoinc Primary Key, index

	for res.Next() {
		if err := res.Scan(&namn, &birth, &sex, &nummer); err != nil {
			log.Printf("Error scanning person row: %v", err)
			continue // Skip this row and continue processing others
		}
		rows = append(rows, PersonRow{
			Namn:  toUtf8(namn),
			Birth: birth,
			Sex:   sex,
			Lopnr: nummer,
		})
	}

	if err := res.Err(); err != nil {
		log.Printf("Error iterating person rows: %v", err)
	}

	data := PageData{People: rows}

	// 2. Execute the template
	var buf bytes.Buffer
	if err := personerTemplates.ExecuteTemplate(&buf, "person_list", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error during rendering.", http.StatusInternalServerError)
		return
	}

	// Write the generated table content to the response writer
	_, _ = w.Write(buf.Bytes())
}

func printPersonerFooter(w http.ResponseWriter) {
	renderPersoner(w, "personer_footer", nil)
}

func raderaPerson(w http.ResponseWriter, lopnr int, db *sql.DB) {
	fmt.Println("raderaPerson lopnr: ", lopnr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`DELETE FROM Personer WHERE (Löpnr=?)`, lopnr)

	if err != nil {
		// Use http.Error instead of log.Fatal for API handlers
		http.Error(w, fmt.Sprintf("Could not delete person: %v", err), http.StatusInternalServerError)
		return
	}
	personerMsg(w, "Person med löpnr %d raderad.", lopnr)
}

func editformPerson(w http.ResponseWriter, lopnr int, db *sql.DB) {
	fmt.Println("editformPerson lopnr: ", lopnr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var namn []byte  // size 50
	var birth string // size 4 (år, 0 för Gemensamt)
	var sex string   // size 10 (text: Gemensamt, Man, Kvinna)

	err := db.QueryRowContext(ctx,
		`SELECT Namn,Född,Kön FROM Personer WHERE (Löpnr=?)`, lopnr).Scan(&namn, &birth, &sex)
	if err != nil {
		// Use http.Error instead of log.Fatal
		http.Error(w, fmt.Sprintf("Could not load person details: %v", err), http.StatusInternalServerError)
		return
	}

	renderPersoner(w, "person_editform", PersonRow{
		Namn:  toUtf8(namn),
		Birth: birth,
		Sex:   sex,
		Lopnr: lopnr,
	})
}

func addformPerson(w http.ResponseWriter) {
	fmt.Println("addformPerson ")

	renderPersoner(w, "person_addform", nil)
}

func skapaPerson(db *sql.DB, namn string, birth int, sex string) error {
	if db == nil {
		return fmt.Errorf("skapaPerson anropad med db=nil")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`INSERT INTO Personer(Namn, Född, Kön) VALUES (?, ?, ?)`, namn, birth, sex)

	if err != nil {
		return fmt.Errorf("failed to insert person: %w", err)
	}
	return nil
}

func addPerson(w http.ResponseWriter, namn string, birth string, sex string, db *sql.DB) {
	fmt.Println("addPerson namn: ", namn)

	birthint, err := strconv.Atoi(birth)
	if err != nil {
		personerMsg(w, "Person ej tillagd, felaktigt födelseår.")
		return // Return instead of log.Fatal
	}

	err = skapaPerson(db, namn, birthint, sex)

	if err != nil {
		// Log the error but don't crash the request handler
		log.Printf("Error adding person: %v", err)
		personerMsg(w, "Kunde inte lägga till personen: %v", err)
		return
	}
	personerMsg(w, "Person %s tillagd.", namn)
}

func updatePerson(w http.ResponseWriter, lopnr int, namn string, birth string, sex string, db *sql.DB) {
	fmt.Println("updatePerson lopnr: ", lopnr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`UPDATE Personer SET Namn = ?, Född = ?, Kön = ? WHERE (Löpnr=?)`, namn, birth, sex, lopnr)

	if err != nil {
		// Use http.Error instead of log.Fatal
		http.Error(w, fmt.Sprintf("Could not update person: %v", err), http.StatusInternalServerError)
		return
	}
	personerMsg(w, "Person %s uppdaterad.", namn)
}

func hanterapersoner(w http.ResponseWriter, req *http.Request) {
	renderPersoner(w, "personer_header", currentDatabase)

	err := req.ParseForm()
	if err != nil {
		log.Printf("Error parsing form: %v", err)
		// Continue rendering the page structure even if form parsing fails partially
	}

	formaction := req.FormValue("action")
	var lopnr = -1
	if len(req.FormValue("lopnr")) > 0 {
		parsedLopnr, err := strconv.Atoi(req.FormValue("lopnr"))
		if err == nil {
			lopnr = parsedLopnr
		}
	}

	// Handle actions that modify data (these write their own response content)
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
		// If no action is specified, we proceed to render the list via printPersoner
		fmt.Println("No specific action taken.")
	}

	// Render the main table content using templates
	printPersoner(w, db)

	// --- End HTML Structure Writing ---
	printPersonerFooter(w)
}

func getPersonNames() []string {
	names := make([]string, 0)

	res, err := db.Query("select DISTINCT Namn from Personer union select DISTINCT Vem from Transaktioner order by Namn")

	if err != nil {
		log.Printf("Error getting person names: %v", err)
		return nil // Return empty slice on error
	}

	var Namn []byte // size 50, index
	for res.Next() {
		_ = res.Scan(&Namn)
		names = append(names, toUtf8(Namn))
	}
	res.Close()
	return names
}

func antalPersoner(db *sql.DB) int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var antal int

	err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM Personer`).Scan(&antal)
	if err != nil {
		log.Printf("Error counting persons: %v", err)
		return 0 // Return 0 on error
	}

	return antal
}

func hämtaPerson(db *sql.DB, lopnr int) person {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var namn []byte  // size 50
	var birth string // size 4 (år, 0 för Gemensamt)
	var sex string   // size 10 (text: Gemensamt, Man, Kvinna)

	err := db.QueryRowContext(ctx,
		`SELECT Namn,Född,Kön FROM Personer WHERE (Löpnr=?)`, lopnr).Scan(&namn, &birth, &sex)
	if err != nil {
		log.Printf("Error fetching person: %v", err)
		// Return zero-value struct on error instead of crashing
		return person{}
	}

	var retperson person

	retperson.namn = toUtf8(namn)
	retperson.birth, _ = strconv.Atoi(birth)
	retperson.sex = sex

	return retperson
}
