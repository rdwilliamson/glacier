package main

import (
	"flag"
	"fmt"
	"os"
)

func job() {
	switch flag.Arg(2) {
	case "archive":
	case "inventroy":
	default:
		fmt.Println("unknown multipart command:", flag.Arg(2))
		os.Exit(1)
	}
}
