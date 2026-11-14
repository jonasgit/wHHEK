//-*- coding: utf-8 -*-

//go:build mkdemodb

// mkdemodb skapar en slumpad sqlite-databas för demo av wHHEK.
//
// Bygg: go build -tags mkdemodb -o mkdemodb
// Kör : ./mkdemodb -o demo.db
//
// Byggd för windows/386 skapas en mdb-fil (Access/Jet, som wHHEK_x86.exe
// kan öppna), annars en sqlite-fil.
//
// Hushållet består av 1-2 vuxna och 0-3 barn. De vuxna är minst 20 år äldre
// än det äldsta barnet. Det finns ett Lönekonto och ett Sparkonto som båda
// börjar på 0-50000 kr. Ca 25000 transaktioner skapas under de senaste 3-10
// åren. Sparkontot slutar strax över sitt startvärde och Lönekontot slutar
// mellan 0 och 50000 kr.

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const (
	demoLon    = "Lönekonto"
	demoSpar   = "Sparkonto"
	demoPlanbo = "Plånboken"
	demoGemen  = "Gemensamt"
	demoDash   = "---"
	dateFmt    = "2006-01-02"
)

// demoTx är en transaktion i minnet. Belopp i öre.
type demoTx struct {
	date            time.Time
	typ, from, to   string
	what, who, text string
	ore             int64
	prio            int // ordning inom samma dag: inkomster först
	seq             int
}

type demoCategory struct {
	what     string
	weight   float64
	min, max float64 // kr, före skalning
	places   []string
	cash     bool // kan betalas kontant
}

var demoCategories = []demoCategory{
	{"Livsmedel", 40, 50, 1800, []string{"Stormarknaden", "Kvartersbutiken", "Lågprisbutiken", "Saluhallen"}, true},
	{"Arbetslunch", 10, 85, 145, []string{"Lunchrestaurangen", "Pizzerian", "Sushibaren"}, true},
	{"Förbrukn.varor", 8, 30, 600, []string{"Järnhandeln", "Stormarknaden", "Lågprisbutiken"}, false},
	{"Hygien", 6, 40, 450, []string{"Apoteket", "Stormarknaden"}, false},
	{"Kläder och skor", 6, 150, 1500, []string{"Klädboden", "Skoaffären", "Sportbutiken"}, false},
	{"Lek och fritid", 10, 50, 900, []string{"Biografen", "Badhuset", "Bokhandeln", "Sportbutiken", "Kiosken"}, true},
	{"Bil", 6, 300, 900, []string{"Bensinstationen", "Bilverkstaden"}, false},
	{"Kollektiva resor", 4, 30, 80, []string{"Lokaltrafiken"}, false},
	{"Läkare/tandläkare/medicin", 2, 100, 1200, []string{"Vårdcentralen", "Tandläkaren", "Apoteket"}, false},
	{"Möbler, husgeråd, TV, radio", 2, 200, 4000, []string{"Möbelvaruhuset", "Elektronikbutiken"}, false},
	{"Övriga utgifter", 4, 50, 1000, []string{"Kiosken", "Bokhandeln", "Järnhandeln"}, true},
}

var demoFixedPlaces = []string{"Hyresvärden", "Elbolaget", "Telebolaget", "Försäkringsbolaget",
	"Fackförbundet", "Förskolan", "Resebyrån", "Arbetsgivaren"}

var femaleNames = []string{"Anna", "Eva", "Maria", "Karin", "Sara", "Emma", "Lena", "Kerstin",
	"Malin", "Johanna", "Linnea", "Elin", "Ida", "Frida", "Sofia", "Hanna", "Maja", "Ella", "Alva", "Wilma"}
var maleNames = []string{"Erik", "Lars", "Karl", "Anders", "Johan", "Per", "Mikael", "Fredrik",
	"Daniel", "Magnus", "Oskar", "Viktor", "Axel", "Elias", "Hugo", "Liam", "Noah", "William", "Lucas", "Alexander"}

