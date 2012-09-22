package main

import (
	"../aws/glacier"
	"fmt"
	"io"
	"os"
)

// $ glacier us-east-1 archive multipart upload <vault> <file> <description>

func multipart(args []string) {
	if len(args) < 1 {
		fmt.Println("no multipart command")
		os.Exit(1)
	}
	command := args[0]
	args = args[1:]

	switch command {
	case "create", "read":
		fmt.Println("not implemented")
		os.Exit(1)
	case "upload":
		if len(args) < 2 {
			fmt.Println("no vault and/or file")
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
		defer file.Close()

		size := uint(1024 * 1024)
		uploadId, err := connection.InitiateMultipart(vault, size, description)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		buffer := make([]byte, size)
		at := uint(0)
		var n int
		for err == nil {
			n, err = file.Read(buffer)
			fmt.Println("uploaded", at, "sending", n, "read error", err)
			if n == 0 {
				break
			}
			err = connection.UploadMultipart(vault, uploadId, at, buffer[:n])
			fmt.Println("upload error", err)
			if err != nil {
				// TODO cancel multipart upload
				fmt.Println(err)
				os.Exit(1)
			}
			at += uint(n)
			fmt.Println("uploaded", at/1024/1024, "MiB so far...")
		}
		if err != io.EOF {
			if err == nil {
				fmt.Println("expecting EOF")
			} else {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		_, err = file.Seek(0, 0)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		treeHash, _, err := glacier.GetTreeHash(file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("completing...")
		location, err := connection.CompleteMultipart(vault, uploadId,
			treeHash, at)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(location)
	default:
		fmt.Println("unknown multipart command:", command)
		os.Exit(1)
	}
}
