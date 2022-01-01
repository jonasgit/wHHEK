//-*- coding: utf-8 -*-

package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func setPasswd(w http.ResponseWriter, pwd2 string, pwd3 string) {
	_, _ = fmt.Fprintf(w, "<html><body>")

	oldpwd := getdbpw(db)
	if oldpwd != " " {
		_, _ = fmt.Fprintf(w, "Det finns redan lösenord satt. Använd byt-knappen för att byta.")
		_, _ = fmt.Fprintf(w, "</body></html>")
		return
	}
	if pwd2 != pwd3 {
		_, _ = fmt.Fprintf(w, "Lösenorden är olika inskrivna.")
		_, _ = fmt.Fprintf(w, "</body></html>")
		return
	}
	_ = setdbpw(db, pwd2)
	_, _ = fmt.Fprintf(w, "Lösenord satt.<p>")
	_, _ = fmt.Fprintf(w, "<a href=\"help1\">Hjälp</a><p>\n")
	_, _ = fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	_, _ = fmt.Fprintf(w, "</body></html>")
}

func changePasswd(w http.ResponseWriter, pwd1 string, pwd2 string, pwd3 string) {
	_, _ = fmt.Fprintf(w, "<html><body>")

	oldpwd := getdbpw(db)
	if pwd1 != oldpwd {
		_, _ = fmt.Fprintf(w, "Angett lösenord stämmer inte.")
		_, _ = fmt.Fprintf(w, "</body></html>")
		return
	}

	if pwd2 != pwd3 {
		_, _ = fmt.Fprintf(w, "Lösenorden är olika inskrivna.")
		_, _ = fmt.Fprintf(w, "</body></html>")
		return
	}
	_ = setdbpw(db, pwd2)
	_, _ = fmt.Fprintf(w, "Lösenord bytt.<p>")
	_, _ = fmt.Fprintf(w, "<a href=\"help1\">Hjälp</a><p>\n")
	_, _ = fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	_, _ = fmt.Fprintf(w, "</body></html>")
}

func delPasswd(w http.ResponseWriter, pwd1 string) {
	_, _ = fmt.Fprintf(w, "<html><body>")

	oldpwd := getdbpw(db)
	if pwd1 != oldpwd {
		_, _ = fmt.Fprintf(w, "Angett lösenord stämmer inte.")
		_, _ = fmt.Fprintf(w, "</body></html>")
		return
	}
	_ = setdbpw(db, " ")
	_, _ = fmt.Fprintf(w, "Lösenord borttaget.<p>")
	_, _ = fmt.Fprintf(w, "<a href=\"help1\">Hjälp</a><p>\n")
	_, _ = fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	_, _ = fmt.Fprintf(w, "</body></html>")
}

//go:embed html/passord.html
var htmlpass string

func passwordmgmt(w http.ResponseWriter, req *http.Request) {
	log.Println("passwordmgmt")

	err := req.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	typ := req.FormValue("submit")
	log.Println("typ: ", typ)
	action := req.FormValue("action")
	log.Println("action: ", action)

	pwd1 := req.FormValue("pwd1")
	log.Println("pwd1: ", pwd1)
	pwd2 := req.FormValue("pwd2")
	log.Println("pwd2: ", pwd2)
	pwd3 := req.FormValue("pwd3")
	log.Println("pwd3: ", pwd3)

	switch action {
	case "set":
		setPasswd(w, pwd2, pwd3)
	case "change":
		changePasswd(w, pwd1, pwd2, pwd3)
	case "delete":
		delPasswd(w, pwd1)
	default:
		t := template.New("Lösenordshantering")
		t, _ = t.Parse(htmlpass)
		_ = t.Execute(w, t)
	}
}
