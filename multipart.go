package main

import (
	"encoding/gob"
	"fmt"
	// "github.com/rdwilliamson/aws/glacier"
	"../aws/glacier"
	"io"
	"os"
	"strconv"
)

// $ glacier us-east-1 archive multipart upload <vault> <file> <description>

// $ glacier us-east-1 archive multipart init <vault> <file> <size> <description>
// $ glacier us-east-1 archive multipart print <file>
// $ glacier us-east-1 archive multipart run <file>
// $ glacier us-east-1 archive multipart abort <file>
// $ glacier us-east-1 archive multipart list parts <file>

// $ glacier us-east-1 archive multipart list uploads <vault>

type multipartData struct {
	Vault       string
	Description string
	PartSize    uint
	FileName    string
	UploadId    string
	Parts       []multipartPart
}

type multipartPart struct {
	Hash     string
	TreeHash string
	Uploaded bool
}

func multipart(args []string) {
	if len(args) < 1 {
		fmt.Println("no multipart command")
		os.Exit(1)
	}
	command := args[0]
	args = args[1:]

	switch command {
	case "init":
		var data multipartData

		if len(args) < 3 {
			fmt.Println("no vault, file name and/or part size")
			os.Exit(1)
		}
		data.Vault = args[0]
		data.FileName = args[1]
		partSize, err := strconv.ParseUint(args[2], 10, 32)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		data.PartSize = uint(partSize) * 1024 * 1024
		args = args[3:]

		if len(args) > 0 {
			data.Description = args[0]
		}

		f, err := os.Open(data.FileName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer f.Close()

		s, _ := f.Stat()
		parts := s.Size() / int64(data.PartSize)
		if s.Size()%int64(data.PartSize) > 0 {
			parts++
		}
		data.Parts = make([]multipartPart, parts)

		th := glacier.NewTreeHash()
		for i := range data.Parts {
			_, err := io.CopyN(th, f, int64(data.PartSize))
			if err != nil && err != io.EOF {
				fmt.Println(err)
				os.Exit(1)
			}
			th.Close()
			data.Parts[i].Hash = th.Hash()
			data.Parts[i].TreeHash = th.TreeHash()
			th.Reset()
		}

		data.UploadId, err = connection.InitiateMultipart(data.Vault,
			data.PartSize, data.Description)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		out, err := os.Create(data.FileName + ".gob")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer out.Close()

		enc := gob.NewEncoder(out)
		err = enc.Encode(data)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	case "print":
		if len(args) < 1 {
			fmt.Println("no file name")
			os.Exit(1)
		}
		fileName := args[0]

		f, err := os.Open(fileName + ".gob")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer f.Close()

		dec := gob.NewDecoder(f)
		var data multipartData
		err = dec.Decode(&data)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("Vault:", data.Vault)
		fmt.Println("Description:", data.Description)
		fmt.Println("Part Size:", data.PartSize/1024/1024, "MiB")
		fmt.Println("Upload ID:", data.UploadId)
		uploaded := 0
		for i := range data.Parts {
			if data.Parts[i].Uploaded {
				uploaded++
			}
		}
		fmt.Println("Parts Uploaded", uploaded, "/", len(data.Parts))

	case "run":
		if len(args) < 2 {
			fmt.Println("no file and/or parts")
			os.Exit(1)
		}
		fileName := args[0]
		parts64, err := strconv.ParseInt(args[1], 10, 64)
		parts := int(parts64)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		gobFile, err := os.Open(fileName + ".gob")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		dec := gob.NewDecoder(gobFile)
		var data multipartData
		err = dec.Decode(&data)
		gobFile.Close()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		file, err := os.Open(fileName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer file.Close()

		start := int64(0)
		for _, v := range data.Parts {
			if v.Uploaded {
				start += int64(data.PartSize)
			} else {
				break
			}
		}

		if len(data.Parts) < parts {
			parts = len(data.Parts)
		}
		for i := 0; i < parts; i++ {
			fmt.Println("uploading from", start, "to", uint(start)+data.PartSize)
			_, err = file.Seek(start, 0)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			start += int64(data.PartSize)
		}

	case "abort":
		if len(args) < 1 {
			fmt.Println("no file name")
			os.Exit(1)
		}
		fileName := args[0]

		f, err := os.Open(fileName + ".gob")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer f.Close()

		dec := gob.NewDecoder(f)
		var data multipartData
		err = dec.Decode(&data)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = connection.AbortMultipart(data.Vault, data.UploadId)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	case "list":
		if len(args) < 1 {
			fmt.Println("no multipart sub command")
		}
		subCommand := args[0]
		args = args[1:]

		switch subCommand {
		case "parts":

		case "uploads":
			if len(args) < 1 {
				fmt.Println("no vault")
				os.Exit(1)
			}
			vault := args[0]

			parts, _, err := connection.ListMultipartUploads(vault, "", 0)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Printf("%+v\n", parts)

		default:
			fmt.Println("unknown multipart sub command:", subCommand)
			os.Exit(1)
		}

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
