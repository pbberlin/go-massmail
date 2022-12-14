package main

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	gm "github.com/zew/go-mail"
)

type Recipient struct {
	Email       string `csv:"email"`
	Sex         int    `csv:"sex"`
	Title       string `csv:"title"`
	Lastname    string `csv:"lastname"`
	NoMail      string `csv:"!Mail !Call"`
	SourceTable string `csv:"src_table"`

	Link     template.HTML `csv:"link"` // avoid escaping
	Language string        `csv:"lang"`

	Anrede    string `csv:"anrede"`
	MonthYear string `csv:"-"` // Oktober 2022, October 2022

	ClosingDatePreliminary string `csv:"-"` // Friday, 11th November 2022   Freitag, den 11. November 2022,
	// two days later
	ClosingDateLastDue string `csv:"-"` // Monday, 14th November 2022   Freitag, den 14. November 2022,
	ExcelLink          string `csv:"-"`
}

func formatDate(dt time.Time, lang string) string {

	ret := dt.Format("Monday, 2. January 2006")

	if lang == "de" {

		m := dt.Month()
		w := dt.Weekday()
		ret =
			strings.Replace(
				ret,
				MonthByInt(int(m), "en"),
				MonthByInt(int(m), "de"),
				-1,
			)

		ret =
			strings.Replace(
				ret,
				WeekdayByInt(int(w), "en"),
				WeekdayByInt(int(w), "de"),
				-1,
			)

		ret += "," // add apposition comma
	}

	return ret
}

func (r *Recipient) SetDerived(wv WaveT) {

	if r.SourceTable == "" {

		// => r.SourceTable NOT EQUAL 'mailadresse'

		if r.Language == "de" {
			if r.Sex == 1 {
				r.Anrede = "Sehr geehrter Herr "
			}
			if r.Sex == 2 {
				r.Anrede = "Sehr geehrte Frau "
			}
			if r.Title != "" {
				r.Anrede += r.Title + " "
			}
			r.Anrede += r.Lastname
		}
		if r.Language == "en" {
			if r.Sex == 1 {
				r.Anrede = "Dear Mr. "
			}
			if r.Sex == 2 {
				r.Anrede = "Dear Ms. "
			}
			if r.Title != "" {
				r.Anrede = "Dear " + r.Title + " "
			}
			r.Anrede += r.Lastname
		}

	}

	// survey identifier
	y := wv.Year
	m := wv.Month
	r.MonthYear = fmt.Sprintf("%v %v", MonthByInt(int(m), r.Language), y)

	// due dates
	prelimi := wv.ClosingDatePreliminary
	lastDue := wv.ClosingDateLastDue
	if false {
		prelimi = time.Date(2022, 11, 11+0, 17, 0, 0, 0, loc)
		lastDue = prelimi.AddDate(0, 0, 3)
	}
	r.ClosingDatePreliminary = formatDate(prelimi, r.Language)
	r.ClosingDateLastDue = formatDate(lastDue, r.Language)

	publication := lastDue.AddDate(0, 0, 1)

	r.ExcelLink = fmt.Sprintf(
		`https://fmtdownload.zew.de/fdl/download/public/%v-%02d-%02d_1100/tab.xlsx`,
		publication.Year(),
		int(publication.Month()),
		int(publication.Day()),
	)

}

func getText(rec Recipient, project string, tsk TaskT, language string) (subject, body string) {

	templateFile := tsk.Name
	// check for explicitly different email template
	if tsk.TemplateName != "" {
		templateFile = tsk.TemplateName
	}
	fn := fmt.Sprintf("%v-%v.md", templateFile, language)
	pth := filepath.Join(".", "tpl", project, fn)
	t, err := template.ParseFiles(pth)
	if err != nil {
		log.Fatalf("could not parse template %v\n\t%v", fn, err)
	}

	// log.Printf("template parse success %v", t.Name())

	sb := &strings.Builder{}
	err = t.ExecuteTemplate(sb, fn, rec)
	if err != nil {
		log.Fatalf("could not execute template %v\n\t%v", fn, err)
	}

	if strings.Contains(sb.String(), "\r\n") {
		log.Fatalf("template %v contains \"r\"n - should be only \"n", t.Name())
	}

	lines := strings.Split(sb.String(), "\n")

	return lines[0], strings.Join(lines[1:], "\n")
}

