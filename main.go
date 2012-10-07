package main

import (
	"flag"
	"fmt"
	"github.com/rdwilliamson/aws"
	// "github.com/rdwilliamson/aws/glacier"
	"../aws/glacier"
	"os"
)

var (
	connection *glacier.Connection
)

func main() {
	flag.Parse()
	// TODO print usage

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
	command := args[0]
	args = args[1:]

	switch command {
	case "vault":
		vault(args)
	case "archive":
		archive(args)
	case "multipart":
		multipart(args)
	case "job":
		job(args)
	default:
		fmt.Println("unknown command:", command)
	}
}
