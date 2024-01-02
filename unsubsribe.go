package main

import (
	"log"
	"os"
	"path/filepath"
)

var unsubscribers = map[string]map[string]bool{}

// fill unsubscribers
func init() {

	dir := filepath.Join(".", "csv", "unsubscribe")
	os.MkdirAll(dir, os.ModePerm)

	wave := WaveT{}
	wave.Year = 1000
	wave.Month = 10

	task := TaskT{}
	task.Name = "unsubscribe"
	task.URL = &UrlT{
		URL: "https://survey2.zew.de/unsubscribe-download",
		TTL: 48 * 3600,
	}

	flat, err := getCSV("unsubscribe", wave, task)
	if err != nil {
		log.Print(err)
		return
	}

	for _, us := range flat {
		if unsubscribers[us.Project] == nil {
			unsubscribers[us.Project] = map[string]bool{}
		}
		unsubscribers[us.Project][us.Email] = true
	}

	for k := range unsubscribers {
		log.Printf(" project %v has %v unsubsribers", k, len(unsubscribers[k]))
	}

	// s, _ := json.MarshalIndent(unsubscribers, "", "\t")
	// log.Print(string(s))
}