// project - fmt, pds, difi
// task - invitation, reminder
func singleEmail(mode, project string, rec Recipient, wv WaveT, tsk TaskT) error {

	if mode != "prod" && mode != "test" {
		return fmt.Errorf("singleEmail mode must be 'prod' or 'test'; is %v", mode)
	}

	m := gm.NewMessagePlain(getText(rec, project, tsk, rec.Language))
	// 	m = gm.NewMessageHTML(getSubject(subject, relayHorst.HostNamePort), getBody(senderHorst, true))
	log.Printf("  recipient: %v", rec.Email)
	log.Printf("  subject:   %v", m.Subject)
	// log.Print(m.Body)
	// return

	m.From = mail.Address{}
	m.From.Name = "Finanzmarkttest"
	m.From.Address = "noreply@zew.de"
	m.To = []string{rec.Email}

	if rec.Email == "" || !strings.Contains(rec.Email, "@") {
		return fmt.Errorf("singleEmail email %v is suspect", rec.Email)
	}

	m.ReplyTo = "finanzmarkttest@zew.de"
	// return-path is a hidden email header
	// indicating where bounced emails will be processed.
	m.AddCustomHeader("Return-Path", m.ReplyTo)

	//
	// attachments
	for _, att := range tsk.Attachments {
		if att.Language != rec.Language {
			continue
		}

		lbl := att.Label
		// xx = tpl.ParseFiles("")
		// if attachment label contains placeholders, replace with wave data
		if strings.Contains(lbl, "%v") {
			lbl = fmt.Sprintf(lbl, wv.Year, int(wv.Month))
		}

		pth := filepath.Join(".", "attachments", project, att.Filename)
		if cfg.AttachmentRoot != "" {
			pth = filepath.Join(cfg.AttachmentRoot, project, att.Filename)
		}
		if err := m.Attach(lbl, pth); err != nil {
			log.Printf("problem with attachment %+v\n\t%v", att, err)
			return err
		}
	}

	m.AddCustomHeader("X-Mailer", "go-mail")

	relayHostKey := cfg.DefaultHorst
	if tsk.RelayHost != "" {
		relayHostKey = tsk.RelayHost
	}
	rh := cfg.RelayHorsts[relayHostKey]

	log.Printf("  sending %q via %s... to %v with %v attach",
		mode, rh.HostNamePort, rec.Lastname, len(m.Attachments),
	)
	if mode != "prod" {
		return nil
	}

	err := gm.Send(
		rh.HostNamePort,
		rh.getAuth(),
		m,
	)
	if err != nil {
		return fmt.Errorf(" error sending lib-email  %v:\n\t%w", relayHostKey, err)
	} else {
		// log.Printf("  lib-email sent")
		return nil
	}

}

// inBetween if start < t < start+24h
func inBetween(desc string, t, start, stop time.Time) bool {

	//     t > start"
	due := t.After(start)

	//     t < stpp
	fresh := stop.After(t)

	if due && fresh {
		log.Printf(
			"%6v: %v < %v < %v",
			desc,
			start.Format("2006-01-02--15:04"),
			t.Format("2006-01-02--15:04"),
			stop.Format("2006-01-02--15:04"),
		)

		return true
	}

	return false

}

// dueTasks searches the config and returns due tasks.
// test runs are executed 24 hours before in advance
func dueTasks() (surveys []string, waves []WaveT, tasks []TaskT) {

	msg := &strings.Builder{}

	now := time.Now()

	for survey, wvs := range cfg.Waves {
		last := len(wvs) - 1
		wv := wvs[last]
		for _, tsk := range cfg.Tasks[survey] {

			if tsk.ExecutionTime.IsZero() {
				log.Printf("\t%v-%-22v has no exec time; skipping", survey, tsk.Name)
				// log.Print(util.IndentedDump(tsk))
				continue
			}

			if inBetween("prod", now, tsk.ExecutionTime, tsk.ExecutionTime.AddDate(0, 0, 1)) {
				surveys = append(surveys, survey)
				waves = append(waves, wv)
				tasks = append(tasks, tsk)
				fmt.Fprintf(msg, "\t%v-%-22v   %v\n", survey, tsk.Name, tsk.Description)
			}

			dayBefore := tsk.ExecutionTime.AddDate(0, 0, -1) //  advance for testing
			if inBetween("advance", now, dayBefore, dayBefore.AddDate(0, 0, 1)) {
				surveys = append(surveys, survey)
				waves = append(waves, wv)
				tsk.testmode = true
				tasks = append(tasks, tsk)
				fmt.Fprintf(msg, "\t%v-%-22v   %v\n", survey, tsk.Name, tsk.Description)
			}
		}

	}

	if len(surveys) > 0 {
		log.Printf("%02v due tasks found:\n%v\n", len(surveys), msg)
	} else {
		log.Printf("no due task(s) found")
	}

	return
}

