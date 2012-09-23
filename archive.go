package main

import (
	"fmt"
	"os"
)

// $ glacier us-east-1 archive upload <vault> <file> <description>
// $ glacier us-east-1 archive delete <vault> <archive>

func archive(args []string) {
	if len(args) < 1 {
		fmt.Println("no archive command")
		os.Exit(1)
	}
	command := args[0]
	args = args[1:]

	switch command {
	case "upload":
		if len(args) < 2 {
			fmt.Println("no vault and file")
			os.Exit(1)
		}
		vault := args[0]
		filename := args[1]
		var description string
		if len(args) > 2 {
			description = args[2]
		} else {
			description = filename
		}

		file, err := os.Open(filename)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		location, err := connection.UploadArchive(vault, file, description)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(location)
	case "delete":
		if len(args) < 2 {
			fmt.Println("no vault and/or archive")
			os.Exit(1)
		}
		vault := args[0]
		archive := args[1]

		err := connection.DeleteArchive(vault, archive)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		fmt.Println("unknown archive command:", command)
		os.Exit(1)
	}
}
