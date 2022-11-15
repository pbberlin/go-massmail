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

func getText(rec Recipient, project, task, language string) (subject, body string) {

	fn := fmt.Sprintf("%v-%v.md", task, language)
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
func singleEmail(mode, project string, rec Recipient, wv WaveT, task TaskT) error {

	if mode != "prod" && mode != "dry" {
		return fmt.Errorf("singleEmail mode must be 'prod' or 'dry'; is %v", mode)
	}

	m := gm.NewMessagePlain(getText(rec, project, task.Name, rec.Language))
	// 	m = gm.NewMessageHTML(getSubject(subject, relayHorst.HostNamePort), getBody(senderHorst, true))
	log.Printf("  subject: %v", m.Subject)
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
	for _, att := range task.Attachments {
		if att.Language != rec.Language {
			continue
		}

		lbl := att.Label
		// xx = tpl.ParseFiles("")
		yr := wv.Year
		mnth := int(wv.Month)
		if strings.Contains(lbl, "%v") {
			lbl = fmt.Sprintf(lbl, yr, mnth)
		}

		pth := filepath.Join(".", "attachments", project, att.Filename)
		if err := m.Attach(lbl, pth); err != nil {
			log.Printf("problem with attachment %+v\n\t%v", att, err)
			return err
		}
	}

	m.AddCustomHeader("X-Mailer", "go-mail")

	relayHost := cfg.RelayHorsts[task.RelayHost]

	log.Printf("  sending %q via %s... to %v with %v attach",
		mode, relayHost.HostNamePort, rec.Lastname, len(m.Attachments),
	)
	if mode != "prod" {
		return nil
	}

	err := gm.Send(
		relayHost.HostNamePort,
		relayHost.getAuth(),
		m,
	)
	if err != nil {
		return fmt.Errorf(" error sending lib-email  %v:\n\t%w", relayHost, err)
	} else {
		// log.Printf("  lib-email sent")
		return nil
	}

}

func getProjectTask() (string, WaveT, TaskT) {

	nw := time.Now()
	nwYr := nw.Year()
	nwMt := nw.Month()

	for projKey, waves := range cfg.Waves {
		for _, wv := range waves {
			if wv.Year == nwYr && wv.Month == nwMt {
				last := len(cfg.Tasks[projKey]) - 1
				tsk := cfg.Tasks[projKey][last]
				return projKey, wv, tsk
			}
		}

	}

	return "", WaveT{}, TaskT{}
}

func ProcessCSV() {

	project, wave, task := getProjectTask()

	fn := fmt.Sprintf("./csv/%v-%v.csv", project, task.Name)
	if operationMode != "prod" {
		fn = fmt.Sprintf("./csv/%v-%v.csv", "testproject", "testtask")
	}
	log.Printf("using filename %v\n", fn)

	inFile, err := os.OpenFile(
		fn,
		os.O_RDWR|os.O_CREATE,
		os.ModePerm,
	)
	if err != nil {
		log.Print(err)
		return
	}
	defer inFile.Close()

	recipients := []*Recipient{}

	// set option for gocsv lib
	// use semicolon as delimiter
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = ';'
		// r.LazyQuotes = true
		// r.TrimLeadingSpace = true
		return r
	})

	if err := gocsv.UnmarshalFile(inFile, &recipients); err != nil {
		log.Print(err)
		return
	}

	log.Print("\ndry")
	for idx1, rec := range recipients {
		rec.SetDerived(wave)
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
		err := singleEmail("dry", project, *rec, wave, task)
		if err != nil {
			log.Printf("error in dry run:\n\t%v", err)
			return
		}
	}

	fmt.Print("\n\n\tcontinue in 4 secs - cancel with CTRL+C\n\t")
	for i := 0; i < 4*5; i++ {
		fmt.Print(".")
		time.Sleep(time.Second / 4)
	}
	fmt.Print("\n")

	// back to start of file
	if _, err := inFile.Seek(0, 0); err != nil {
		log.Print(err)
		return
	}

	log.Print("\nprod")
	for idx1, rec := range recipients {
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
		if false {
			err := singleEmail("prod", project, *rec, wave, task)
			if err != nil {
				log.Printf("error in prod run:\n\t%v", err)
				return
			}
		}
		time.Sleep(time.Second / 5)
	}

}
