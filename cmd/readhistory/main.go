package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ecc1/medtronic"
	"github.com/ecc1/nightscout"
)

var (
	verbose = flag.Bool("v", false, "print record details")
	model   = flag.Int("m", 523, "pump model")
	nsFlag  = flag.Bool("t", false, "format as Nightscout treatments")

	timeBlank = strings.Repeat(" ", len(medtronic.UserTimeLayout))
)

func main() {
	flag.Parse()
	family := medtronic.Family(*model % 100)
	for _, file := range flag.Args() {
		f, err := os.Open(file)
		if err != nil {
			log.Fatal(err)
		}
		data := readBytes(f)
		_ = f.Close()
		records, err := medtronic.DecodeHistory(data, family)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			*verbose = true
		}
		if *verbose {
			fmt.Println(nightscout.JSON(records))
		} else if *nsFlag {
			medtronic.ReverseHistory(records)
			fmt.Println(nightscout.JSON(medtronic.Treatments(records)))
		} else {
			for _, r := range records {
				printRecord(r)
			}
		}
	}
}

func readBytes(r io.Reader) []byte {
	var data []byte
	s := ""
	for {
		n, err := fmt.Fscan(r, &s)
		if n == 0 {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		b, err := strconv.ParseUint(s, 16, 8)
		if err != nil {
			log.Fatal(err)
		}
		data = append(data, byte(b))
	}
	return data
}

func printRecord(r medtronic.HistoryRecord) {
	t := time.Time(r.Time)
	tStr := timeBlank
	if !t.IsZero() {
		tStr = t.Format(medtronic.UserTimeLayout)
	}
	fmt.Printf("%s %v\n", tStr, r.Type())
}
