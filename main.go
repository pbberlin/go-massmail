package main

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/domodwyer/mailyak"
	"github.com/gocarina/gocsv"
	"github.com/jackpal/gateway"
)

type Recipient struct {
	ID          string `csv:"id"`
	Email       string `csv:"email"`
	Sex         int    `csv:"sex"`
	Title       string `csv:"title"`
	Lastname    string `csv:"lastname"`
	NoMail      string `csv:"!Mail !Call"` // value noMail
	SourceTable string `csv:"src_table"`   // either emtpy string or 'mailadresse', serves to control computation of derived fields

	Link     template.HTML `csv:"link"` // avoid escaping
	Language string        `csv:"lang"`

	SMTP string `json:"smtp,omitempty"` // effective smtp server sending

	Anrede                 string `csv:"anrede"`
	MonthYear              string `csv:"-"` // Oktober 2022, October 2022
	QuarterYear            string `csv:"-"` // Q1 2023
	ClosingDatePreliminary string `csv:"-"` // Friday, 11th November 2022   Freitag, den 11. November 2022,
	// two days later
	ClosingDateLastDue string `csv:"-"` // Monday, 14th November 2022   Freitag, den 14. November 2022,
	ExcelLink          string `csv:"-"`
}

// IP addresses need to be configurable
// map[string]bytes positive
// map[string]bytes negative
// 2023-03 - no longer used - see DomainsToRelayHorsts
func isInternalGateway() bool {

	ipGW, err := gateway.DiscoverGateway()
	if err != nil {
		log.Printf("discovering gateway yielded error %v", err)
		return false
	}

	guestGW := net.IPv4(192, 168, 178, 1)
	if ipGW.Equal(guestGW) {
		return false
	}

	membersGW1 := net.IPv4(192, 168, 50, 1)
	membersGW2 := net.IPv4(192, 168, 50, 2)
	if ipGW.Equal(membersGW1) || ipGW.Equal(membersGW2) {
		return true
	}

	vpnGW1 := net.IPv4(192, 168, 26, 175)
	if ipGW.Equal(vpnGW1) {
		return true
	}
	vpnGW2 := net.IPv4(192, 168, 199, 6)
	if ipGW.Equal(vpnGW2) {
		return true
	}
	// starbucks via VPN
	vpnGW3 := net.IPv4(10, 128, 128, 128)
	if ipGW.Equal(vpnGW3) {
		return true
	}

	internalGW := net.IPv4(10, 7, 10, 60)
	if ipGW.Equal(internalGW) {
		return true
	}

	log.Printf("cannot classify gateway %v", ipGW)
	os.Exit(0)

	return false
}

func formatDate(dt time.Time, lang string) string {

	if lang == "en" {
		ret := dt.Format("Monday, 2th January 2006")
		return ret
	}

	if lang == "de" {

		ret := dt.Format("Monday, 2. January 2006")

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
		return ret
	}

	// any other
	ret := dt.Format("Monday, 2th January 2006")
	return ret

}

// stackoverflow.com/questions/30376921
func fileCopy(in io.Reader, dst string) (err error) {

	// Already exists?
	if _, err := os.Stat(dst); err == nil {
		// return nil  // skip
	}
	err = nil

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("error creating destination file: %w", err)
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	//
	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("error io.Copy: %v", err)
	}

	err = out.Sync()
	return

}

