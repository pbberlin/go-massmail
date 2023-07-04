package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"time"
)

var operationMode string // test or prod

var startTime time.Time

const stfmt = "2006-01-02T15:04" // start time format

var loc *time.Location // init in load config

func init() {

	log.SetFlags(log.Lshortfile | log.Ltime)

	writeExampleConfig()

	//
	//
	// read prod config
	bts2, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("could not read config\n\t%v", err)
	}

	err = json.Unmarshal(bts2, &cfg)
	if err != nil {
		log.Fatalf("could not unmarsch config\n\t%v", err)
	}

	// time zone
	loc, err = time.LoadLocation(cfg.Location)
	if err != nil {
		log.Printf("configured location %v failed; using UTC_-2\n\t%v", cfg.Location, err)
		// loc = time.FixedZone("UTC_-2", 2*60*60)
		loc = time.FixedZone("UTC_+2", 2*60*60)
	}

	// relay host integrity
	if _, ok := cfg.RelayHorsts[cfg.DefaultHorst]; !ok {
		log.Fatalf("cfg.DefaultHorst must be a key to RelayHorsts; %v", cfg.DefaultHorst)
	}
	for project, tasks := range cfg.Tasks {
		for _, t := range tasks {
			if t.RelayHost != "" {
				if _, ok := cfg.RelayHorsts[cfg.DefaultHorst]; !ok {
					log.Fatalf("project %v -  task %v - RelayHost %v does not exist", project, t.Name, t.RelayHost)
				}
			}
		}
	}

	// consistency
	for project, _ := range cfg.Tasks {
		if _, ok := cfg.Projects[project]; !ok {
			log.Fatalf("task %v has no project", project)
		}
	}

	for project, _ := range cfg.Waves {
		if _, ok := cfg.Projects[project]; !ok {
			log.Fatalf("wave %v has no project", project)
		}
	}

	// same as
	for project, tasks := range cfg.Tasks {
		for idx1, t := range tasks {
			if t.SameAs != "" {

			findSameAs:
				for candProj, candTasks := range cfg.Tasks {
					if candProj != project {
						continue
					}
					for _, candT := range candTasks {
						if candT.Name == t.SameAs {
							log.Printf("%v-%-22v will use %v", project, t.Name, candT.Name)

							// more info at t.SameAs
							// ========================
							//   preserve original settings
							orig := t

							// clobber everything else from 'sameAs'
							t = candT

							// restore some original values
							t.Name = orig.Name // determines recipient list
							t.Description = orig.Description
							if orig.RelayHost != "" {
								t.RelayHost = orig.RelayHost
							}
							if orig.URL != nil {
								t.URL = orig.URL
							}
							if orig.Attachments != nil {
								t.Attachments = orig.Attachments
							}
							if orig.ExecutionInterval != "" {
								t.ExecutionInterval = orig.ExecutionInterval
							}
							// this is the tricky setting - more info at t.SameAs
							if orig.TemplateName != "" {
								t.TemplateName = orig.TemplateName
							}

							t.SameAs = "" // prevent transitive copies

							cfg.Tasks[candProj][idx1] = t
							break findSameAs

						}
					}
				}

			}
		}
	}

	// log.Print(util.IndentedDump(cfg))

	//
	//
	// flags
	// (requiring loc set above)
	dsc1 := "mode must be 'test' or 'prod' \n\tgo-massmail -mode=test" // can be one or two leading hyphens
	flg1 := flag.String(
		"mode",            // -mode=xxx
		"invalid-default", // default value
		dsc1,
	)

	dsc2 := "parseable date and time - 2006-01-02T15:04" // can be one or two leading hyphens
	flg2 := flag.String(
		"start",                  // start=2006-01-02T15:04
		time.Now().Format(stfmt), // default value
		dsc2,
	)

	flag.Parse()

	{
		if *flg1 != "test" && *flg1 != "prod" {
			log.Fatalf("mode must be 'test' or 'prod', was %q\n\tgo-massmail -mode=test", *flg1)
		}
		log.Printf("\tmode is %q\n", *flg1)
		operationMode = *flg1
	}

	{
		log.Printf("\tstart time %q\n", *flg2)
		var err error
		startTime, err = time.ParseInLocation(stfmt, *flg2, loc)
		if err != nil {
			log.Printf("\terror parsing start %q - %v", *flg2, err)
			log.Fatalf("\tstart must be parseable '2006-01-02T15:04', was %q\n\tgo-massmail -start=%v", *flg2, time.Now().Format(stfmt))
		}
		dist := time.Until(startTime) // more succint, but less explicit than   startTime.Sub(time.Now())
		if dist < -60*time.Second {
			log.Fatalf("\tstart time %q is %d secs in the past", *flg2, dist/time.Second)
		}
		if dist > 24*time.Hour {
			log.Fatalf("\tstart time cannot be more than 24 hours in the future; %q", *flg2)
		}
	}

}

