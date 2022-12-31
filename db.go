//-*- coding: utf-8 -*-

package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// FileExists  The below function checks if a regular file (not directory) with a
func FileExists(filepath string) bool {

	fileinfo, err := os.Stat(filepath)

	if os.IsNotExist(err) {
		return false
	}
	// Return false if the fileinfo says the file path is a directory.
	return !fileinfo.IsDir()
}

func openSqlite(filename string) *sql.DB {
	currentDatabase = "NONE"
	dbtype = 0

	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatal(err)
	}

	currentDatabase = filename
	dbtype = 2

	return db
}

func SkapaTomDB(filename string) {
	if FileExists(filename) {
		// Delete file
		err := os.Remove(filename)
		if err != nil {
			log.Println("Failed to remove file. ", err)
		} else {
			log.Println("SkapaTomDB file removed. OK.")
		}
	} else {
		log.Println("SkapaTomDB file did not exist. OK.")
	}

	// Create file
	db = openSqlite(filename)

	if db == nil {
		log.Println("Failed to create database. ")
	} else {
		log.Println("SkapTomDB database created. OK.")
	}

	InitiateDB(db)
}

func InitiateDB(db *sql.DB) {
	if db == nil {
		log.Println("InitiateDB: No DB.")
		return
	}
	log.Println("InitiateDB: Started.")

	sqlStmt := `
  create table Personer (Löpnr integer not null primary key AUTOINCREMENT, Namn text, Född INTEGER, Kön text);
  delete from Personer;
  `
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	sqlStmt = `
  create table Transaktioner (Löpnr integer not null primary key AUTOINCREMENT,FrånKonto TEXT,TillKonto TEXT,Typ TEXT,Datum TEXT,Vad TEXT,Vem TEXT,Belopp DECIMAL(19,4),Saldo DECIMAL(19,4),Fastöverföring BOOLEAN,Text TEXT);
  delete from Transaktioner;
  `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	sqlStmt = `
  create table DtbVer (VerNum text,Benämning text,Losenord text);
  delete from DtbVer;
  `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	sqlStmt = `
  create table BetalKonton (Löpnr integer not null primary key AUTOINCREMENT, Konto TEXT, Kontonummer TEXT, Kundnummer TEXT , Sigillnummer TEXT);
  delete from BetalKonton;
  `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	sqlStmt = `
  create table Betalningar (Löpnr integer not null primary key AUTOINCREMENT,FrånKonto TEXT,TillPlats TEXT,Typ TEXT,Datum TEXT,Vad TEXT,Vem TEXT,Belopp DECIMAL(19,4),Text TEXT,Ranta DECIMAL(19,4),FastAmort DECIMAL(19,4),RorligAmort DECIMAL(19,4),OvrUtg DECIMAL(19,4),LanLopnr INTEGER,Grey TEXT);
  delete from Betalningar;
  `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	sqlStmt = `
  create table Överföringar (Löpnr integer not null primary key AUTOINCREMENT,FrånKonto TEXT,TillKonto TEXT,Belopp DECIMAL(19,4),Datum TEXT,HurOfta TEXT,Vad TEXT,Vem TEXT,Kontrollnr INTEGER,TillDatum TEXT,Rakning TEXT);
  delete from Överföringar;
  `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	sqlStmt = `
  create table Konton (Löpnr integer not null primary key AUTOINCREMENT, KontoNummer TEXT,Benämning TEXT,Saldo DECIMAL(19,4),StartSaldo DECIMAL(19,4),StartManad TEXT,SaldoArsskifte DECIMAL(19,4),ArsskifteManad text);
  delete from Konton;
  `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	sqlStmt = `
  create table LÅN (Löpnr integer not null primary key AUTOINCREMENT,Langivare TEXT,EgenBeskrivn TEXT,LanNummer TEXT,TotLanebelopp DECIMAL(19,4),StartDatum TEXT,RegDatum TEXT,RantJustDatum TEXT,SlutBetDatum TEXT,AktLaneskuld DECIMAL(19,4),RorligDel DECIMAL(19,4),FastDel DECIMAL(19,4),FastRanta REAL,RorligRanta REAL,HurOfta TEXT,Ranta DECIMAL(19,4),FastAmort DECIMAL(19,4),RorligAmort DECIMAL(19,4),OvrUtg DECIMAL(19,4),Rakning TEXT,Vem TEXT,FrånKonto TEXT,Grey TEXT,Anteckningar TEXT,BudgetRanta TEXT,BudgetAmort TEXT,BudgetOvriga TEXT);
  delete from LÅN;
  `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	sqlStmt = `
  create table Platser (Löpnr integer not null primary key AUTOINCREMENT, Namn text, Gironummer text, Typ text, RefKonto);
  delete from Platser;
  `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	sqlStmt = `
  create table Budget (Löpnr integer not null primary key AUTOINCREMENT,Typ TEXT,Inkomst TEXT,HurOfta INTEGER,StartMånad TEXT,Jan DECIMAL(19,4),Feb DECIMAL(19,4),Mar DECIMAL(19,4),Apr DECIMAL(19,4),Maj DECIMAL(19,4),Jun DECIMAL(19,4),Jul DECIMAL(19,4),Aug DECIMAL(19,4),Sep DECIMAL(19,4),Okt DECIMAL(19,4),Nov DECIMAL(19,4),Dec DECIMAL(19,4),Kontrollnr INTEGER);
  delete from Budget;
  `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	/* Data for table BetalKonton */

	/* Data for table Betalningar */

	/* Data for table Budget */
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Lön efter skatt','J',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,1);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Barnbidrag','J',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,2);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Underhållsbidrag','J',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,3);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Bidragsförskott','J',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,4);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Bostadsbidrag','J',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,5);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Studiestöd','J',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,6);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Utbildningsbidrag','J',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,7);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Arbetslöshetsersättning','J',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,8);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Pension','J',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,9);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Sjukpenning','J',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,10);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Föräldrapenning','J',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,11);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Övriga inkomster','J',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,12);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Arbetslunch','N',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,13);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Bostad/hyra utan lån och ränta','N',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,14);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Bostadslån och ränta','N',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,15);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Kollektiva resor','N',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,16);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Bil','N',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,17);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Fackavgifter','N',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,18);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Läkare/tandläkare/medicin','N',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,19);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Barnomsorg','N',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,20);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Underhåll till barn','N',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,21);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Amorteringar','N',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,22);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Räntor','N',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,23);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Övriga utg.-lån','N',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,24);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Övriga utgifter','N',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,25);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Dagstidning, Tel, TV-licens','N',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,26);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Förbrukn.varor','N',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,27);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Hemförsäkring','N',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,28);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Hushålls-el','N',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,29);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Hygien','N',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,30);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Kläder och skor','N',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,31);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Lek och fritid','N',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,32);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Livsmedel','N',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,33);")
	InsertRow("INSERT INTO [Budget] ([Typ],[Inkomst],[HurOfta],[StartMånad],[Jan],[Feb],[Mar],[Apr],[Maj],[Jun],[Jul],[Aug],[Sep],[Okt],[Nov],[Dec],[Kontrollnr],[Löpnr]) VALUES ('Möbler, husgeråd, TV, radio','N',1,'1',0,0,0,0,0,0,0,0,0,0,0,0,NULL,34);")

	/* Data for table DtbVer */
	InsertRow("INSERT INTO [DtbVer] ([VerNum],[Benämning],[Losenord]) VALUES ('3.01','Databas med stöd för betalning till Postgirot',' ');")

	/* Data for table Konton */
	InsertRow("INSERT INTO [Konton] ([KontoNummer],[Benämning],[Saldo],[StartSaldo],[StartManad],[Löpnr],[SaldoArsskifte],[ArsskifteManad]) VALUES ('0','Plånboken',0,0,'Jan',1,0,'Jan');")

	/* Data for table LÅN */

	/* Data for table Personer */
	InsertRow("INSERT INTO [Personer] ([Namn],[Född],[Kön],[Löpnr]) VALUES ('Gemensamt','0','Gemensamt',1);")

	/* Data for table Platser */

	/* Data for table Transaktioner */

	/* Data for table Överföringar */

	log.Println("InitiateDB: Done.")
}

func InsertRow(sqlStmt string) {
	if db == nil {
		log.Println("InsertRow: No DB.")
		return
	}

	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
}

func getdbpw(db *sql.DB) string {
	var Losenord []byte // size 8

	err := db.QueryRow("SELECT Losenord FROM DtbVer").Scan(&Losenord)
	if err != nil {
		log.Fatal(err)
	}
	pw := toUtf8(Losenord)
	//log.Printf("getdbpwd %s %d\n", pw, len(pw))

	return pw
}

func setdbpw(db *sql.DB, pwd string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := db.ExecContext(ctx,
		`UPDATE DtbVer SET Losenord = ?`,
		pwd)
	return err
}
