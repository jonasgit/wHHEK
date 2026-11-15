//-*- coding: utf-8 -*-

package main

import (
	_ "embed"
	"html/template"
	"net/http"
)

//go:embed html/setpassord.html
var sethtmlpass string

type SetPwdData struct {
	IsPwdexists bool
	IsPwddiff   bool
	IsPwdok     bool
}

func setPasswd(w http.ResponseWriter, pwd2 string, pwd3 string) {
	oldpwd := getdbpw(db)

	pwdexists := oldpwd != " "
	pwddiff := pwd2 != pwd3
	pwdok := !(pwdexists || pwddiff)

	if pwdok {
		_ = setdbpw(db, pwd2)
	}

	t := template.New("Lösenordshantering")
	t, _ = t.Parse(sethtmlpass)
	data := SetPwdData{
		IsPwdexists: pwdexists,
		IsPwddiff:   pwddiff,
		IsPwdok:     pwdok,
	}
	_ = t.Execute(w, data)
}

//go:embed html/chpass.html
var htmlchpass string

type ChPwdData struct {
	IsPwdMiss bool
	IsPwddiff bool
	IsPwdok   bool
}

func changePasswd(w http.ResponseWriter, pwd1 string, pwd2 string, pwd3 string) {
	oldpwd := getdbpw(db)

	pwdmiss := pwd1 != oldpwd
	pwddiff := pwd2 != pwd3
	pwdok := !(pwdmiss || pwddiff)

	if pwdok {
		_ = setdbpw(db, pwd2)
	}

	t := template.New("Lösenordshantering")
	t, _ = t.Parse(htmlchpass)
	data := ChPwdData{
		IsPwdMiss: pwdmiss,
		IsPwddiff: pwddiff,
		IsPwdok:   pwdok,
	}
	_ = t.Execute(w, data)
}

//go:embed html/delpass.html
var htmldelpass string

type DelPwdData struct {
	IsPwdMiss bool
	IsPwdok   bool
}

func delPasswd(w http.ResponseWriter, pwd1 string) {
	oldpwd := getdbpw(db)

	pwdmiss := pwd1 != oldpwd
	pwdok := !pwdmiss

	if pwdok {
		_ = setdbpw(db, " ")
	}

	t := template.New("Lösenordshantering")
	t, _ = t.Parse(htmldelpass)
	data := DelPwdData{
		IsPwdMiss: pwdmiss,
		IsPwdok:   pwdok,
	}
	_ = t.Execute(w, data)
}

//go:embed html/passord.html
var htmlpass string

func passwordmgmt(w http.ResponseWriter, req *http.Request) {
	//log.Println("passwordmgmt")

	if !parseFormOrFail(w, req) {
		return
	}
	if db == nil {
		writeMessage(w, "Ingen databas är öppen.")
		return
	}

	//typ := req.FormValue("submit")
	//log.Println("typ: ", typ)
	action := req.FormValue("action")
	//log.Println("action: ", action)

	pwd1 := req.FormValue("pwd1")
	//log.Println("pwd1: ", pwd1)
	pwd2 := req.FormValue("pwd2")
	//log.Println("pwd2: ", pwd2)
	pwd3 := req.FormValue("pwd3")
	//log.Println("pwd3: ", pwd3)

	// Lösenordet sparas i DtbVer.Losenord som rymmer 8 tecken.
	if action == "set" || action == "change" {
		if pwd2 == "" || pwd2 == " " {
			writeMessage(w, "Lösenordet ändrades inte: nytt lösenord saknas.")
			return
		}
		if err := checkText(pwd2, maxLenLosenord); err != nil {
			writeMessage(w, "Lösenordet ändrades inte: lösenordet "+err.Error()+".")
			return
		}
	}

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
