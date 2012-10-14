package main

import (
	"flag"
	"fmt"
	"github.com/rdwilliamson/aws"
	"github.com/rdwilliamson/aws/glacier"
	"os"
	"runtime/pprof"
)

var (
	connection *glacier.Connection
	retrys     int
)

func main() {
	flag.IntVar(&retrys, "retrys", 3, "number of retrys when uploading multipart part")
	cpu := flag.String("cpuprofile", "", "cpu profile file")
	help := flag.Bool("help", false, "print usage")
	flag.Parse()

	if *help {
		fmt.Println(`glacier <region> archive upload <vault> <file> <description>
glacier <region> archive delete <vault> <archive>
glacier <region> job inventory <vault> <topic> <description>
glacier <region> job archive <vault> <archive> <topic> <description>
glacier <region> job list <vault>
glacier <region> job describe <vault> <job>
glacier <region> job get inventory <vault> <job>
glacier <region> job get archive <vault> <job> <file>
glacier <region> multipart init <vault> <file> <size> <description>
glacier <region> multipart print <file>
glacier <region> multipart run <file> <parts>
glacier <region> multipart abort <file>
glacier <region> multipart list parts <file>
glacier <region> vault create <name>
glacier <region> vault delete <name>
glacier <region> vault describe <name>
glacier <region> vault list
glacier <region> vault notifications set <name> <topic>
glacier <region> vault notifications get <name>
glacier <region> vault notifications delete <name>`)
		return
	}

	if *cpu != "" {
		f, err := os.Create(*cpu)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

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