func iterTasks() {

	surveys, waves, tasks := dueTasks()
	for idx, survey := range surveys {
		processTask(survey, waves[idx], tasks[idx])
	}

}

func processTask(survey string, wv WaveT, tsk TaskT) {

	log.Printf("\n\n\t%v-%-22v   %v\n\t==================", survey, tsk.Name, tsk.Description)

	participantFile := tsk.Name
	fn := fmt.Sprintf("./csv/%v-%v.csv", survey, participantFile)
	log.Printf("using filename %v\n", fn)

	inFile, err := os.OpenFile(
		fn,
		// os.O_RDWR|os.O_CREATE,
		os.O_RDWR,
		os.ModePerm,
	)
	if err != nil {
		log.Print(err)
		return
	}
	defer inFile.Close()

	recs := []*Recipient{} // recipients

	// set option for gocsv lib
	// use semicolon as delimiter
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = ';'
		// r.LazyQuotes = true
		// r.TrimLeadingSpace = true
		return r
	})

	if err := gocsv.UnmarshalFile(inFile, &recs); err != nil {
		log.Print(err)
		return
	}

	if operationMode != "prod" || tsk.testmode {

		lnTR := len(cfg.TestRecipients)
		if lnTR < 1 {
			log.Printf("TestRecipients must be set in config")
			return
		}

		// langs is a map of languages containing a list of recipient IDs.
		// For each TestRecipient email and each language, we want to find
		// a recipient record to use as test
		// For example
		// 		de: [120  0]
		// 		en: [ 68 82]
		langs := map[string][]int{}

		//
		// add recipients with email _similar_ to any in TestRecipients
		for i1 := 0; i1 < len(recs); i1++ {
			for _, testEmail := range cfg.TestRecipients {
				if testEmail == recs[i1].Email {
					langs[recs[i1].Language] = append(langs[recs[i1].Language], i1)
				}
			}
		}

		//
		// all distinct languages and their first, second, ...  occurrence
		for i := 0; i < len(recs); i++ {
			if len(langs[recs[i].Language]) < lnTR {
				langs[recs[i].Language] = append(langs[recs[i].Language], i)
			}
		}
		log.Printf("  distinct languages - recipients at %v", langs)

		subsetRec := []*Recipient{}

		for _, idxes := range langs {

			for i := 0; i < len(idxes); i++ {

				subsetRec = append(subsetRec, recs[idxes[i]])

				log.Printf("    test %v using %-32v with %v", recs[idxes[i]].Language, recs[idxes[i]].Email, cfg.TestRecipients[i])
				lastIdx := len(subsetRec) - 1
				subsetRec[lastIdx].Email = cfg.TestRecipients[i]
			}

		}

		recs = subsetRec

	}

	log.Print("\n\t preflight")
	for idx1, rec := range recs {
		rec.SetDerived(wv)
		// if idx1 > 5 || idx1 < len(recipients)-5 {
		// 	continue
		// }
		log.Printf(
			"#%03v - %12v  %26v  %v",
			idx1+1,
			rec.MonthYear,
			rec.ClosingDatePreliminary,
			rec.Anrede,
		)
		err := singleEmail("test", survey, *rec, wv, tsk)
		if err != nil {
			log.Printf("error in preflight run:\n\t%v", err)
			return
		}
	}

	var inpChr []byte = make([]byte, 1)
	_ = inpChr

	fmt.Print("\tcontinue in 4 secs - cancel with CTRL+C\n\t")
	for i := 0; i < 4*5; i++ {
		// os.Stdin.Read(inpChr)
		// if inpChr[0] == 66 { // b
		// 	break
		// }
		fmt.Print(".")
		time.Sleep(time.Second / 4)
	}
	fmt.Print("\n")

	// back to start of file
	if _, err := inFile.Seek(0, 0); err != nil {
		log.Print(err)
		return
	}

	log.Print("\n\t prod")
	for idx1, rec := range recs {
		log.Printf("#%03v - %2v - %1v - %10v %-16v - %-32v ",
			idx1+1,
			rec.Language, rec.Sex,
			rec.Title, rec.Lastname,
			rec.Email,
		)
		if strings.Contains(rec.NoMail, "noMail") {
			log.Printf("  skipping 'noMail'")
			continue
		}
		if true {
			err := singleEmail("prod", survey, *rec, wv, tsk)
			if err != nil {
				log.Printf("error in prod run:\n\t%v", err)
				return
			}
		}
		time.Sleep(time.Second / 5)
	}

}
