package main

import (
	"../aws/glacier"
	"flag"
	"fmt"
	"io"
	"os"
)

// $ glacier us-east-1 archive multipart upload <description> <file> <vault>

func multipart() {
	switch flag.Arg(2) {
	case "create":
		in, err := os.Open(flag.Arg(3))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer in.Close()

		out, err := os.Create(flag.Arg(4))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer out.Close()

		err = glacier.CreateTreeHash(in, out)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "read":
		in, err := os.Open(flag.Arg(3))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer in.Close()

		err = glacier.ReadTreeHash(in)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "upload":
		fileName := flag.Arg(4)
		file, err := os.Open(fileName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer file.Close()

		size := uint(1024 * 1024)
		vault := flag.Arg(3)
		uploadId, err := connection.InitiateMultipart(vault, size, fileName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// hasher := sha256.New()
		buffer := make([]byte, size)
		at := uint(0)
		var n int
		for err == nil {
			n, err = file.Read(buffer)
			fmt.Println("uploaded", at, "sending", n, "read error", err)
			if n == 0 {
				break
			}
			// hasher.Write(buffer)
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
		fmt.Println("unknown multipart command:", flag.Arg(2))
		os.Exit(1)
	}
}