// SetDerived computes helper fields from base columns
func (r *Recipient) SetDerived(wv *WaveT, tsk *TaskT) {

	if r.SourceTable == "" {

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

	} else if r.SourceTable == "mailadresse" {

		// this database table has no language column;
		// default language is 'de'.
		// we derive the language from the 'anrede'
		if strings.Contains(r.Anrede, "Dear") {
			r.Language = "en"
		}

	} else if r.SourceTable == "pds" {

		if !strings.Contains(r.Anrede, "Dear Sir or Madam") {
			r.Anrede = strings.TrimSpace(r.Anrede)
			r.Lastname = strings.TrimSpace(r.Lastname)
			r.Anrede = "Dear " + r.Anrede + " " + r.Lastname
		}
		r.Language = "en"

	} else if r.SourceTable == "pds-old" {
		r.NoMail = "noMail"

	}

	if r.ID != "" {
		if _, ok := tsk.UserIDSkip[r.ID]; ok {
			r.NoMail = "noMail"

		}
	}

	// survey identifier
	y := wv.Year
	m := wv.Month
	r.MonthYear = fmt.Sprintf("%v %v", MonthByInt(int(m), r.Language), y)

	quarter := int(m-1)/3 + 1
	r.QuarterYear = fmt.Sprintf("Q%v %v", quarter, y)

	// due dates
	prelimi := wv.ClosingDatePreliminary
	lastDue := wv.ClosingDateLastDue
	if false {
		prelimi = time.Date(2022, 11, 11+0, 17, 0, 0, 0, loc)
		lastDue = prelimi.AddDate(0, 0, 3)
	}
	r.ClosingDatePreliminary = formatDate(prelimi, r.Language)
	r.ClosingDateLastDue = formatDate(lastDue, r.Language)

	tenDaysPast := time.Now().Add(-10 * 24 * 3600 * time.Second)
	for _, t := range []time.Time{prelimi, lastDue} {
		if !t.IsZero() && tenDaysPast.After(t) {
			log.Fatalf("%v: ClosingDate* %v is older than %v", tsk.Name, formatDate(t, r.Language), formatDate(tenDaysPast, r.Language))
		}
	}

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

	ext := "md"
	if tsk.HTML {
		ext = "html"
	}

	fn := fmt.Sprintf("%v-%v.%v", templateFile, language, ext)
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

	relayHostKey := cfg.DefaultHorst
	if tsk.RelayHost != "" {
		relayHostKey = tsk.RelayHost
	}
	rh := cfg.RelayHorsts[relayHostKey]

	nameDomain := strings.Split(rec.Email, "@")
	if len(nameDomain) != 2 {
		err := fmt.Errorf("rec.Email seems malformed %v", rec.Email)
		return err
	}
	domain := "@" + nameDomain[1]

	// no longer used - see DomainsToRelayHorsts
	if key, ok := cfg.DomainsToRelayHorsts[domain]; ok {
		if true || isInternalGateway() {
			if _, ok := cfg.RelayHorsts[key]; ok {
				log.Printf("\trecipient domain %v via internal SMTP host %v", domain, key)
				rh = cfg.RelayHorsts[key]
				relayHostKey = key
			} else {
				err := fmt.Errorf("email domain %v points to SMTP host %v, which does not exist", domain, key)
				return err
			}
		} else {
			log.Printf("\trecipient domain %v - we are not internal", domain)
		}
	}

	m := mailyak.New(
		rh.HostNamePort, // "email.zew.de:587",
		rh.getAuth(),
	)

	if cfg.Projects[project].From.Address == "" {
		return fmt.Errorf("Task.From or Config.DefaultFrom email must be set")
	}
	m.From(cfg.Projects[project].From.Address)
	m.FromName(cfg.Projects[project].From.Name)

	// m.ReplyTo = m.From.Address
	if cfg.Projects[project].ReplyTo != "" {
		m.ReplyTo(cfg.Projects[project].ReplyTo)
	}
	if cfg.Projects[project].Bounce != "" {
		// return-path is a hidden email header
		// indicating where bounced emails will be processed.
		// m.AddCustomHeader("Return-Path", cfg.Projects[project].Bounce)
		m.AddHeader("Return-Path", cfg.Projects[project].Bounce)
	}
	m.AddHeader("List-Unsubscribe", fmt.Sprintf("<maito:%v>", cfg.Projects[project].ReplyTo))

	if rec.Email == "" || !strings.Contains(rec.Email, "@") {
		return fmt.Errorf("email field %q is suspect \n\t%+v", rec.Email, rec)
	}
	m.To(rec.Email)

	//
	// attachments
	attCtr := 0
	for _, att := range tsk.Attachments {

		if att.Language != rec.Language {
			continue
		}

		if filepath.Ext(att.Filename) != filepath.Ext(att.Label) {
			err := fmt.Errorf("file %v must have a label with matching extension", att.Filename)
			log.Print(err)
			return err
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

		fi, err := os.Stat(pth)
		if err != nil {
			log.Printf("error getting file info for %v\n\t%v", pth, err)
			return err
		}

		modTimePlus := fi.ModTime().Add(20 * 24 * 3600 * time.Second)
		if time.Now().After(modTimePlus) {
			err := fmt.Errorf("file over 20 days old: %v ", filepath.Base(pth))
			log.Print(err)
			return err
		}

		f, err := os.OpenFile(pth, os.O_RDONLY, 0x777)
		if err != nil {
			log.Printf("error doing attachment %+v\n\t%v", att, err)
			return err
		}
		m.Attach(lbl, f)
		attCtr++
	}

	m.AddHeader("X-Mailer", "go-massmail")

	rec.SMTP = rh.HostNamePort

	subj, bod := getText(rec, project, tsk, rec.Language)
	log.Printf("  subject:   %v", subj)
	m.Subject(subj)
	if tsk.HTML {
		// m.Plain().Set("Get a real email client")
		m.Plain().Set(bod)
		m.HTML().Set(bod)
	} else {
		m.Plain().Set(bod)
	}

	log.Printf("  sending %q via %s... to %v with %v attach(s)",
		mode, rh.HostNamePort, rec.Lastname, attCtr,
	)

	if strings.Contains(rec.NoMail, "noMail") {
		log.Printf("    skipping 'noMail'")
		return nil
	}

	if mode != "prod" {
		return nil
	}

	delayEff := rh.Delay
	if rh.Delay == 0 {
		rh.Delay = 200 // 200 milliseconds default
	}
	time.Sleep(time.Millisecond * time.Duration(delayEff))

	// err := gm.Send(
	// 	rh.HostNamePort,
	// 	rh.getAuth(),
	// 	m,
	// )
	err := m.Send()
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

			if tsk.ExecutionTime.IsZero() && tsk.ExecutionInterval == "" {
				log.Printf("\t%v-%-22v - neither exec time nor invertval; skipping", survey, tsk.Name)
				// log.Print(util.IndentedDump(tsk))
				continue
			} else if tsk.ExecutionTime.IsZero() && tsk.ExecutionInterval != "" {
				if tsk.ExecutionInterval == "daily" {
					surveys = append(surveys, survey)
					waves = append(waves, wv)
					tasks = append(tasks, tsk)
					fmt.Fprintf(msg, "\t%v-%-22v   %v\n", survey, tsk.Name, tsk.Description)
					continue
				}
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
				fmt.Fprintf(msg, "\t%v-%-26v   %v\n", survey, tsk.Name, tsk.Description)
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

func processTask(project string, wv WaveT, tsk TaskT) {

	log.Printf("\n\n\t%v-%-22v   %v - %v att(s)\n\t==================", project, tsk.Name, tsk.Description, len(tsk.Attachments))

	// CSV file containing participants
	fn := fmt.Sprintf("./csv/%v/%v.csv", project, tsk.Name)
	fnCopy := fmt.Sprintf("./csv/%v/%v-%d-%02d.csv", project, tsk.Name, wv.Year, wv.Month)
	log.Printf("using filename %v\n", fn)

	inFile, err := os.OpenFile(
		fn,
		os.O_RDWR,
		os.ModePerm,
	)
	if err != nil && !os.IsNotExist(err) {
		log.Print(err)
		return
	}

	// update of CSV required?
	conditionExist := !os.IsNotExist(err)
	conditionStale := false

	if conditionExist {
		// checking for stale - if file already exists
		stat, err := inFile.Stat()
		if err != nil {
			log.Printf("inFile Stat error: %v", err)
			return
		}
		ttl := time.Duration(0)
		if tsk.URL != nil {
			ttl = tsk.URL.TTL
		}
		stale := stat.ModTime().Add(ttl * time.Second) // ModTime => last downloaded
		if time.Now().After(stale) {
			hasWGet := "no wget URL"
			if tsk.URL != nil && tsk.URL.URL != "" {
				hasWGet = "trying wget"
			}
			log.Printf("      filename %v  is stale - %v", fn, hasWGet)
			conditionStale = true
		} else {
			log.Printf("      filename %v  is fresh", fn)
		}
	} else {
		log.Printf("      filename %v  not exists", fn)
		if tsk.URL == nil || tsk.URL.URL == "" {
			return
		}
	}

	if !conditionExist || conditionStale {
		if tsk.URL != nil && tsk.URL.URL != "" {
			log.Printf("downloading from %v\n", tsk.URL.URL)
			opts := WGetOpts{
				URL:     tsk.URL.URL,
				OutFile: fn,
				Verbose: true,
				User:    tsk.URL.User,
			}
			err := wget(opts, os.Stderr)
			if err != nil {
				log.Printf("wget error %v", err)
				return
			}
			inFile, err = os.OpenFile(
				fn,
				// os.O_RDWR|os.O_CREATE,
				os.O_RDWR,
				os.ModePerm,
			)
			if err != nil {
				log.Printf("error opening wgetted file: %v", err)
				return
			}
		}
	}

	defer inFile.Close()

	err = fileCopy(inFile, fnCopy)
	if err != nil {
		log.Printf("fileCopy error %v", err)
		return
	}
	_, err = inFile.Seek(0, 0) // after the copy operation above
	if err != nil {
		log.Printf("seek back to start error %v", err)
		return
	}

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

		lnTR := len(cfg.Projects[project].TestRecipients)
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
			for _, testEmail := range cfg.Projects[project].TestRecipients {
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

				log.Printf("    test lang %v using %-42v with %v", recs[idxes[i]].Language, recs[idxes[i]].Email, cfg.Projects[project].TestRecipients[i])
				lastIdx := len(subsetRec) - 1
				subsetRec[lastIdx].Email = cfg.Projects[project].TestRecipients[i]
			}

		}

		recs = subsetRec

	}

	log.Print("\n\t preflight")
	for idx1, rec := range recs {
		rec.SetDerived(&wv, &tsk)
		log.Printf(
			"#%03v %-28v %-28v - %v",
			idx1+1,
			// rec.MonthYear,
			rec.Email,
			rec.Anrede,
			rec.ClosingDatePreliminary,
		)
		err := singleEmail("test", project, *rec, wv, tsk)
		if err != nil {
			log.Printf("error in preflight run:\n\t%v", err)
			return
		}
	}

	const waitSeconds = 8
	bt1 := loopAsync(waitSeconds)
	if bt1 == 97 {
		log.Print("aborting...")
		os.Exit(0)
	} else if bt1 == 99 {
		log.Print("continue")
	} else if bt1 == 115 {
		log.Print("next task")
		return
	} else {
		log.Print("loopAsync returned invalid byte code; aborting...")
		os.Exit(0)
	}

	// back to start of file
	if _, err := inFile.Seek(0, 0); err != nil {
		log.Print(err)
		return
	}

	log.Print("\n\t prod")
	for idx1, rec := range recs {
		log.Printf(
			"#%03v %-28v %v  %v %v%v %v",

			idx1+1,

			rec.Email,
			rec.Anrede,

			rec.ClosingDatePreliminary,
			rec.Language, rec.Sex,
			rec.MonthYear,
		)
		err := singleEmail("prod", project, *rec, wv, tsk)
		if err != nil {
			log.Printf("error in prod run:\n\t%v", err)
			// log.Printf("\t%v", project)
			// log.Printf("\t%v", wv)
			// log.Printf("\t%v", tsk)
			// log.Printf("\t%v", *rec)
			return
		}

	}

}
