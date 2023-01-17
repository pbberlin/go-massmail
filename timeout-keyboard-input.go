package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"
)

func keyboardTimeout(waitSeconds int) byte {

	fmt.Printf("\tcontinue in %v secs\n", waitSeconds)
	fmt.Print("\tc - continue task\n")
	fmt.Print("\ts - skip to next task\n")
	fmt.Print("\ta - abort (CTRL+C)\n")

	for i := 0; i < waitSeconds*5; i++ {

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
			var bt1 byte // extracting the first byte, default is 0
			if len(bts) > 0 {
				bt1 = bts[0]
			}
			outp := fmt.Sprintf("%02d: got text %v - byte %v - byte1 %v \n", i, txt, bts, bt1)
			_ = outp

			// a  -  bt1 ==  97   => abort
			// c  -  bt1 ==  99   => continue
			// s  -  bt1 == 115   => skip (next)
			if bt1 == 97 {
				log.Print("aborted")
				os.Exit(0)
			}
			if bt1 == 99 {
				// break continueSending
				return 99
			}
			if bt1 == 115 {
				log.Print("next task")
				return 115
			}
			// else - ignored

		}
		fmt.Print(".")
		time.Sleep(time.Second / 5)
	}
	fmt.Print("\n")

	return 0

}
