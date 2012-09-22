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
	command    string
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
	args := flag.Args()
	if len(args) < 1 {
		// TODO print usage
		fmt.Println("no region argument")
		os.Exit(1)
	}

	var region *aws.Region
	for _, v := range aws.Regions {
		if v.Name == args[0] {
			region = v
			break
		}
	}
	if region == nil {
		fmt.Println("could not find region:", args[0])
		os.Exit(1)
	}
	args = args[1:]
	connection = glacier.NewConnection(secret, access, region)

	if len(args) < 1 {
		fmt.Println("no command argument")
		os.Exit(1)
	}
	switch args[0] {
	case "vault":
		vault(args[1:])
	case "archive":
		archive(args[1:])
	case "multipart":
		multipart(args[1:])
	case "job":
		job(args[1:])
	default:
		fmt.Println("unknown command:", flag.Arg(1))
	}
}