type RelayHorst struct {
	HostNamePort string `json:"host_name_port,omitempty"`
	// smtp auth
	Username string `json:"username,omitempty"`
	// smtp password, see getenv
	Delay int `json:"delay,omitempty"` // between messages, milliseconds
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

	if false {
		// this fails for exchange servers
		return smtp.PlainAuth(
			"",
			rh.Username,
			pw,
			pureHost,
		)

	}

	return Auth2(
		rh.Username,
		pw,
	)

}

// AttachmentT represents a file attachment for an email
type AttachmentT struct {
	Label    string `json:"label,omitempty"`
	Filename string `json:"filename,omitempty"`
	Language string `json:"language,omitempty"` // matching recipient language - recipient lists are multi-language
	Inline   bool   `json:"inline,omitempty"`
}

// UrlT to download the most recent CSVs from
type UrlT struct {
	URL  string
	TTL  time.Duration // local file is considered stale after TTL has passed
	User string        // http base64 auth, password from environ
}

// WaveT - data that changes with each wave
// but is not task specific
type WaveT struct {
	Year                   int        `json:"year,omitempty"`
	Month                  time.Month `json:"month,omitempty"`
	ClosingDatePreliminary time.Time  `json:"closing_date_preliminary,omitempty"`
	ClosingDateLastDue     time.Time  `json:"closing_date_last_due,omitempty"`
}

// TaskT additional specific data for a wave
type TaskT struct {
	Name        string `json:"name,omitempty"` // no hyphens
	Description string `json:"description,omitempty"`

	//
	// use metadata from another task;
	//   recipient list    is different - because it is derived from task.Name
	//   	template name  is also defaults to task.Name,
	// 	 	template name  can be shared between source and destination
	// 			by setting TemplateName for both
	SameAs string `json:"same_as,omitempty"`

	TemplateName string        `json:"template_name,omitempty"` // default is Name
	Attachments  []AttachmentT `json:"attachments,omitempty"`
	// distinct SMTP server for distinct tasks
	// if empty, then default horst will be chosen
	RelayHost string `json:"relay_host,omitempty"`

	HTML bool `json:"html,omitempty"` // is HTML or plain text

	// CSV file name has not setting - it is always project-taskname-lang.csv
	URL *UrlT `json:"url,omitempty"` // 'wget' URL for recipients CSV

	// a quick fix for sending reminders
	UserIDSkip map[string]string `json:"user_id_skip,omitempty"`

	ExecutionTime     time.Time `json:"execution_time,omitempty"`     // when should the task be started - for cron jobs and parallel tasks
	ExecutionInterval string    `json:"execution_interval,omitempty"` // similar to cron, supersedes Execution time

	testmode bool `json:"-"`
}

// ProjectT is for data across all waves and tasks
type ProjectT struct {
	// sender name - _shown_ by email clients
	// as pointer to avoid json clutter
	From *mail.Address `json:"from,omitempty"`
	// email for responses and auto-responses, if different from 'from', defaults to from
	ReplyTo string `json:"replyto,omitempty"`

	// email for errors due to unknown recipients or postbox full or rejection etc, defaults to from
	// either <noreply@zew.de> or email of admin or operator.
	// Must be reachable from the SMTP gateway; i.e. bounce@zew.de is not reachable by zimbra.zew.de
	Bounce string `json:"bounce,omitempty"`

	TestRecipients []string `json:"test_recipients,omitempty"`
}

type configT struct {
	Location       string                `json:"loc,omitempty"` // todo
	AttachmentRoot string                `json:"attachment_root,omitempty"`
	RelayHorsts    map[string]RelayHorst `json:"relay_horsts,omitempty"`
	DefaultHorst   string                `json:"default_horst,omitempty"` // one of relayhorsts

	// An attempt to route different recipient domains via different smtp gateways.
	// Because zimbra cannot send to internal recipient (zew.de).
	// Because or testing of problematic recipients, such as metzler.com or lbswest.de.
	// Not further pursued, since we dont use zimbra anymore.
	DomainsToRelayHorsts map[string]string `json:"domains_to_relay_horsts,omitempty"`

	// Projects, waves and tasks a related to each other via the map key; i.e. "fmt" or "pds"
	Projects map[string]ProjectT `json:"projects,omitempty"`
	Waves    map[string][]WaveT  `json:"waves,omitempty"`
	Tasks    map[string][]TaskT  `json:"tasks,omitempty"`
}

