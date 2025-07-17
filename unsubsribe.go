package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var unsubscribers = map[string]map[string]map[string]bool{}

func restore(s string) string {
	// s := `https://survey2.zew.de/unsubscribe/fmt/resultshhyyexpectationhhyydata/peterddttbuchmannddtt68aattgmailddttcom?emailqquupeterddttbuchmannddtt68aattgmailddttcommmppprojectqquufmtmmpptaskqquuresultshhyyexpectationhhyydata`
	s = strings.ReplaceAll(s, "mmpp", "&")
	s = strings.ReplaceAll(s, "qquu", "=")
	s = strings.ReplaceAll(s, "ddtt", ".")
	s = strings.ReplaceAll(s, "aatt", "@")
	s = strings.ReplaceAll(s, "hhyy", "-")

	s = strings.ReplaceAll(s, "pct40", "@") // old - temporarily
	s = strings.TrimSpace(s)

	return s
}

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

	flat, err := getCSV("unsubscribe", wave, task, false)
	if err != nil {
		log.Print(err)
		return
	}

	for _, us := range flat {

		us.Project = restore(us.Project)
		us.Task = restore(us.Task)
		us.Email = restore(us.Email)

		if unsubscribers[us.Project] == nil {
			unsubscribers[us.Project] = map[string]map[string]bool{}
		}
		if unsubscribers[us.Project][us.Task] == nil {
			unsubscribers[us.Project][us.Task] = map[string]bool{}
		}

		unsubscribers[us.Project][us.Task][us.Email] = true
	}

	for k := range unsubscribers {
		log.Printf(" project %v has %v unsubsribers", k, len(unsubscribers[k]))
	}

	// temporarily dump
	dbg, _ := json.MarshalIndent(unsubscribers, "", "\t")
	log.Print(string(dbg))
}
