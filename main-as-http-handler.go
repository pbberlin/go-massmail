package main

import (
	"log"
	"net/http"
	"net/http/httptest"
)

// RegistrationFMTEnH shows a registraton form for the FMT
func RegistrationFMTEnH(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// if false {
	// 	gm.ExampleRaw()
	// }
	// ExampleUsingLib()
	// gm.ReadCSVExample()
	ProcessCSV()
}

func main() {

	log.SetFlags(log.Lshortfile | log.Ltime)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	RegistrationFMTEnH(w, req)

}
