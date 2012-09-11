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
	case "list":
		_, vaults, err := connection.ListVaults(0, "")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("%+v\n", vaults)
	default:
		fmt.Println("unknown vault command:", flag.Arg(2))
	}
}
