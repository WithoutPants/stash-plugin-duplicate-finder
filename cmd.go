package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rivo/duplo"
)

func cmdMain() {
	// default is to accept sprite directory and output csv of all matches
	path := os.Args[1]

	fmt.Fprintln(os.Stderr, "Outputting duplicates to csv")

	f, err := os.Create("duplicates.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// output to stdout
	a := api{}
	a.cfg.DBFilename = "df-hashstore.db"
	hdFunc := func(checksum string, matches duplo.Matches) {
		if len(matches) > 0 {
			match := matches[0]
			fmt.Fprintf(f, "%s,%s,%.f\n", checksum, match.ID.(string), -match.Score)
			fmt.Printf("%s - %s [%.f]\n", checksum, match.ID.(string), -match.Score)
		}
	}

	c := make(chan bool, 1)

	go func() {
		err = a.processFiles(path, hdFunc)
		if err != nil {
			panic(err)
		}
		c <- true
	}()

	for {
		select {
		case <-c:
			return
		default:
			_, err := os.Stat(".stop")
			if err == nil {
				a.stopping = true
			}
			time.Sleep(5 * time.Second)
		}
	}
}
