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
	retries    int
)

func main() {
	flag.IntVar(&retries, "retries", 3, "number of retries when uploading multipart part")
	cpu := flag.String("cpuprofile", "", "cpu profile file")
	help := flag.Bool("help", false, "print usage")
	flag.Parse()

	if *help {
		fmt.Println(`glacier archive upload <region> <vault> <file> [<description>]
glacier archive delete <region> <vault> <archive>
glacier job inventory <region> <vault> [<topic> <description>]
glacier job archive <region> <vault> <archive> [<topic> <description>]
glacier job list <region> <vault>
glacier job describe <region> <vault> <job>
glacier job get inventory <region> <vault> <job>
glacier job get archive <region> <vault> <job> <file>
glacier job run <region> <vault> <archive> <size> <file> [<topic> <description>]
glacier job resume <file>
glacier multipart init <region> <vault> <file> <size> [<description>]
glacier multipart run <region> <vault> <file> <size> [<description>]
glacier multipart print <file>
glacier multipart resume <file> [<parts>]
glacier multipart abort <file>
glacier multipart list parts <file>
glacier multipart list uploads <vault>
glacier vault create <region> <vault>
glacier vault delete <region> <vault>
glacier vault describe <region> <vault>
glacier vault list <region>
glacier vault notifications set <region> <vault> <topic>
glacier vault notifications get <region> <vault>
glacier vault notifications delete <region> <vault>`)
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

	args := flag.Args()

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

func prettySize(size uint64) string {
	if size >= 1024*1024*1024 {
		return fmt.Sprintf("%.1f GiB", float32(size)/1024.0/1024.0/1024.0)
	}
	if size >= 1024*1024 {
		return fmt.Sprintf("%.1f MiB", float32(size)/1024.0/1024.0)
	}
	if size >= 1024 {
		return fmt.Sprintf("%.1f KiB", float32(size)/1024.0)
	}
	return fmt.Sprint(size)
}

func getConnection(args []string) []string {
	// TODO other ways to supply them
	secret, access := aws.KeysFromEnviroment()
	if secret == "" || access == "" {
		fmt.Println("could not get keys")
		os.Exit(1)
	}

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

	connection = glacier.NewConnection(secret, access, region)

	return args[1:]
}
