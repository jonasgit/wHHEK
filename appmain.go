//-*- coding: utf-8 -*-

//go:build !mkdemodb

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pkg/browser" // BSD-2-Clause
)

func main() {
	helpPtr := flag.Bool("help", false, "Skriv ut hjälptext.")
	llmPortPtr := flag.Int("llmport", llmDefaultPort, "Port för LLM-servern på localhost, som används av sidan Chat.")
	llmContextPtr := flag.Int("llmcontext", llmDefaultContextPercent, "Kontextlängd för LLM-modellen, i procent av modellens största (1-100). Gäller LM Studio.")

	flag.Parse()

	if *helpPtr {
		flag.Usage()
		os.Exit(1)
	}

	if err := llmCheckOptions(*llmPortPtr, *llmContextPtr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		flag.Usage()
		os.Exit(2)
	}
	llmBaseURL = llmURLForPort(*llmPortPtr)
	llmContextPercent = *llmContextPtr

	dbdecimaldot = !detectdbdec()
	log.Println("Setting dbdecimaldot: ", dbdecimaldot)

	http.HandleFunc("/hello", hello)
	http.HandleFunc("/r/e/transaction", r_e_transaction)
	//http.HandleFunc("/r/e/fastutg", r_e_fastutg)
	http.HandleFunc("/r/", restapi)
	http.HandleFunc("/htmx.js", htmx)
	http.HandleFunc("/img/bars.svg", imgbars)
	http.HandleFunc("/headers", headers)
	http.HandleFunc("/open", opendb)
	http.HandleFunc("/createdb", createdb)
	http.HandleFunc("/pwd", checkpwd)
	http.HandleFunc("/close", closedb)
	http.HandleFunc("/quit", quitapp)
	http.HandleFunc("/newtrans", newtransaction)
	http.HandleFunc("/hnewtrans", addtransaction)
	http.HandleFunc("/fixedtrans", fixedtransactionHTML)
	http.HandleFunc("/editfixedtrans", editfixedtransactionHTML)
	http.HandleFunc("/editfastutg", editfastutgHTML)
	http.HandleFunc("/editfastink", editfastinkHTML)
	http.HandleFunc("/editfastovf", editfastovfHTML)
	//	http.HandleFunc("/addtrans", addtransaction)
	http.HandleFunc("/monthly", monthly)
	http.HandleFunc("/transactions", transactions)
	http.HandleFunc("/htransactions", htransactions)
	http.HandleFunc("/platser", hanteraplatser)
	http.HandleFunc("/personer", hanterapersoner)
	http.HandleFunc("/konton", hanterakonton)
	http.HandleFunc("/budget", hanteraBudget)
	http.HandleFunc("/yresult", hanteraYResult)
	http.HandleFunc("/y3result", hanteraY3Result)
	http.HandleFunc("/y10result", hanteraY10Result)
	http.HandleFunc("/ybr", hanteraYBR)
	http.HandleFunc("/summary", generateSummary)
	http.HandleFunc("/acccmp", compareaccount)
	http.HandleFunc("/passwd", passwordmgmt)
	http.HandleFunc("/help1", help1)
	http.HandleFunc("/faq1", faq1)
	http.HandleFunc("/components", comps)
	http.HandleFunc("/chat", chat)
	http.HandleFunc("/chatpage", chatPage)
	http.HandleFunc("/chatask", chatAsk)
	http.HandleFunc("/", root)

	//ip, _ := externalIP()
	fmt.Println("Öppna URL i webläsaren:  http://localhost:8090/")
	//fmt.Printf(" eller :  http://%s:8090/\n", ip)
	browser.OpenURL("http://localhost:8090/")
	srv := &http.Server{
		Addr:           "127.0.0.1:8090",
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	httpSrv = srv
	httpError := srv.ListenAndServe()
	if httpError != nil && httpError != http.ErrServerClosed {
		log.Println("While serving HTTP: ", httpError)
	}
	closeDB()
}
