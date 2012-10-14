package main

import (
	"fmt"
	"io"
	"os"
)

// $ glacier us-east-1 job inventory <vault> <topic> <description>
// $ glacier us-east-1 job archive <vault> <archive> <topic> <description>
// $ glacier us-east-1 job list <vault>
// $ glacier us-east-1 job describe <vault> <job>
// $ glacier us-east-1 job get inventory <vault> <job>
// $ glacier us-east-1 job get archive <vault> <job> <file>

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
			topic = args[0]
		}
		if len(args) > 1 {
			description = args[1]
		}

		jobId, err := connection.InitiateInventoryJob(vault, topic, description)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(jobId)

	case "archive":
		if len(args) < 2 {
			fmt.Println("no vault")
			os.Exit(1)
		}
		vault := args[0]
		archive := args[1]
		args = args[2:]

		var description, topic string
		if len(args) > 0 {
			topic = args[0]
		}
		if len(args) > 1 {
			description = args[1]
		}

		jobId, err := connection.InitiateRetrievalJob(vault, archive, topic, description)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(jobId)

	case "list":
		if len(args) < 1 {
			fmt.Println("no vault")
			os.Exit(1)
		}
		vault := args[0]

		jobs, _, err := connection.ListJobs(vault, "", "", "", 0)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("%+v\n", jobs)

	case "describe":
		if len(args) < 2 {
			fmt.Println("no vault and/or job id")
			os.Exit(1)
		}
		vault := args[0]
		jobId := args[1]

		job, err := connection.DescribeJob(vault, jobId)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("%+v\n", *job)

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

			fmt.Println("Vault ARN:", inventory.VaultARN)
			fmt.Println("Inventory Date:", inventory.InventoryDate)
			for _, v := range inventory.ArchiveList {
				fmt.Println("Archive ID:", v.ArchiveId)
				fmt.Println("Archive Description:", v.ArchiveDescription)
				fmt.Println("Creation Date:", v.CreationDate)
				fmt.Println("Size:", v.Size)
				fmt.Println("SHA256 Tree Hash:", v.SHA256TreeHash)
				fmt.Println()
			}

		case "archive":
			// TODO retrieve parts and handle errors
			if len(args) < 3 {
				fmt.Println("no vault, job id, and/or output file")
				os.Exit(1)
			}
			vault := args[0]
			job := args[1]
			fileName := args[2]

			file, err := os.Create(fileName)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			defer file.Close()

			archive, err := connection.GetRetrievalJob(vault, job, 0, 0)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			defer archive.Close()

			_, err = io.Copy(file, archive)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		default:
			fmt.Println("unknown job sub command:", subCommand)
			os.Exit(1)
		}
	default:
		fmt.Println("unknown job command:", command)
		os.Exit(1)
	}
}
