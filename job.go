package main

import (
	"fmt"
	"os"
)

// $ glacier us-east-1 job inventory <vault> <description> <topic>

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

		err := connection.InitiateInventoryJob(vault, description, topic)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "archive":
		fmt.Println("not implemented")
		os.Exit(1)
	default:
		fmt.Println("unknown job command:", command)
		os.Exit(1)
	}
}
