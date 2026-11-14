//-*- coding: utf-8 -*-

package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	llmDefaultPort           = 1234
	llmDefaultContextPercent = 50

	llmCheckTimeout  = 3 * time.Second
	llmAnswerTimeout = 5 * time.Minute
	llmLoadTimeout   = 3 * time.Minute

	// Övre gräns för hur mycket som läses från LLM-servern.
	llmMaxResponseBytes = 16 << 20

	// Kontextlängd att räkna med när servern inte kan ange den, till exempel när
	// modellen inte är laddad än. LM Studio laddar då modellen med 4096 som standard.
	llmDefaultContextTokens = 4096

	// Uppskattning av antal tokens, uppmätt mot LM Studio med gemma-4: en siffra är
	// ungefär en token och övrig text ungefär tre tecken per token (3,7 för svensk text).
	llmCharsPerToken    = 3
	llmTokensPerMessage = 6
)

// llmContextNumbers hittar storlekarna i felet exceed_context_size_error. LM Studio
// skickar llama.cpp:s fel som en sträng, så citattecknen kan vara escapade.
var llmContextNumbers = regexp.MustCompile(`n_prompt_tokens\\*":(\d+),\\*"n_ctx\\*":(\d+)`)

var (
	// Adress till den lokala LLM-servern (OpenAI-kompatibelt API, t.ex. LM Studio).
	// Sätts i main() med -llmport.
	llmBaseURL = llmURLForPort(llmDefaultPort)
	// Hur stor del av modellens största kontextlängd som ska användas. Sätts i main()
	// med -llmcontext.
	llmContextPercent = llmDefaultContextPercent

	// llmContextMu gör att bara en begäran i taget kan ladda om modellen.
	llmContextMu sync.Mutex
	// llmContextFailures minns misslyckade försök att ändra kontextlängden, så att
	// modellen inte laddas om vid varje sidvisning och fråga.
	llmContextFailures = map[string]string{}
	// llmContextSettled minns vilken längd servern gav när den ombads ändra den. LM Studio
	// avrundar uppåt, till exempel 43253 till 43264, och utan det här skulle modellen
	// laddas om vid varje sidvisning och fråga.
	llmContextSettled = map[string]int{}
	// llmContextInfoError minns det senast loggade felet per server, så att samma fel
	// inte loggas vid varje sidvisning och fråga.
	llmContextInfoError = map[string]string{}
)

func llmURLForPort(port int) string {
	return fmt.Sprintf("http://localhost:%d", port)
}

// llmCheckOptions kontrollerar värdena på -llmport och -llmcontext.
func llmCheckOptions(port, contextPercent int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("-llmport ska vara ett portnummer mellan 1 och 65535, inte %d", port)
	}
	if contextPercent < 1 || contextPercent > 100 {
		return fmt.Errorf("-llmcontext ska vara mellan 1 och 100 procent, inte %d", contextPercent)
	}
	return nil
}

// llmProblem är ett fel där servern fungerar men inte kan besvara just den här frågan.
type llmProblem string

func (p llmProblem) Error() string { return string(p) }

// chatBackground skickas som bakgrund med varje fråga: instruktionen att svara med
// SQL och databasens struktur. TestChatBackgroundMatchesDatabase kontrollerar att den stämmer med InitiateDB.
//
//go:embed chat_schema.txt
var chatBackground string

var chatTemplates = template.Must(template.ParseFS(htmlTemplates, "html/chat.html"))

type ChatPageData struct {
	ServerURL   string
	ServerError string
	NoDatabase  bool
	Model       string
	// Modellens kontextlängd, 0 om den inte är känd.
	ContextTokens  int
	ContextWarning string
}

type ChatReplyData struct {
	ServerURL   string
	Question    string
	Result      *chatResult
	SQL         string
	FirstSQL    string // Modellens första fråga, om den behövde ett nytt försök
	FirstError  string // Databasens fel på den första frågan
	Message     string // Inmatningsfel, t.ex. tom fråga
	Problem     string // Servern fungerar men kunde inte besvara frågan
	ServerError string
	History     string // Samtalet som JSON, skickas med nästa fråga
	Dropped     int    // Antal äldre frågor som togs bort ur samtalet vid den här frågan
	Remembered  int    // Antal frågor som finns kvar i samtalet, inklusive den här
}

