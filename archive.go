package main

import (
	"flag"
	"fmt"
	"os"
)

func archive() {
	switch flag.Arg(2) {
	case "upload":
		file, err := os.Open(flag.Arg(4))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		location, err := connection.UploadArchive(flag.Arg(4), file,
			flag.Arg(3))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(location)
	case "delete":
	default:
		fmt.Println("unknown archive command:", flag.Arg(2))
	}
}
