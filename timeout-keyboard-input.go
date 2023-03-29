package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"
)

func keyboardInput(ret chan byte) {

	fmt.Print("\tc - execute current task - now\n")
	fmt.Print("\ts - skip to next task\n")
	fmt.Print("\ta - abort (CTRL+C)\n")
	fmt.Print("\t       press ENTER to submit\n")

	// probem with scanner - [Enter] is required to make it work. Probably the buffering of the stdin must be disabled for Linux and Windows distinctively
	scanner := bufio.NewScanner(os.Stdin)
	// scanner.Split(bufio.ScanBytes) // [character] and [Enter] are scanned separately, but [Enter] is still required
	// scanner := bufio.ScanBytes()   // same behavior as with Split(bufio.ScanBytes)
	// scanner.Buffer(make([]byte, 2), 2)  // only leads to overflow

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			log.Fatalf("error reading os.Stdin: %v", err)
		}
		// we get a line including [Enter]
		txt := scanner.Text()
		bts := scanner.Bytes()
		var bt1 byte // extracting first byte, default is 0
		if len(bts) > 0 {
			bt1 = bts[0]
		}
		outp := fmt.Sprintf("got text %v - byte %v - byte1 %v \n", txt, bts, bt1)
		_ = outp

		// a  -  bt1 ==  97   => abort
		// c  -  bt1 ==  99   => continue
		// s  -  bt1 == 115   => skip (next)
		if bt1 == 97 {
			log.Print("aborted")
			ret <- 97
			// os.Exit(0)
			return
		}
		if bt1 == 99 {
			// break continueSending
			// return 99
			log.Print("continue current task now")
			ret <- 99
			return
		}
		if bt1 == 115 {
			log.Print("next task")
			// return 115
			ret <- 115
			return
		}
		// else - any other byte code ignored

		//
		// try again...
	}

}

// loopSync cannot be combined with keyboard inut
func loopSync(waitSeconds int) {
	fmt.Printf("\tcontinue in %v secs\n", waitSeconds)
	for i := 0; i < waitSeconds*5; i++ {
		fmt.Print(".")
		time.Sleep(time.Second / 5)
	}
	fmt.Print("\n")
}

// loopAsync combines waiting with keyboard input.
// Return values are
//
//	'a' =  97   => abort
//	'c' =  99   => continue
//	's' = 115   => skip (next)
func loopAsync(waitSeconds int) byte {

	// channel 1
	keyInput := make(chan byte)
	go keyboardInput(keyInput)

	// channel 2
	deadline := time.Now().Add(time.Duration(waitSeconds * int(time.Second)))
	ticker := time.NewTicker(2 * time.Second) // make it slow - so the printing of "." does not conflict with human input

	//
	// combining channel 1 and 2
	fmt.Printf("\tcontinue in %v secs\n", waitSeconds)
	for {
		select {
		case tick := <-ticker.C:
			fmt.Print(".")
			if tick.After(deadline) {
				ticker.Stop()
				// close(ticker.C)
				fmt.Print("\n")
				// break mark1
				return 99
			}
		case key := <-keyInput:
			ticker.Stop()
			fmt.Print("\n")
			return key
		}
	}

}
