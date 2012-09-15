package main

import (
	"flag"
	"fmt"
	"github.com/rdwilliamson/aws"
	// "github.com/rdwilliamson/aws/glacier"
	"../aws/glacier"
	"os"
)

// $ glacier us-east-1 (vault|archive|etc)

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
		os.Exit(1)
	}

	// connection to region
	if flag.NArg() < 1 {
		// TODO print usage
		fmt.Println("no region argument")
		os.Exit(1)
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
		os.Exit(1)
	}
	connection = glacier.NewConnection(secret, access, region)

	if flag.NArg() < 2 {
		fmt.Println("no command argument")
	}
	switch flag.Arg(1) {
	case "vault":
		vault()
	case "multipart":
		multipart()
	default:
		fmt.Println("unknown command:", flag.Arg(1))
	}
}
