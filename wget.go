package main

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type WGetOpts struct {
	URL     string
	OutFile string
	Verbose bool
	User    string
}

func getHttpTransport(secureProtocol string) (http.RoundTripper, error) {
	minSecureProtocol := uint16(0)
	maxSecureProtocol := uint16(0)
	switch secureProtocol {
	case "auto":
		maxSecureProtocol = minSecureProtocol
	case "TLSv1":
		minSecureProtocol = tls.VersionTLS10
	case "":
		//OK
	default:
		return nil, fmt.Errorf("unrecognised secure protocol '%v'", secureProtocol)
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         minSecureProtocol,
			MaxVersion:         maxSecureProtocol,
		},
	}
	return tr, nil
}

func progress(perc int64) string {
	equalses := perc * 38 / 100
	if equalses < 0 {
		equalses = 0
	}
	spaces := 38 - equalses
	if spaces < 0 {
		spaces = 0
	}
	prog := strings.Repeat("=", int(equalses)) + ">" + strings.Repeat(" ", int(spaces))
	return prog
}

func wget(opts WGetOpts, errPipe io.Writer) error {

	opts.URL = strings.TrimSpace(opts.URL)
	opts.OutFile = strings.TrimSpace(opts.OutFile)

	if len(opts.URL) < 5 || len(opts.OutFile) < 2 {
		return fmt.Errorf("url or filename too short or empty")
	}

	startTime := time.Now()
	request, err := http.NewRequest("GET", opts.URL, nil)
	//resp, err := http.Get(link)
	if err != nil {
		return err
	}

	tr, err := getHttpTransport("")
	if err != nil {
		return err
	}
	client := &http.Client{Transport: tr}

	if opts.Verbose {
		for headerName, headerValue := range request.Header {
			fmt.Fprintf(errPipe, "Request header %s: %s\n", headerName, headerValue)
		}
	}

	// http base64 authentication
	if opts.User != "" {
		// stackoverflow.com/questions/16673766
		env := fmt.Sprintf("PW_%v", strings.ToUpper(opts.User))
		pass := os.Getenv(env)
		if pass == "" {
			return fmt.Errorf("http base64 user %q needs environment password %v", opts.User, env)
		}
		auth := opts.User + ":" + pass
		authEnc := base64.StdEncoding.EncodeToString([]byte(auth))
		request.Header.Add("Authorization", "Basic "+authEnc)
	}

	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Fprintf(errPipe, "Http response status: %s\n", resp.Status)
	if opts.Verbose {
		for headerName, headerValue := range resp.Header {
			fmt.Fprintf(errPipe, "Response header %s: %s\n", headerName, headerValue)
		}
	}

	lenS := resp.Header.Get("Content-Length")
	length := int64(-1)
	if lenS != "" {
		length, err = strconv.ParseInt(lenS, 10, 32)
		if err != nil {
			return err
		}
	}

	typ := resp.Header.Get("Content-Type")
	fmt.Fprintf(errPipe, "Content-Length: %v Content-Type: %s\n", lenS, typ)

	contentRange := resp.Header.Get("Content-Range")
	rangeEffective := false
	if contentRange != "" {
		//TODO parse it?
		rangeEffective = true
	}
	_ = rangeEffective

	//
	//
	var out io.Writer
	var outFile *os.File
	fmt.Fprintf(errPipe, "Saving to: '%v'\n\n", opts.OutFile)
	openFlags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	outFile, err = os.OpenFile(opts.OutFile, openFlags, fs.ModePerm)
	if err != nil {
		return err
	}
	defer outFile.Close()
	out = outFile

	//
	//
	buf := make([]byte, 4068)
	tot := int64(0)
	i := 0
	for {

		// read a chunk
		n, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		tot += int64(n)

		// write a chunk
		if _, err := out.Write(buf[:n]); err != nil {
			return err
		}
		i += 1
		if length > -1 {
			if length < 1 {
				fmt.Fprintf(errPipe, "\r     [ <=>                                  ] %d\t-.--KB/s eta ?s             ", tot)
			} else {
				//show percentage
				perc := (100 * tot) / length
				prog := progress(perc)
				nowTime := time.Now()
				totTime := nowTime.Sub(startTime)
				spd := float64(tot/1000) / totTime.Seconds()
				remKb := float64(length-tot) / float64(1000)
				eta := remKb / spd
				fmt.Fprintf(errPipe, "\r%3d%% [%s] %d\t%0.2fKB/s eta %0.1fs             ", perc, prog, tot, spd, eta)
			}
		} else {
			//show dots
			if math.Mod(float64(i), 20) == 0 {
				fmt.Fprint(errPipe, ".")
			}
		}
	}

	nowTime := time.Now()
	totTime := nowTime.Sub(startTime)
	spd := float64(tot/1000) / totTime.Seconds()
	if length < 1 {
		fmt.Fprintf(errPipe, "\r     [ <=>                                  ] %d\t-.--KB/s in %0.1fs             ", tot, totTime.Seconds())
		fmt.Fprintf(errPipe, "\n (%0.2fKB/s) - '%v' saved [%v]\n", spd, opts.OutFile, tot)
	} else {
		perc := (100 * tot) / length
		prog := progress(perc)
		fmt.Fprintf(errPipe, "\r%3d%% [%s] %d\t%0.2fKB/s in %0.1fs             ", perc, prog, tot, spd, totTime.Seconds())
		fmt.Fprintf(errPipe, "\n '%v' saved [%v/%v]\n", opts.OutFile, tot, length)
	}
	if err != nil {
		return err
	}
	if outFile != nil {
		err = outFile.Close()
	}
	return err
}