type demoPerson struct {
	name  string
	born  int
	sex   string
	adult bool
}

type demoGen struct {
	rng   *rand.Rand
	txs   []demoTx
	seq   int
	today time.Time
}

func (g *demoGen) uni(min, max float64) float64 { return min + (max-min)*g.rng.Float64() }

func (g *demoGen) add(tx demoTx) {
	if !tx.date.After(g.today) {
		if tx.typ == "Insättning" {
			tx.prio = -1 // inkomster före utgifter samma dag
		}
		g.seq++
		tx.seq = g.seq
		g.txs = append(g.txs, tx)
	}
}

func kr2ore(kr float64) int64 { return int64(math.Round(kr * 100)) }

// round100 avrundar till närmaste hundralapp, i öre.
func round100(kr float64) int64 { return int64(math.Round(kr/100)) * 10000 }

func dateOf(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

// salaryDate ger löningsdagen (den 25:e, fredagen innan om helg).
func salaryDate(y int, m time.Month) time.Time {
	d := dateOf(y, m, 25)
	switch d.Weekday() {
	case time.Saturday:
		d = d.AddDate(0, 0, -1)
	case time.Sunday:
		d = d.AddDate(0, 0, -2)
	}
	return d
}

func lastDay(y int, m time.Month) time.Time { return dateOf(y, m+1, 0) }

// effect ger transaktionens påverkan på kontot acc, i öre (samma regler som saldoKonto).
func (tx *demoTx) effect(acc string) int64 {
	var e int64
	if tx.to == acc && (tx.typ == "Uttag" || tx.typ == "Fast Inkomst" || tx.typ == "Insättning" || tx.typ == "Överföring") {
		e += tx.ore
	}
	if tx.from == acc && (tx.typ == "Uttag" || tx.typ == "Inköp" || tx.typ == "Fast Utgift" || tx.typ == "Överföring") {
		e -= tx.ore
	}
	return e
}

func pick[T any](g *demoGen, s []T) T { return s[g.rng.IntN(len(s))] }

func main() {
	outPtr := flag.String("o", "demo"+demoExt(), "Filnamn för den nya databasen ("+demoExt()+").")
	seedPtr := flag.Uint64("seed", 0, "Slumpfrö (0 = slumpa fram ett).")
	countPtr := flag.Int("n", 25000, "Ungefärligt antal transaktioner.")
	forcePtr := flag.Bool("f", false, "Skriv över filen om den redan finns.")
	flag.Parse()

	if !strings.HasSuffix(strings.ToLower(*outPtr), demoExt()) {
		fmt.Fprintf(os.Stderr, "Filnamnet måste sluta på %s i den här versionen av mkdemodb.\n", demoExt())
		os.Exit(1)
	}
	if FileExists(*outPtr) && !*forcePtr {
		fmt.Fprintf(os.Stderr, "Filen %s finns redan. Använd -f för att skriva över.\n", *outPtr)
		os.Exit(1)
	}
	seed := *seedPtr
	if seed == 0 {
		seed = uint64(time.Now().UnixNano())
	}

	now := time.Now()
	g := &demoGen{
		rng:   rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15)),
		today: dateOf(now.Year(), now.Month(), now.Day()),
	}

	// --- Tidsperiod ---
	years := 3 + g.rng.IntN(8) // 3-10 år
	sm := g.today.AddDate(-years, 0, 0)
	start := salaryDate(sm.Year(), sm.Month())
	if !start.After(g.today.AddDate(-years, 0, 0)) {
		// Se till att vi håller oss inom de senaste "years" åren.
		nm := sm.AddDate(0, 1, 0)
		start = salaryDate(nm.Year(), nm.Month())
	}

	// --- Personer ---
	thisYear := g.today.Year()
	nAdults := 1 + g.rng.IntN(2)
	nChildren := g.rng.IntN(4)
	var children []demoPerson
	oldestAge := -1
	usedNames := map[string]bool{}
	uniqueName := func(list []string) string {
		for {
			n := pick(g, list)
			if !usedNames[n] {
				usedNames[n] = true
				return n
			}
		}
	}
	for range nChildren {
		age := g.rng.IntN(18) // 0-17 år
		oldestAge = max(oldestAge, age)
		sex := pick(g, []string{"Kvinna", "Man"})
		list := femaleNames
		if sex == "Man" {
			list = maleNames
		}
		children = append(children, demoPerson{uniqueName(list), thisYear - age, sex, false})
	}
	minAdultAge := max(18+years, 22)
	if nChildren > 0 {
		minAdultAge = max(minAdultAge, oldestAge+20+1) // minst 20 år äldre oavsett födelsedag
	}
	var adults []demoPerson
	firstSex := pick(g, []string{"Kvinna", "Man"})
	for i := range nAdults {
		sex := firstSex
		if i == 1 {
			if g.rng.Float64() < 0.85 {
				sex = map[string]string{"Kvinna": "Man", "Man": "Kvinna"}[firstSex]
			}
		}
		list := femaleNames
		if sex == "Man" {
			list = maleNames
		}
		age := minAdultAge + g.rng.IntN(15)
		adults = append(adults, demoPerson{uniqueName(list), thisYear - age, sex, true})
	}

	// --- Konton ---
	lonStart := kr2ore(math.Round(g.uni(0, 50000)))
	sparStart := kr2ore(math.Round(g.uni(0, 50000)))
	sparTarget := sparStart + kr2ore(math.Round(float64(sparStart)/100*g.uni(0.01, 0.04)+g.uni(300, 3000)))
	lonTargetWanted := kr2ore(math.Round(g.uni(2000, 40000)))

	// --- Hushållsekonomi ---
	nPersons := nAdults + nChildren
	rent := g.uni(5000, 7000) + 1500*float64(nPersons) + g.uni(-500, 1500)
	salaries := make([]float64, nAdults)
	for i := range salaries {
		salaries[i] = g.uni(25000, 38000)
		if nAdults == 1 {
			salaries[i] += 3000
		}
	}
	tel := g.uni(300, 700)
	insurance := g.uni(150, 350)
	union := make([]float64, nAdults)
	commuteCard := make([]bool, nAdults)
	for i := range union {
		union[i] = g.uni(250, 500)
		commuteCard[i] = g.rng.Float64() < 0.5
	}
	saveRate := g.uni(0.04, 0.10)

	// Säkerställ att det blir pengar över till rörliga utgifter.
	fixedEst := rent + 1500 + tel + insurance + 1500*float64(nChildren) + 1000*float64(nAdults)
	var salSum float64
	for _, s := range salaries {
		salSum += s
	}
	if need := fixedEst + salSum*saveRate + 9000; salSum < need {
		for i := range salaries {
			salaries[i] += (need - salSum) / float64(nAdults)
		}
	}

	who := func() string {
		r := g.rng.Float64()
		if nAdults == 2 && r < 0.4 {
			return adults[1].name
		}
		if r < 0.8 {
			return adults[0].name
		}
		return demoGemen
	}

	// --- Perioder (från lön till lön) ---
	var periods []time.Time
	for d := start; !d.After(g.today); {
		periods = append(periods, d)
		n := d.AddDate(0, 0, 10) // hamnar alltid i nästa månad
		d = salaryDate(n.Year(), n.Month())
	}
	periodEnd := func(i int) time.Time {
		if i+1 < len(periods) {
			return periods[i+1].AddDate(0, 0, -1)
		}
		return g.today
	}

	// --- Fasta och planerade transaktioner ---
	sparBal := sparStart
	for pi, ps := range periods {
		pe := periodEnd(pi)
		yearsIn := ps.Year() - start.Year()
		raise := math.Pow(1.03, float64(yearsIn))

		// Löner
		var monthSal float64
		for i, a := range adults {
			s := math.Round(salaries[i] * raise * g.uni(0.98, 1.02)) // lite variation (OB, skatt)
			monthSal += s
			g.add(demoTx{date: ps, typ: "Insättning", from: demoDash, to: demoLon, what: "Lön efter skatt",
				who: a.name, text: "Lön", ore: kr2ore(s)})
		}

		// Barnbidrag den 20:e (i perioden), för barn under 16
		bd := dateOf(ps.Year(), ps.Month(), 20).AddDate(0, 1, 0)
		if !bd.After(pe) {
			nb := 0
			for _, c := range children {
				age := bd.Year() - c.born
				if age >= 0 && age < 16 {
					nb++
				}
			}
			if nb > 0 {
				g.add(demoTx{date: bd, typ: "Insättning", from: demoDash, to: demoLon, what: "Barnbidrag",
					who: demoGemen, text: "Barnbidrag", ore: kr2ore(1250 * float64(nb))})
			}
		}

		// Fasta utgifter några dagar efter lönen
		day := func(n int) time.Time {
			d := ps.AddDate(0, 0, n)
			if d.After(pe) {
				d = pe
			}
			return d
		}
		fixed := func(d time.Time, place, what, w, text string, kr float64) {
			if d.Equal(pe) && pe.Sub(ps).Hours() < 24*8 {
				return // ofullständig sista period: fakturan har inte kommit än
			}
			g.add(demoTx{date: d, typ: "Fast Utgift", from: demoLon, to: place, what: what, who: w,
				text: text, ore: kr2ore(math.Round(kr))})
		}
		m := day(5).Month()
		fixed(day(4), "Hyresvärden", "Bostad/hyra utan lån och ränta", demoGemen, "Hyra", rent*math.Pow(1.02, float64(yearsIn)))
		season := 1 + 0.6*math.Cos(2*math.Pi*float64(m-1)/12) // dyrare el på vintern
		fixed(day(6), "Elbolaget", "Hushålls-el", demoGemen, "El", g.uni(500, 900)*season)
		fixed(day(7), "Telebolaget", "Dagstidning, Tel, TV-licens", demoGemen, "Bredband och mobil", tel)
		fixed(day(8), "Försäkringsbolaget", "Hemförsäkring", demoGemen, "Hemförsäkring", insurance)
		for i, a := range adults {
			fixed(day(3), "Fackförbundet", "Fackavgifter", a.name, "Medlemsavgift", union[i])
			if commuteCard[i] {
				fixed(day(1), "Lokaltrafiken", "Kollektiva resor", a.name, "Månadskort", 970)
			}
		}
		nPre := 0
		for _, c := range children {
			age := ps.Year() - c.born
			if age >= 1 && age <= 5 {
				nPre++
			}
		}
		if nPre > 0 {
			fixed(day(5), "Förskolan", "Barnomsorg", demoGemen, "Förskoleavgift", math.Min(1688, 1500*float64(nPre)))
		}

		// Kontantuttag och kontantköp
		for range 1 + g.rng.IntN(3) {
			d := day(g.rng.IntN(int(pe.Sub(ps).Hours()/24) + 1))
			amount := round100(g.uni(200, 1000))
			w := who()
			g.add(demoTx{date: d, typ: "Uttag", from: demoLon, to: demoPlanbo, what: demoDash, who: w,
				text: "Bankomat", ore: amount})
			k := 1 + g.rng.IntN(5)
			parts := g.splitKr(amount/100, k)
			for _, p := range parts {
				var c demoCategory
				for c = pick(g, demoCategories); !c.cash; c = pick(g, demoCategories) {
				}
				cd := d.AddDate(0, 0, g.rng.IntN(8))
				if cd.After(g.today) {
					cd = g.today
				}
				g.add(demoTx{date: cd, typ: "Inköp", from: demoPlanbo, to: pick(g, c.places), what: c.what,
					who: w, text: " ", ore: p * 100})
			}
		}

		// Sparande dagen efter lön
		if pi+1 < len(periods) {
			amt := round100(monthSal * saveRate)
			g.add(demoTx{date: day(1), typ: "Överföring", from: demoLon, to: demoSpar, what: demoDash,
				who: demoGemen, text: "Månadssparande", ore: amt})
			sparBal += amt
		}

		// Ibland tas sparpengar ut till något större inköp
		if pi+1 < len(periods) && sparBal > sparTarget+300000 && g.rng.Float64() < 0.12 {
			amt := round100(float64(sparBal-sparTarget-200000) / 100 * g.uni(0.5, 1.0))
			if amt > 0 {
				d := day(2 + g.rng.IntN(20))
				g.bigPurchase(d, amt, who())
				sparBal -= amt
			}
		}

		// Ränta på sparkontot vid årsskiftet
		ye := lastDay(ps.Year(), 12)
		if !ye.Before(ps) && !ye.After(pe) && sparBal > 0 {
			r := kr2ore(math.Round(float64(sparBal) / 100 * g.uni(0.005, 0.02)))
			if r > 0 {
				g.add(demoTx{date: ye, typ: "Insättning", from: demoDash, to: demoSpar, what: "Övriga inkomster",
					who: demoGemen, text: "Ränta", ore: r})
				sparBal += r
			}
		}
	}

	// Justera sparkontot så att det slutar strax över startvärdet.
	if diff := sparTarget - sparBal; diff > 0 {
		g.add(demoTx{date: g.today, typ: "Överföring", from: demoLon, to: demoSpar, what: demoDash,
			who: demoGemen, text: "Sparande", ore: diff})
	} else if diff < 0 {
		g.bigPurchase(g.today, -diff, who())
	}

	// --- Rörliga utgifter på lönekontot, så att saldot håller sig positivt ---
	fixedCount := len(g.txs)
	wantCount := int(math.Round(float64(*countPtr) * g.uni(0.97, 1.03)))
	varCount := max(wantCount-fixedCount, len(periods))
	totalDays := g.today.Sub(start).Hours()/24 + 1

	bal := lonStart
	assigned := 0
	cumDays := 0.0
	nPer := len(periods)
	for pi, ps := range periods {
		pe := periodEnd(pi)
		days := pe.Sub(ps).Hours()/24 + 1
		cumDays += days
		n := int(math.Round(float64(varCount)*cumDays/totalDays)) - assigned
		assigned += n
		n = max(n, 1)

		var net, inflowAfterStart int64
		for i := range g.txs {
			tx := &g.txs[i]
			if tx.date.Before(ps) || tx.date.After(pe) {
				continue
			}
			e := tx.effect(demoLon)
			net += e
			if e > 0 && tx.date.After(ps) {
				inflowAfterStart += e
			}
		}

		minV := kr2ore(150 * days)
		var target int64
		if pi == nPer-1 {
			target = lonTargetWanted
		} else {
			frac := float64(pi+1) / float64(nPer)
			target = lonStart + int64(float64(lonTargetWanted-lonStart)*frac) + kr2ore(math.Round(g.uni(-3000, 3000)))
			target = max(target, inflowAfterStart+50000)
		}
		v := bal + net - target
		if v < minV {
			v = minV
		}
		if pi == nPer-1 {
			// Sista (ofta ofullständiga) perioden: håll slutsaldot inom 0-50000.
			end := bal + net - v
			if end > 5000000 {
				v = bal + net - 5000000
			}
			if end < 0 {
				v = max(bal+net, 0)
			}
		}
		g.variablePurchases(ps, pe, n, v, who, children)
		bal += net - v
	}

	sort.SliceStable(g.txs, func(i, j int) bool {
		a, b := g.txs[i], g.txs[j]
		if !a.date.Equal(b.date) {
			return a.date.Before(b.date)
		}
		if a.prio != b.prio {
			return a.prio < b.prio
		}
		return a.seq < b.seq
	})

	// --- Skriv databasen ---
	createDemoDB(*outPtr)
	defer closeDB()

	for _, p := range append(adults, children...) {
		if err := skapaPerson(db, p.name, p.born, p.sex); err != nil {
			log.Fatal(err)
		}
	}
	startMonth := strconv.Itoa(int(start.Month()))
	if err := addKonto(demoLon, decimal.New(lonStart, -2), startMonth, db); err != nil {
		log.Fatal(err)
	}
	if err := addKonto(demoSpar, decimal.New(sparStart, -2), startMonth, db); err != nil {
		log.Fatal(err)
	}
	places := map[string]bool{}
	for _, c := range demoCategories {
		for _, p := range c.places {
			places[p] = true
		}
	}
	for _, p := range demoFixedPlaces {
		places[p] = true
	}
	placeNames := make([]string, 0, len(places))
	for p := range places {
		placeNames = append(placeNames, p)
	}
	sort.Strings(placeNames)
	for _, p := range placeNames {
		if err := skapaPlats(db, p, "", false, ""); err != nil {
			log.Fatal(err)
		}
	}

	// Jet (Access-ODBC) klarar inte en transaktion med tiotusentals rader
	// (MaxLocksPerFile är 9500 som standard) och kraschar då i drivrutinen,
	// så där committas i mindre omgångar.
	batchSize := len(g.txs) + 1
	if JetDBSupport {
		batchSize = 1000
	}
	var dbtx *sql.Tx
	for i, tx := range g.txs {
		if i%batchSize == 0 {
			if dbtx != nil {
				if err := dbtx.Commit(); err != nil {
					log.Fatal(err)
				}
			}
			var err error
			if dbtx, err = db.Begin(); err != nil {
				log.Fatal(err)
			}
		}
		addTransaktionSQL(dbtx, tx.typ, tx.from, tx.to, tx.date.Format(dateFmt), tx.what, tx.who,
			decimal.New(tx.ore, -2), tx.text)
	}
	if dbtx != nil {
		if err := dbtx.Commit(); err != nil {
			log.Fatal(err)
		}
	}

	for _, acc := range []string{demoLon, demoSpar, demoPlanbo} {
		updateKontoSaldo(db, acc, saldoKonto(db, acc, ""))
	}

	// --- Sammanfattning ---
	fmt.Printf("Skapade %s (seed %d)\n", *outPtr, seed)
	fmt.Printf("Period: %s - %s (%d år)\n", start.Format(dateFmt), g.today.Format(dateFmt), years)
	for _, p := range append(adults, children...) {
		fmt.Printf("  %-10s född %d  %s\n", p.name, p.born, p.sex)
	}
	fmt.Printf("Transaktioner: %d\n", antalTransaktioner(db))
	for _, acc := range []string{demoLon, demoSpar, demoPlanbo} {
		fmt.Printf("  %-10s start %10s  slut %10s\n", acc,
			Dec2Str(startOf(acc, lonStart, sparStart)), Dec2Str(saldoKonto(db, acc, "")))
	}
}

