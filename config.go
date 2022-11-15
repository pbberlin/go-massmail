package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
	"time"
)

var operationMode string // test or prod

var loc *time.Location // init in load config

func init() {

	writeExampleConfig()

	// defining  string flag
	// strFlag := flag.String("language", "Golang", "Golang is the awesome google language")
	om := flag.String(
		"mode",
		"invalid-default", // default
		"mode must be 'test' or 'prod' \n\tgo-massmail -mode=test", // can be one or two leading hyphens
	)
	flag.Parse()

	operationMode = *om
	if operationMode != "test" && operationMode != "prod" {
		log.Fatalf("mode must be 'test' or 'prod', was %q\n\tgo-massmail -mode=test", operationMode)
	}
	log.Printf("\n\toperation mode is %q\n", operationMode)

	// read prod config
	bts2, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("could not read config\n\t%v", err)
	}

	err = json.Unmarshal(bts2, &cfg)
	if err != nil {
		log.Fatalf("could not unmarsch config\n\t%v", err)
	}

	loc, err = time.LoadLocation(cfg.Location)
	if err != nil {
		log.Printf("configured location %v failed; using UTC_-2\n\t%v", cfg.Location, err)
		loc = time.FixedZone("UTC_-2", -2*60*60)
	}

}

type RelayHorst struct {
	HostNamePort string
	Username     string
	// password, see getenv
}

func (rh RelayHorst) PasswortEnv() string {
	pureHost := strings.Split(rh.HostNamePort, ":")[0]
	env := fmt.Sprintf("PW_%v", pureHost)
	env = strings.Replace(env, ".", "", -1)
	env = strings.ToUpper(env)
	return env
}

func (rh RelayHorst) getAuth() (auth smtp.Auth) {

	if rh.Username == "" {
		return nil
	}

	pureHost := strings.Split(rh.HostNamePort, ":")[0]
	env := rh.PasswortEnv()
	pw := os.Getenv(env)
	if pw == "" {
		log.Fatalf(`Set password for %v via ENV %v
		SET %v=secret 
		export %v=secret  
		`,
			pureHost, env,
			env,
			env,
		)
	}

	return smtp.PlainAuth(
		"",
		rh.Username,
		pw,
		pureHost,
	)
}

type WaveT struct {
	Year                   int        `json:"year,omitempty"`
	Month                  time.Month `json:"month,omitempty"`
	ClosingDatePreliminary time.Time  `json:"closing_date_preliminary,omitempty"`
	ClosingDateLastDue     time.Time  `json:"closing_date_last_due,omitempty"`
}

type AttachmentT struct {
	Language string
	Label    string
	Filename string
}

type TaskT struct {
	Name        string        `json:"name,omitempty"`
	Time        time.Time     `json:"time,omitempty"`
	Attachments []AttachmentT `json:"attachments,omitempty"`
	RelayHost   string        `json:"relay_host,omitempty"` // distinct SMTP server for distinct tasks
}

type configT struct {
	Location    string                `json:"loc,omitempty"` // todo
	RelayHorsts map[string]RelayHorst `json:"relay_horsts,omitempty"`
	Waves       map[string][]WaveT    `json:"waves,omitempty"`
	Tasks       map[string][]TaskT    `json:"tasks,omitempty"`
}

func writeExampleConfig() {

	// writeExample already needs a location
	locPreliminary := time.FixedZone("UTC_-2", -2*60*60)

	var example = configT{

		Location: "Europe/Berlin",

		RelayHorsts: map[string]RelayHorst{
			"zimbra.zew.de": {
				HostNamePort: "zimbra.zew.de:25",
				Username:     "fmt-relay",
			},
			"hermes.zew.de": {
				HostNamePort: "hermes.zew.de:25",
			},
			//  from intern
			"hermes.zew-private.de": {
				HostNamePort: "hermes.zew-private.de:25",
			},
			"email.zew.de": {
				HostNamePort: "email.zew.de:25",
			},
		},

		Waves: map[string][]WaveT{
			"fmt": {
				{
					Year:                   2022,
					Month:                  11,
					ClosingDatePreliminary: time.Date(2022, 11, 11+0, 17, 0, 0, 0, locPreliminary),
					ClosingDateLastDue:     time.Date(2022, 11, 11+3, 17, 0, 0, 0, locPreliminary),
				},
				{
					Year:                   2022,
					Month:                  12,
					ClosingDatePreliminary: time.Date(2022, 12, 07+0, 17, 0, 0, 0, locPreliminary),
					ClosingDateLastDue:     time.Date(2022, 12, 07+3, 17, 0, 0, 0, locPreliminary),
				},
			},
		},
		Tasks: map[string][]TaskT{
			"fmt": {
				{
					Name:      "invitation",
					RelayHost: "zimbra.zew.de",
				},
				{
					Name:      "reminder",
					RelayHost: "zimbra.zew.de",
				},
				{
					Name: "results",
					Time: time.Date(2022, 11, 13, 10, 0, 0, 0, locPreliminary),
					Attachments: []AttachmentT{
						{
							Language: "de",
							Label:    "ZEW-FMT-Datenblatt-%v-%02v.pdf",
							Filename: "tabellen/tab.pdf",
						},
						{
							Language: "de",
							Label:    "ZEW-FMT-Pressemitteilung-%v-%02v.pdf",
							Filename: "pressemitteilungen/pressemitteilung_dt.pdf",
						},
						{
							Language: "de",
							Label:    "ZEW-Index-Press-Release-%v-%02v.pdf",
							Filename: "pressemitteilungen/pressemitteilung_en.pdf",
						},
						{
							Language: "en",
							Label:    "ZEW-Index-Data-Table.pdf",
							Filename: "tabellen/tab-engl.pdf",
						},
						{
							Language: "en",
							Label:    "ZEW-Index-Press-Release-%v-%02v.pdf",
							Filename: "pressemitteilungen/pressemitteilung_en.pdf",
						},
					},
					RelayHost: "zimbra.zew.de",
				},
			},
		},
	}

	bts1, _ := json.MarshalIndent(example, "  ", "  ")
	os.WriteFile("example-config.json", bts1, 0777)

}

var cfg = configT{}
