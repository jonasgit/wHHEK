//-*- coding: utf-8 -*-

package main

import (
	_ "embed"
	"fmt"
	"log"
	"html/template"
	"net/http"
)

func setPasswd(w http.ResponseWriter, pwd1 string, pwd2 string, pwd3 string) {
	fmt.Fprintf(w, "<html><body>")
	
	oldpwd := getdbpw(db)
	if oldpwd != " " {
		fmt.Fprintf(w, "Det finns redan lösenord satt. Använd byt-knappen för att byta.")
		fmt.Fprintf(w, "</body></html>")
		return
	}
	if pwd2 != pwd3 {
		fmt.Fprintf(w, "Lösenorden är olika inskrivna.")
		fmt.Fprintf(w, "</body></html>")
		return
	}
	setdbpw(db, pwd2)
	fmt.Fprintf(w, "Lösenord satt.<p>")
	fmt.Fprintf(w, "<a href=\"help1\">Hjälp</a><p>\n")
	fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	fmt.Fprintf(w, "</body></html>")
}

func changePasswd(w http.ResponseWriter, pwd1 string, pwd2 string, pwd3 string) {
	fmt.Fprintf(w, "<html><body>")
	
	oldpwd := getdbpw(db)
	if pwd1 != oldpwd {
		fmt.Fprintf(w, "Angett lösenord stämmer inte.")
		fmt.Fprintf(w, "</body></html>")
		return
	}
	
	if pwd2 != pwd3 {
		fmt.Fprintf(w, "Lösenorden är olika inskrivna.")
		fmt.Fprintf(w, "</body></html>")
		return
	}
	setdbpw(db, pwd2)
	fmt.Fprintf(w, "Lösenord bytt.<p>")
	fmt.Fprintf(w, "<a href=\"help1\">Hjälp</a><p>\n")
	fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	fmt.Fprintf(w, "</body></html>")
}

func delPasswd(w http.ResponseWriter, pwd1 string, pwd2 string, pwd3 string) {
	fmt.Fprintf(w, "<html><body>")
	
	oldpwd := getdbpw(db)
	if pwd1 != oldpwd {
		fmt.Fprintf(w, "Angett lösenord stämmer inte.")
		fmt.Fprintf(w, "</body></html>")
		return
	}
	setdbpw(db, " ")
	fmt.Fprintf(w, "Lösenord borttaget.<p>")
	fmt.Fprintf(w, "<a href=\"help1\">Hjälp</a><p>\n")
	fmt.Fprintf(w, "<a href=\"summary\">Översikt</a>\n")
	fmt.Fprintf(w, "</body></html>")
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
	case "set" : setPasswd(w, pwd1, pwd2, pwd3)
	case "change" : changePasswd(w, pwd1, pwd2, pwd3)
	case "delete" : delPasswd(w, pwd1, pwd2, pwd3)
		default :
		t := template.New("Lösenordshantering")
		t, _ = t.Parse(htmlpass)
		t.Execute(w, t)
	}
}
