package main

import (
	"flag"
	"fmt"
	"os"
)

// $ glacier us-east-1 vault create <name>
// $ glacier us-east-1 vault delete <name>
// $ glacier us-east-1 vault describe <name>
// $ glacier us-east-1 vault list
// $ glacier us-east-1 vault notification set <name> [options]
// $ glacier us-east-1 vault notification get <name>
// $ glacier us-east-1 vault notification delete <name>

func vault() {
	switch flag.Arg(2) {
	case "create":
		err := connection.CreateVault(flag.Arg(3))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "delete":
		err := connection.DeleteVault(flag.Arg(3))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "describe":
		vault, err := connection.DescribeVault(flag.Arg(3))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("%+v\n", vault)
	case "list":
		_, vaults, err := connection.ListVaults(0, "")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("%+v\n", vaults)
	case "notification":
		if flag.NArg() < 5 {
			fmt.Println("no notification command")
			os.Exit(1)
		}
		switch flag.Arg(3) {
		case "set":
		case "get":
		case "delete":
		default:
			fmt.Println("unknown notification command:", flag.Arg(3))
			os.Exit(1)
		}
	default:
		fmt.Println("unknown vault command:", flag.Arg(2))
		os.Exit(1)
	}
}
