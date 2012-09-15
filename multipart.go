package main

import (
	"../aws/glacier"
	"flag"
	"fmt"
	"os"
)

func multipart() {
	switch flag.Arg(2) {
	case "create":
		in, err := os.Open(flag.Arg(3))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer in.Close()

		out, err := os.Create(flag.Arg(4))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer out.Close()

		err = glacier.CreateTreeHash(in, out)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "read":
		in, err := os.Open(flag.Arg(3))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer in.Close()

		err = glacier.ReadTreeHash(in)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		fmt.Println("unknown multipart command:", flag.Arg(2))
		os.Exit(1)
	}
}
