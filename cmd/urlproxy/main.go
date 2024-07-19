package main

import (
	"encoding/base64"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	safeBrowsingAPIKey = os.Getenv("SB_API_KEY")
)

type SafetyProvider interface {
	Init() error
	Shutdown() error
	Check(url *url.URL) (bool, error)
}

var (
	sbProvider *SBProvider
	upProvider *URLHausProvider
	ppProvider *PhishTankProvider

	templateBlock        = template.Must(template.ParseFiles("templates/block.html"))
	templateConfirmation = template.Must(template.ParseFiles("templates/confirmation.html"))
)

func check(w http.ResponseWriter, req *http.Request) {

	if req.Method != "GET" {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	requestEncoded := strings.TrimPrefix(req.URL.Path, "/check/")

	if len(requestEncoded) == 0 {
		http.Error(w, "missing url", http.StatusBadRequest)
		return
	}

	urlBinary, err := base64.RawURLEncoding.DecodeString(requestEncoded)

	if err != nil {
		http.Error(w, "invalid url encoding", http.StatusBadRequest)
		return
	}

	u, err := url.Parse(string(urlBinary))
	if err != nil {
		// TODO better user feedback in case of invalid URLs?
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	// TODO refactor check function, check links on pages too
	// Add checks for URL redirector

	threatSb, err := sbProvider.Check(u)
	if err != nil {
		// TODO better user feedback in case of invalid URLs?
		http.Error(w, "internal server error, please try again later", http.StatusInternalServerError)
		return
	}

	threatUp, err := upProvider.Check(u)
	if err != nil {
		// TODO better user feedback in case of invalid URLs?
		http.Error(w, "internal server error, please try again later", http.StatusInternalServerError)
		return
	}

	threatPp, err := ppProvider.Check(u)
	if err != nil {
		// TODO better user feedback in case of invalid URLs?
		http.Error(w, "internal server error, please try again later", http.StatusInternalServerError)
		return
	}

	if threatSb || threatUp || threatPp {

		log.Printf("Requested: '%v' - Threat status SafeBrowsing %v - URLHaus %v - PhishTank %v", string(urlBinary), threatSb, threatUp, threatPp)
		templateBlock.Execute(w, struct{ URL string }{u.String()})

	} else {

		templateConfirmation.Execute(w, struct{ URL string }{u.String()})

	}

}

func main() {

	if len(safeBrowsingAPIKey) == 0 {
		log.Fatal("Missing Google SafeBrowsing API Key")
	}

	sbProvider = new(SBProvider)
	log.Println(sbProvider.Init())

	upProvider = new(URLHausProvider)
	log.Println(upProvider.Init())

	ppProvider = new(PhishTankProvider)
	log.Println(ppProvider.Init())

	http.HandleFunc("/check/", check)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.ListenAndServe(":8080", nil)
}