// demoExt ger filändelsen för databasformatet: mdb (Jet) i 32-bitars
// Windows-bygget, annars sqlite.
func demoExt() string {
	if JetDBSupport {
		return ".mdb"
	}
	return ".db"
}

// createDemoDB skapar en tom databas och öppnar den som den globala db.
func createDemoDB(filename string) {
	if JetDBSupport {
		// Samma decimalavkänning som wHHEK gör vid start.
		dbdecimaldot = !detectdbdec()
		if FileExists(filename) {
			if err := os.Remove(filename); err != nil {
				log.Fatal(err)
			}
		}
		// TOMDB.MDB innehåller redan tabellerna, Plånboken, Gemensamt och budgettyperna.
		if err := os.WriteFile(filename, TOMDB, 0644); err != nil {
			log.Fatal(err)
		}
		db = openJetDB(filename, false)
	} else {
		SkapaTomDB(filename)
		if db != nil {
			db.SetMaxOpenConns(1) // pragman gäller per anslutning
			if _, err := db.Exec("PRAGMA synchronous=OFF"); err != nil {
				log.Fatal(err)
			}
		}
	}
	if db == nil {
		log.Fatal("Kunde inte skapa databasen.")
	}
}

func startOf(acc string, lon, spar int64) decimal.Decimal {
	switch acc {
	case demoLon:
		return decimal.New(lon, -2)
	case demoSpar:
		return decimal.New(spar, -2)
	}
	return decimal.Zero
}