// chatMessage är ett meddelande i ett samtal, på formen som LLM-servern använder.
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// llmReadBody läser svaret från LLM-servern och returnerar ett fel om statuskoden inte är 200.
func llmReadBody(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(resp.Body, llmMaxResponseBytes))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		if bytes.Contains(body, []byte("exceed_context_size_error")) {
			return nil, llmContextExceeded(body)
		}
		return nil, fmt.Errorf("servern svarade med status %d: %s", resp.StatusCode, llmSnippet(body))
	}
	return body, nil
}

// llmContextExceeded beskriver felet när frågan inte får plats i modellens kontext.
func llmContextExceeded(body []byte) error {
	const advice = "Korta frågan, eller starta programmet med en högre -llmcontext."
	if m := llmContextNumbers.FindSubmatch(body); m != nil {
		return llmProblem(fmt.Sprintf("Frågan är för lång för modellen: %s tokens, men det finns bara plats för %s. %s", m[1], m[2], advice))
	}
	return llmProblem("Frågan är för lång för modellen. " + advice)
}

// llmSnippet kortar av serverns svar så att det får plats i ett felmeddelande.
func llmSnippet(body []byte) string {
	const maxLen = 500
	if len(body) > maxLen {
		body = body[:maxLen]
	}
	return strings.TrimSpace(strings.ToValidUTF8(string(body), ""))
}

// llmFindModel kontrollerar att LLM-servern svarar och returnerar den första
// modellen som kan användas för chat. Servern listar även inbäddningsmodeller,
// och de kan inte besvara frågor.
func llmFindModel(ctx context.Context, baseURL string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, llmCheckTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/v1/models", nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := llmReadBody(resp)
	if err != nil {
		return "", err
	}

	var models struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &models); err != nil {
		return "", fmt.Errorf("oväntat svar från %s/v1/models: %w", baseURL, err)
	}
	for _, model := range models.Data {
		if !strings.Contains(strings.ToLower(model.ID), "embed") {
			return model.ID, nil
		}
	}
	return "", errors.New("servern svarar men har ingen chattmodell")
}

// llmInstance är en laddad instans av en modell i LM Studio.
type llmInstance struct {
	ID     string `json:"id"`
	Config struct {
		ContextLength int `json:"context_length"`
	} `json:"config"`
}

