package main

import (
	"fmt"
	"os"
)

// $ glacier us-east-1 job inventory <vault> <description> <topic>
// $ glacier us-east-1 job list <vault>
// $ glacier us-east-1 job get inventory <vault> <job>

func job(args []string) {
	if len(args) < 1 {
		fmt.Println("no job command")
		os.Exit(1)
	}
	command := args[0]
	args = args[1:]

	switch command {
	case "inventory":
		if len(args) < 1 {
			fmt.Println("no vault")
			os.Exit(1)
		}
		vault := args[0]
		args = args[1:]

		var description, topic string
		if len(args) > 0 {
			description = args[0]
		}
		if len(args) > 1 {
			topic = args[1]
		}

		jobId, err := connection.InitiateInventoryJob(vault, description, topic)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(jobId)
	case "archive":
		fmt.Println("not implemented")
		os.Exit(1)
	case "list":
		if len(args) < 1 {
			fmt.Println("no vault")
			os.Exit(1)
		}
		vault := args[0]

		jobs, _, err := connection.ListJobs(vault, "", "", "", "")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("%+v\n", jobs)
	case "get":
		if len(args) < 1 {
			fmt.Println("no job sub command")
			os.Exit(1)
		}
		subCommand := args[0]
		args = args[1:]

		switch subCommand {
		case "inventory":
			if len(args) < 2 {
				fmt.Println("no vault and/or job id")
				os.Exit(1)
			}
			vault := args[0]
			job := args[1]

			inventory, err := connection.GetInventoryJob(vault, job)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Printf("%+v\n", inventory)
		case "archive":
			fmt.Println("not implemented")
			os.Exit(1)
		default:
			fmt.Println("unknown job sub command:", subCommand)
			os.Exit(1)
		}
	default:
		fmt.Println("unknown job command:", command)
		os.Exit(1)
	}
}