// splitKr delar upp total (hela kronor) i k slumpvisa delar, var och en minst 1 kr.
func (g *demoGen) splitKr(total int64, k int) []int64 {
	if int64(k) > total {
		k = int(max(total, 1))
	}
	w := make([]float64, k)
	var sum float64
	for i := range w {
		w[i] = 0.2 + g.rng.Float64()
		sum += w[i]
	}
	parts := make([]int64, k)
	var used int64
	for i := range parts {
		parts[i] = max(1, int64(math.Floor(float64(total-int64(k))*w[i]/sum))+1)
		used += parts[i]
	}
	parts[k-1] += total - used
	return parts
}

// bigPurchase: uttag från sparkontot till lönekontot följt av ett större inköp samma dag.
func (g *demoGen) bigPurchase(d time.Time, ore int64, w string) {
	type big struct{ place, what, text string }
	b := pick(g, []big{
		{"Resebyrån", "Lek och fritid", "Semesterresa"},
		{"Bilverkstaden", "Bil", "Bilreparation"},
		{"Möbelvaruhuset", "Möbler, husgeråd, TV, radio", "Ny soffa"},
		{"Elektronikbutiken", "Möbler, husgeråd, TV, radio", "Ny TV"},
		{"Tandläkaren", "Läkare/tandläkare/medicin", "Tandbehandling"},
	})
	g.add(demoTx{date: d, typ: "Överföring", from: demoSpar, to: demoLon, what: demoDash, who: w,
		text: b.text, ore: ore, prio: -1})
	g.add(demoTx{date: d, typ: "Inköp", from: demoLon, to: b.place, what: b.what, who: w,
		text: b.text, ore: ore})
}

