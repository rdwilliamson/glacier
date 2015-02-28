package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/rdwilliamson/aws"
	"github.com/rdwilliamson/aws/glacier"
)

type multipartData struct {
	Region      string
	Vault       string
	Description string
	PartSize    int64
	FileName    string
	UploadId    string
	Parts       []multipartPart
	TreeHash    string
	Size        int64
}

type multipartPart struct {
	Hash     string
	TreeHash string
	Uploaded bool
}

type limitedReadSeeker struct {
	R         io.ReadSeeker
	N         int64
	OriginalN int64
}

func (l *limitedReadSeeker) Read(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.R.Read(p)
	l.N -= int64(n)
	return
}

var (
	uploadData multipartData
)

func (l *limitedReadSeeker) Seek(offset int64, whence int) (ret int64, err error) {
	switch whence {
	case 0:
		_, err := l.R.Seek(l.N-l.OriginalN+offset, 1)
		l.N = l.OriginalN - offset
		return offset, err
	case 1:
		// untested
		_, err := l.R.Seek(offset, 1)
		l.N -= offset
		return l.OriginalN - l.N, err
	case 2:
		// untested
		_, err := l.R.Seek(l.N-offset, 1)
		l.N = offset
		return l.OriginalN - l.N, err
	}
	panic("invalid whence")
}

func parseRegion(region string) {
	for _, v := range aws.Regions {
		if region == v.Region {
			getConnection([]string{v.Name})
			break
		}
	}
	if connection == nil {
		fmt.Println("error getting region")
		os.Exit(1)
	}
}

func multipart(args []string) {
	if len(args) < 1 {
		fmt.Println("no multipart command")
		os.Exit(1)
	}
	command := args[0]
	args = args[1:]

	switch command {
	case "init", "run":
		args = getConnection(args)
		uploadData.Region = connection.Signature.Region.Region

		if len(args) < 3 {
			fmt.Println("no vault, file name and/or part size")
			os.Exit(1)
		}
		uploadData.Vault = args[0]
		uploadData.FileName = args[1]
		partSize, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		uploadData.PartSize = partSize * 1024 * 1024
		args = args[3:]

		if len(args) > 0 {
			uploadData.Description = args[0]
		}

		f, err := os.Open(uploadData.FileName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer f.Close()

		s, _ := f.Stat()
		parts := s.Size() / uploadData.PartSize
		if s.Size()%uploadData.PartSize > 0 {
			parts++
		}
		uploadData.Parts = make([]multipartPart, parts)

		partHasher := glacier.NewTreeHash()
		wholeHasher := glacier.NewTreeHash()
		hasher := io.MultiWriter(partHasher, wholeHasher)
		for i := range uploadData.Parts {
			n, err := io.CopyN(hasher, f, uploadData.PartSize)
			if err != nil && err != io.EOF {
				fmt.Println(err)
				os.Exit(1)
			}
			uploadData.Size += n
			partHasher.Close()
			uploadData.Parts[i].Hash = string(toHex(partHasher.Hash()))
			uploadData.Parts[i].TreeHash = string(toHex(partHasher.TreeHash()))
			partHasher.Reset()
		}
		wholeHasher.Close()
		uploadData.TreeHash = string(toHex(wholeHasher.TreeHash()))

		uploadData.UploadId, err = connection.InitiateMultipart(uploadData.Vault, uploadData.PartSize,
			uploadData.Description)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		out, err := os.Create(uploadData.FileName + ".gob")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		enc := gob.NewEncoder(out)
		err = enc.Encode(uploadData)
		out.Close()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if command == "init" {
			return
		}

		fallthrough

	case "resume":
		var parts int
		if command == "resume" {
			if len(args) < 1 {
				fmt.Println("no file")
				os.Exit(1)
			}
			fileName := args[0]
			args = args[1:]

			if len(args) > 0 {
				parts64, err := strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				parts = int(parts64)
			}

			gobFile, err := os.Open(fileName + ".gob")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			dec := gob.NewDecoder(gobFile)
			err = dec.Decode(&uploadData)
			gobFile.Close()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			parseRegion(uploadData.Region)
		}

		file, err := os.Open(uploadData.FileName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer file.Close()

		var start int64
		index := 0
		for _, v := range uploadData.Parts {
			if v.Uploaded {
				start += uploadData.PartSize
				index++
			} else {
				break
			}
		}

		if len(uploadData.Parts) < parts {
			parts = len(uploadData.Parts)
		}
		if parts == 0 {
			parts = len(uploadData.Parts)
		}

		i, try := 0, 0
		for i < parts {
			if index >= len(uploadData.Parts) {
				break
			}

			_, err := file.Seek(start, 0)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			body := &limitedReadSeeker{file, uploadData.PartSize, uploadData.PartSize}

			err = connection.UploadMultipart(uploadData.Vault, uploadData.UploadId, start, body)

			if err != nil {
				fmt.Println(err)
				if try++; try > retries {
					fmt.Println("too many retrys")
					os.Exit(1)
				}
				continue
			}

			i++
			try = 0

			uploadData.Parts[index].Uploaded = true
			gobFile, err := os.Create(uploadData.FileName + ".gob.new")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			enc := gob.NewEncoder(gobFile)
			err = enc.Encode(uploadData)
			gobFile.Close()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			err = os.Remove(uploadData.FileName + ".gob")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			err = os.Rename(uploadData.FileName+".gob.new", uploadData.FileName+".gob")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			start += uploadData.PartSize
			index++
		}

		done := true
		for _, v := range uploadData.Parts {
			if !v.Uploaded {
				done = false
				break
			}
		}

		if done {
			archiveId, err := connection.CompleteMultipart(uploadData.Vault, uploadData.UploadId, uploadData.TreeHash,
				uploadData.Size)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(archiveId)

			err = os.Remove(uploadData.FileName + ".gob")
			if err != nil {
				fmt.Println(err)
			}
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

		fmt.Println("Region:", data.Region)
		fmt.Println("Vault:", data.Vault)
		fmt.Println("Description:", data.Description)
		fmt.Println("Part Size:", prettySize(data.PartSize))
		fmt.Println("Upload ID:", data.UploadId)
		uploaded := 0
		for i := range data.Parts {
			if data.Parts[i].Uploaded {
				uploaded++
			}
		}
		fmt.Println("Parts Uploaded:", uploaded, "/", len(data.Parts))
		fmt.Println("Tree Hash:", data.TreeHash)
		fmt.Println("Size:", data.Size, prettySize(data.Size))

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

		parseRegion(data.Region)

		err = connection.AbortMultipart(data.Vault, data.UploadId)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = os.Remove(fileName + ".gob")
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
			if len(args) < 1 {
				fmt.Println("no file")
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

			parseRegion(data.Region)

			parts, err := connection.ListMultipartParts(data.Vault, data.UploadId, "", 0)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Printf("%+v\n", *parts)

		case "uploads":
			args = getConnection(args)

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

			for _, v := range parts {
				fmt.Println("Archive Description:", v.ArchiveDescription)
				fmt.Println("Creation Data:", v.CreationDate)
				fmt.Println("Multipart Upload ID:", v.MultipartUploadId)
				fmt.Println("Part Size:", prettySize(v.PartSizeInBytes))
				fmt.Println("Vault ARN:", v.VaultARN)
				fmt.Println()
			}

		default:
			fmt.Println("unknown multipart sub command:", subCommand)
			os.Exit(1)
		}

	default:
		fmt.Println("unknown multipart command:", command)
		os.Exit(1)
	}
}
