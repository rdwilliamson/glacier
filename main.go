package main

import (
	"flag"
	"fmt"
	"github.com/rdwilliamson/aws"
	// "github.com/rdwilliamson/aws/glacier"
	"../aws/glacier"
)

// $ glacier us-east-1 vault create <name>
// $ glacier us-east-1 vault delete <name>
// $ glacier us-east-1 vault describe <name>
// $ glacier us-east-1 vault list
// $ glacier us-east-1 vault notification set <name> [options]
// $ glacier us-east-1 vault notification get <name>
// $ glacier us-east-1 vault notification delete <name>

var (
	connection *glacier.Connection
)

func main() {
	flag.Parse()

	// get keys
	// TODO other ways to supply them
	secret, access := aws.KeysFromEnviroment()
	if secret == "" || access == "" {
		fmt.Println("could not get keys")
		return
	}

	// connection to region
	if flag.NArg() < 1 {
		fmt.Println("no region argument")
		return
	}
	var region *aws.Region
	for _, v := range aws.Regions {
		if v.Name == flag.Arg(0) {
			region = v
			break
		}
	}
	if region == nil {
		fmt.Println("could not find region:", flag.Arg(0))
		return
	}
	connection = glacier.NewConnection(secret, access, region)

	if flag.NArg() < 2 {
		fmt.Println("no command argument")
	}
	switch flag.Arg(1) {
	case "vault":
		vault()
	default:
		fmt.Println("unknown command:", flag.Arg(1))
	}
}

func vault() {
	switch flag.Arg(2) {
	case "list":
		_, vaults, err := connection.ListVaults(0, "")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("%+v\n", vaults)
	default:
		fmt.Println("unknown vault command:", flag.Arg(2))
	}
}