// variablePurchases skapar n inköp från lönekontot under perioden med summan total (öre).
func (g *demoGen) variablePurchases(ps, pe time.Time, n int, total int64, who func() string,
	children []demoPerson) {
	if total <= 0 {
		return
	}
	if int64(n)*100 > total {
		n = int(max(total/100, 1))
	}
	var wsum float64
	for _, c := range demoCategories {
		wsum += c.weight
	}
	days := int(pe.Sub(ps).Hours()/24) + 1
	txs := make([]demoTx, n)
	raw := make([]float64, n)
	var rawSum float64
	for i := range txs {
		r := g.rng.Float64() * wsum
		var c demoCategory
		for _, c = range demoCategories {
			if r < c.weight {
				break
			}
			r -= c.weight
		}
		u := g.rng.Float64()
		raw[i] = c.min + (c.max-c.min)*u*u
		rawSum += raw[i]
		w := who()
		if c.what == "Lek och fritid" && len(children) > 0 && g.rng.Float64() < 0.3 {
			w = pick(g, children).name
		}
		txs[i] = demoTx{date: ps.AddDate(0, 0, g.rng.IntN(days)), typ: "Inköp", from: demoLon,
			to: pick(g, c.places), what: c.what, who: w, text: " "}
	}
	// Varje inköp är minst 1 kr; resten fördelas efter de slumpade beloppen.
	f := float64(total-int64(n)*100) / rawSum
	var used int64
	big := 0
	for i := range txs {
		txs[i].ore = 100 + int64(math.Floor(raw[i]*f))
		used += txs[i].ore
		if txs[i].ore > txs[big].ore {
			big = i
		}
	}
	txs[big].ore += total - used
	for _, tx := range txs {
		g.add(tx)
	}
}