// llmModelInfo hämtar modellens största kontextlängd och laddade instanser.
func llmModelInfo(ctx context.Context, baseURL, model string) (int, []llmInstance, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/v1/models", nil)
	if err != nil {
		return 0, nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	// Servrar utan LM Studios API, t.ex. llama.cpp och Ollama, har inte adressen.
	if resp.StatusCode == http.StatusNotFound {
		return 0, nil, nil
	}
	body, err := llmReadBody(resp)
	if err != nil {
		return 0, nil, err
	}

	var info struct {
		Models []struct {
			Key              string        `json:"key"`
			MaxContextLength int           `json:"max_context_length"`
			LoadedInstances  []llmInstance `json:"loaded_instances"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return 0, nil, err
	}
	for _, m := range info.Models {
		if m.Key == model {
			return m.MaxContextLength, m.LoadedInstances, nil
		}
	}
	return 0, nil, nil
}

func llmPostJSON(ctx context.Context, url string, payload any) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = llmReadBody(resp)
	return err
}

func llmLoadModel(ctx context.Context, baseURL, model string, contextTokens int) error {
	return llmPostJSON(ctx, baseURL+"/api/v1/models/load", map[string]any{"model": model, "context_length": contextTokens})
}

// llmEnsureContext ber LM Studio använda den del av modellens största kontextlängd som
// anges med -llmcontext.
func llmEnsureContext(ctx context.Context, baseURL, model string) (int, error) {
	return llmEnsureContextPercent(ctx, baseURL, model, llmContextPercent)
}

// llmEnsureContextPercent laddar om modellen om den har en annan kontextlängd än percent
// procent av den största. Returnerar längden som gäller, eller 0 om den inte är känd.
// Servrar utan LM Studios API lämnas som de är, utan fel.
func llmEnsureContextPercent(ctx context.Context, baseURL, model string, percent int) (int, error) {
	llmContextMu.Lock()
	defer llmContextMu.Unlock()

	infoCtx, cancel := context.WithTimeout(ctx, llmCheckTimeout)
	defer cancel()
	maxContext, instances, err := llmModelInfo(infoCtx, baseURL, model)
	if err != nil {
		if llmContextInfoError[baseURL] != err.Error() {
			llmContextInfoError[baseURL] = err.Error()
			log.Println("chat: kan inte läsa modellens kontextlängd: ", err)
		}
		return 0, nil
	}
	delete(llmContextInfoError, baseURL)
	target := maxContext * percent / 100
	if target <= 0 {
		return 0, nil
	}

	key := fmt.Sprintf("%s %s %d", baseURL, model, target)
	settled := llmContextSettled[key]
	previous := 0
	current := 0
	var wrong []llmInstance
	for _, instance := range instances {
		if length := instance.Config.ContextLength; length == target || length == settled {
			current = length
		} else {
			wrong = append(wrong, instance)
			previous = length
		}
	}
	if current > 0 && len(wrong) == 0 {
		return current, nil
	}

	if failure, failed := llmContextFailures[key]; failed {
		return previous, errors.New(failure)
	}

	// Avbryts inte om webbläsaren lämnar sidan, annars kan modellen bli urladdad.
	loadCtx, cancelLoad := context.WithTimeout(context.WithoutCancel(ctx), llmLoadTimeout)
	defer cancelLoad()

	fail := func(err error) (int, error) {
		failure := fmt.Sprintf("Kunde inte ändra kontextlängden till %d tokens: %v", target, err)
		llmContextFailures[key] = failure
		log.Println("chat: ", failure)
		return previous, errors.New(failure)
	}

	log.Printf("chat: laddar om %s med kontextlängden %d", model, target)
	for _, instance := range wrong {
		err := llmPostJSON(loadCtx, baseURL+"/api/v1/models/unload", map[string]string{"instance_id": instance.ID})
		if err != nil {
			return fail(err)
		}
	}
	if current == 0 {
		if err := llmLoadModel(loadCtx, baseURL, model, target); err != nil {
			if previous > 0 {
				if restoreErr := llmLoadModel(loadCtx, baseURL, model, previous); restoreErr != nil {
					log.Println("chat: kunde inte ladda tillbaka modellen: ", restoreErr)
					previous = 0
				}
			}
			return fail(err)
		}
		current = target
		if _, loaded, err := llmModelInfo(loadCtx, baseURL, model); err == nil && len(loaded) > 0 {
			current = loaded[0].Config.ContextLength
		}
		llmContextSettled[key] = current
	}
	return current, nil
}

// llmContextTokens returnerar modellens kontextlängd. Bara LM Studio anger den,
// så med andra servrar eller vid fel används llmDefaultContextTokens.
func llmContextTokens(ctx context.Context, baseURL, model string) int {
	ctx, cancel := context.WithTimeout(ctx, llmCheckTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/v0/models", nil)
	if err != nil {
		return llmDefaultContextTokens
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return llmDefaultContextTokens
	}
	defer resp.Body.Close()

	body, err := llmReadBody(resp)
	if err != nil {
		return llmDefaultContextTokens
	}

	var models struct {
		Data []struct {
			ID                  string `json:"id"`
			LoadedContextLength int    `json:"loaded_context_length"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &models); err != nil {
		return llmDefaultContextTokens
	}
	for _, m := range models.Data {
		if m.ID == model && m.LoadedContextLength > 0 {
			return m.LoadedContextLength
		}
	}
	return llmDefaultContextTokens
}

// llmAsk skickar samtalet till LLM-servern och returnerar svaret, och om svaret
// kapades för att modellen fick slut på utrymme.
func llmAsk(ctx context.Context, baseURL, model string, messages []chatMessage) (reply string, truncated bool, err error) {
	payload, err := json.Marshal(struct {
		Model    string        `json:"model"`
		Messages []chatMessage `json:"messages"`
		Stream   bool          `json:"stream"`
	}{
		Model:    model,
		Messages: messages,
	})
	if err != nil {
		return "", false, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", false, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	body, err := llmReadBody(resp)
	if err != nil {
		return "", false, err
	}

	var completion struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &completion); err != nil {
		return "", false, fmt.Errorf("oväntat svar från servern: %w", err)
	}
	if len(completion.Choices) == 0 {
		return "", false, fmt.Errorf("servern skickade inget svar: %s", llmSnippet(body))
	}
	reply = strings.TrimSpace(completion.Choices[0].Message.Content)
	truncated = completion.Choices[0].FinishReason == "length"
	if reply == "" {
		if truncated {
			return "", false, llmProblem("Modellen fick slut på utrymme innan den hann svara. Korta frågan, välj Rensa historik eller starta programmet med en högre -llmcontext.")
		}
		return "", false, llmProblem("Modellen skickade ett tomt svar. Försök igen.")
	}
	return reply, truncated, nil
}

// chatParseHistory läser samtalet som webbsidan skickar med frågan. Historiken
// finns bara i webbsidan, så den försvinner när sidan laddas om.
func chatParseHistory(history string) ([]chatMessage, error) {
	if history == "" {
		return nil, nil
	}
	var messages []chatMessage
	err := json.Unmarshal([]byte(history), &messages)
	return messages, err
}

// chatEstimateTokens uppskattar hur många tokens en text blir.
func chatEstimateTokens(text string) int {
	digits, others := 0, 0
	for _, r := range text {
		if unicode.IsDigit(r) {
			digits++
		} else {
			others++
		}
	}
	return digits + others/llmCharsPerToken
}

// chatTrimHistory tar bort de äldsta frågorna och svaren tills samtalet ryms i hälften
// av det som bakgrunden lämnar av kontexten, så att resten räcker till svaret. Annars
// tar LM Studio utan varning bort meddelanden mitt i samtalet när kontexten blir full.
// Returnerar det som är kvar och antalet borttagna frågor. Bakgrunden och den nya
// frågan tas aldrig bort.
func chatTrimHistory(history []chatMessage, question, background string, contextTokens int) ([]chatMessage, int) {
	limit := (contextTokens - chatEstimateTokens(background) - llmTokensPerMessage) / 2
	total := chatEstimateTokens(question) + llmTokensPerMessage
	for _, m := range history {
		total += chatEstimateTokens(m.Content) + llmTokensPerMessage
	}

	dropped := 0
	for len(history) > 0 && total > limit {
		// Historiken består av par: en fråga och dess svar.
		pair := history[:min(2, len(history))]
		for _, m := range pair {
			total -= chatEstimateTokens(m.Content) + llmTokensPerMessage
		}
		history = history[len(pair):]
		dropped++
	}
	return history, dropped
}

// chatAskLLM skickar frågan tillsammans med tidigare meddelanden, kör SQL-frågan som
// modellen svarar med och fyller i resultatet och det uppdaterade samtalet i data.
// Resultatet skickas inte till modellen och sparas inte i samtalet.
func chatAskLLM(ctx context.Context, baseURL string, database *chatDatabase, history []chatMessage, data *ChatReplyData) error {
	model, err := llmFindModel(ctx, baseURL)
	if err != nil {
		return err
	}

	// Modellen kan ha laddats om med standardvärden sedan sidan visades.
	if _, err := llmEnsureContext(ctx, baseURL, model); err != nil {
		log.Println("chat: ", err)
	}

	contextTokens := llmContextTokens(ctx, baseURL, model)
	background := chatSystemPrompt(database.Dialect, time.Now(), chatNameLists(ctx, database))
	history, data.Dropped = chatTrimHistory(history, data.Question, background, contextTokens)

	messages := append(history, chatMessage{Role: "user", Content: data.Question})
	data.Remembered = len(history)/2 + 1
	// Bakgrunden sparas inte i samtalet utan läggs till varje gång.
	request := append([]chatMessage{{Role: "system", Content: background}}, messages...)

	// Om databasen inte kan köra frågan får modellen felet och ett nytt försök. Det
	// misslyckade försöket visas på sidan men sparas inte i samtalet.
	query, err := chatAskAndRun(ctx, baseURL, model, database, request, data)
	var failed *chatQueryError
	if errors.As(err, &failed) {
		data.FirstSQL = failed.query
		data.FirstError = failed.Error()
		log.Println("chat: nytt försök efter fel: ", failed)
		query, err = chatAskAndRun(ctx, baseURL, model, database, append(request, chatRetryMessages(failed)...), data)
		if errors.As(err, &failed) {
			return llmProblem(fmt.Sprintf("SQL-frågan gick inte att köra, inte heller efter ett nytt försök: %v.", failed))
		}
	}
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(append(messages, chatMessage{Role: "assistant", Content: query}))
	if err != nil {
		return err
	}
	data.History = string(encoded)
	return nil
}

// chatAskAndRun frågar modellen och kör SQL-frågan den svarar med.
func chatAskAndRun(ctx context.Context, baseURL, model string, database *chatDatabase, request []chatMessage, data *ChatReplyData) (string, error) {
	reply, truncated, err := llmAsk(ctx, baseURL, model, request)
	if err != nil {
		return "", err
	}
	if truncated {
		return "", llmProblem("Modellen fick slut på utrymme innan SQL-frågan var klar. Korta frågan eller välj Rensa historik.")
	}
	return chatRunReply(ctx, database, reply, data)
}

// chatAllowSlowReply förlänger webbserverns skrivgräns för den här begäran.
// Servern i main() har 10 sekunder, och det räcker sällan för att få svar från
// en språkmodell.
func chatAllowSlowReply(w http.ResponseWriter) {
	deadline := time.Now().Add(llmAnswerTimeout + llmCheckTimeout + 10*time.Second)
	// ErrNotSupported kommer bara från httptest.ResponseRecorder i testerna.
	if err := http.NewResponseController(w).SetWriteDeadline(deadline); err != nil && !errors.Is(err, http.ErrNotSupported) {
		log.Println("chat: kunde inte förlänga skrivgränsen: ", err)
	}
}

// chat visar sidan direkt med en väntesymbol. Innehållet hämtas sedan från
// /chatpage, som kan behöva vänta på LLM-servern.
func chat(w http.ResponseWriter, req *http.Request) {
	log.Println("Func chat")
	if err := chatTemplates.ExecuteTemplate(w, "chat.html", nil); err != nil {
		log.Println("While serving HTTP chat: ", err)
	}
}

func serveChat(w http.ResponseWriter, req *http.Request, baseURL string, database *chatDatabase) {
	// Att ladda om modellen kan ta längre tid än webbserverns tidsgräns.
	chatAllowSlowReply(w)

	data := ChatPageData{ServerURL: baseURL, NoDatabase: database == nil}
	model, err := llmFindModel(req.Context(), baseURL)
	if err != nil {
		data.ServerError = err.Error()
	} else {
		data.Model = model
		data.ContextTokens, err = llmEnsureContext(req.Context(), baseURL, model)
		if err != nil {
			data.ContextWarning = err.Error()
		}
	}

	if err := chatTemplates.ExecuteTemplate(w, "chat_page", data); err != nil {
		log.Println("While serving HTTP chatpage: ", err)
	}
}

func serveChatAsk(w http.ResponseWriter, req *http.Request, baseURL string, database *chatDatabase) {
	if req.Method != http.MethodPost {
		http.Error(w, "Endast POST", http.StatusMethodNotAllowed)
		return
	}

	data := ChatReplyData{
		ServerURL: baseURL,
		Question:  strings.TrimSpace(req.FormValue("question")),
	}

	// Fel visas i webbsidan. htmx uppdaterar inte sidan vid andra statuskoder än 2xx.
	history, historyErr := chatParseHistory(req.FormValue("history"))
	if database == nil {
		data.Message = "Ingen databas är öppen."
	} else if data.Question == "" {
		data.Message = "Skriv en fråga."
	} else if historyErr != nil {
		log.Println("chat: ", historyErr)
		data.Message = "Samtalet kunde inte läsas. Välj Rensa historik och försök igen."
	} else {
		chatAllowSlowReply(w)

		ctx, cancel := context.WithTimeout(req.Context(), llmAnswerTimeout)
		defer cancel()

		if err := chatAskLLM(ctx, baseURL, database, history, &data); err != nil {
			log.Println("chat: ", err)
			var problem llmProblem
			if errors.As(err, &problem) {
				data.Problem = problem.Error()
			} else {
				data.ServerError = err.Error()
			}
		}
	}

	if err := chatTemplates.ExecuteTemplate(w, "chat_reply", data); err != nil {
		log.Println("While serving HTTP chat_reply: ", err)
	}
}

func chatPage(w http.ResponseWriter, req *http.Request) {
	log.Println("Func chatPage")
	serveChat(w, req, llmBaseURL, chatCurrentDatabase())
}

func chatAsk(w http.ResponseWriter, req *http.Request) {
	log.Println("Func chatAsk")
	serveChatAsk(w, req, llmBaseURL, chatCurrentDatabase())
}
