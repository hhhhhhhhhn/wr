package main

import (
	"flag"
	"os"
)

type flags struct {
	file string
}

func getFlags() (f flags) {
	help := flag.Bool("help", false, "print help")
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if len(flag.Args()) > 0 {
		f.file = flag.Args()[0]
	} else {
		f.file = "wr.txt"
	}

	return f
}