func writeExampleConfig() {

	// writeExample already needs a location
	locPreliminary := time.FixedZone("UTC_+2", 2*60*60)

	var example = configT{

		Location:       "Europe/Berlin",
		AttachmentRoot: `.\attachments\`,
		// AttachmentRoot: `C:\Users\pbu\Documents\zew_work\daten\`,

		RelayHorsts: map[string]RelayHorst{
			"email.zew.de": {
				HostNamePort: "email.zew.de:587",
				Username:     "pbu",
				Delay:        15200,
			},
			"zimbra.zew.de": {
				HostNamePort: "zimbra.zew.de:25",
				Username:     "fmt-relay",
			},

			//
			//  from intern
			"hermes.zew-private.de": {
				HostNamePort: "hermes.zew-private.de:25",
			},
			"hermes.zew.de": {
				HostNamePort: "hermes.zew.de:25",
			},
		},

		DefaultHorst: "zimbra.zew.de",

		DomainsToRelayHorsts: map[string]string{
			"@zew.de": "email.zew.de",
		},

		Projects: map[string]ProjectT{
			"fmt": {
				//
				From: &mail.Address{
					Name:    "Finanzmarkttest",
					Address: "noreply@zew.de",
				},
				ReplyTo: "finanzmarkttest@zew.de",

				Bounce: "noreply@zew.de",

				TestRecipients: []string{
					"peter.buchmann@web.de",
					"peter.buchmann.68@gmail.com",
					"peter.buchmann@zew.de",
					"no-existing-recipient@gmail.com",
				},
			},
			"pds": {
				//
				From: &mail.Address{
					Name:    "Private Debt Survey",
					Address: "noreply@zew.de",
				},
				ReplyTo: "private-debt-survey@zew.de",

				Bounce: "noreply@zew.de",

				TestRecipients: []string{
					"peter.buchmann@web.de",
					"peter.buchmann.68@gmail.com",
					"peter.buchmann@zew.de",
					"no-existing-recipient@gmail.com",
				},
			},
		},

		Waves: map[string][]WaveT{
			"pds": {
				{
					Year:  2023,
					Month: 01,
				},
			},
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
			"pds": {
				{
					Name:          "invitation",
					Description:   "PDS invitation",
					ExecutionTime: time.Date(2022, 12, 07, 11, 0, 0, 0, locPreliminary),
				},
				{
					Name:          "reminder",
					Description:   "PDS reminder",
					ExecutionTime: time.Date(2022, 12, 07, 11, 0, 0, 0, locPreliminary),
					UserIDSkip: map[string]string{
						"10005": "10005",
						"10016": "10016",
						"10037": "10037",
						"10041": "10041",
						"10054": "10054",
						"10055": "10055",
						"10056": "10056",
						"10057": "10057",
						"10058": "10058",
						"10059": "10059",
						"10065": "10065",
						"10067": "10067",
						"10068": "10068",
						"10070": "10070",
						"10071": "10071",
						"10078": "10078",
						"10079": "10079",
						"10080": "10080",
						"10082": "10082",
						"10086": "10086",
						"10087": "10087",
					},
				},
			},
			"fmt": {
				{
					Name:          "invitation",
					Description:   "Montag",
					ExecutionTime: time.Date(2022, 11, 07, 11, 0, 0, 0, locPreliminary),
					// RelayHost: "email.zew.de",
					RelayHost: "zimbra.zew.de",
					URL: &UrlT{
						URL:  "http://fmt-2020.zew.local/fmt/individualbericht-curl.php?mode=invitation",
						TTL:  60 * 60, // deadline for new participants
						User: "pbu",
					},
					TemplateName: "invitation",
				},
				{
					Name:          "reminder",
					Description:   "Freitag",
					ExecutionTime: time.Date(2022, 11, 11, 11, 0, 0, 0, locPreliminary),
					URL: &UrlT{
						URL:  "http://fmt-2020.zew.local/fmt/individualbericht-curl.php?mode=reminder",
						TTL:  0, // reminders should not be stale
						User: "pbu",
					},
					TemplateName: "reminder",
				},
				{
					Name:          "results1a",
					Description:   "Dienstags um 11 - 270 Teilnehmer",
					ExecutionTime: time.Date(2022, 11, 15, 11, 0, 0, 0, locPreliminary),
					TemplateName:  "results1",
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
				},
				{
					Name:         "results1b",
					Description:  "Dienstags um 11 - Ergebnisverteiler - ca. 30 Interessenten FMT-Dt",
					TemplateName: "results1",
					SameAs:       "results1a",
				},
				{
					Name:          "results2a",
					Description:   "Finanzmarkt Report am Freitag - 270 teilnehmer",
					ExecutionTime: time.Date(2022, 11, 18, 10, 0, 0, 0, locPreliminary),
					Attachments: []AttachmentT{
						{
							Language: "de",
							// should be named next month
							// Label:    "ZEW-Finanzmarktreport-%v-%02v.pdf",
							Label:    "ZEW-Finanzmarktreport.pdf",
							Filename: "fmr/report.pdf",
						},
					},
				},
				{
					Name:          "results2b",
					Description:   "Finanzmarkt Report am Freitag - Ergebnisverteiler - ca. 30 Interessenten FMT-Dt",
					ExecutionTime: time.Date(2022, 11, 18, 10, 0, 0, 0, locPreliminary),
					Attachments: []AttachmentT{
						{
							Language: "de",
							// should be named next month
							// Label:    "ZEW-Finanzmarktreport-%v-%02v.pdf",
							Label:    "ZEW-Finanzmarktreport.pdf",
							Filename: "fmr/report.pdf",
						},
					},
				},
			},
		},
	}

	bts1, _ := json.MarshalIndent(example, "  ", "  ")
	os.WriteFile("example-config.json", bts1, 0777)

}

var cfg = configT{}
