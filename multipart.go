package main

import (
	"encoding/gob"
	"fmt"
	"github.com/rdwilliamson/aws"
	"github.com/rdwilliamson/aws/glacier"
	"io"
	"net"
	"os"
	"reflect"
	"strconv"
	"syscall"
)

// $ glacier us-east-1 multipart init <vault> <file> <size> <description>
// $ glacier us-east-1 multipart print <file>
// $ glacier us-east-1 multipart run <file> <parts>
// $ glacier us-east-1 multipart abort <file>
// $ glacier us-east-1 multipart list parts <file>

type multipartData struct {
	Vault       string
	Description string
	PartSize    uint
	FileName    string
	UploadId    string
	Parts       []multipartPart
	TreeHash    string
	Size        uint64
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

		partHasher := glacier.NewTreeHash()
		wholeHasher := glacier.NewTreeHash()
		hasher := io.MultiWriter(partHasher, wholeHasher)
		for i := range data.Parts {
			n, err := io.CopyN(hasher, f, int64(data.PartSize))
			if err != nil && err != io.EOF {
				fmt.Println(err)
				os.Exit(1)
			}
			data.Size += uint64(n)
			partHasher.Close()
			data.Parts[i].Hash = partHasher.Hash()
			data.Parts[i].TreeHash = partHasher.TreeHash()
			partHasher.Reset()
		}
		wholeHasher.Close()
		data.TreeHash = wholeHasher.TreeHash()

		data.UploadId, err = connection.InitiateMultipart(data.Vault, data.PartSize, data.Description)
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
		fmt.Printf("Part Size: %dMiB\n", data.PartSize/1024/1024)
		fmt.Println("Upload ID:", data.UploadId)
		uploaded := 0
		for i := range data.Parts {
			if data.Parts[i].Uploaded {
				uploaded++
			}
		}
		fmt.Println("Parts Uploaded", uploaded, "/", len(data.Parts))
		fmt.Println("Tree Hash:", data.TreeHash)
		fmt.Println("Size:", data.Size, prettySize(data.Size))

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

		start := uint64(0)
		index := 0
		for _, v := range data.Parts {
			if v.Uploaded {
				start += uint64(data.PartSize)
				index++
			} else {
				break
			}
		}

		if len(data.Parts) < parts {
			parts = len(data.Parts)
		}
		if parts == 0 {
			parts = len(data.Parts)
		}

		try := 0
		for i := 0; i < parts; i++ {
			if index >= len(data.Parts) {
				break
			}

			_, err = file.Seek(int64(start), 0)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			body := &limitedReadSeeker{file, int64(data.PartSize), int64(data.PartSize)}

			err = connection.UploadMultipart(data.Vault, data.UploadId, start, body)

			if err != nil {
				switch err.(type) {
				case *net.OpError:
					opError := err.(*net.OpError)
					fmt.Println("caught net.OpError")
					fmt.Println("Error:", opError.Error())
					fmt.Println("Temporary:", opError.Temporary())
					fmt.Println("Timeout:", opError.Timeout())
					fmt.Println("Err", reflect.TypeOf(opError.Err))
					switch opError.Err.(type) {
					case syscall.Errno:
						fmt.Println("caught syscall.Errno")
						errno := opError.Err.(syscall.Errno)
						fmt.Println("Error:", errno.Error())
						fmt.Println("Temporary:", errno.Temporary())
						fmt.Println("Timeout:", errno.Timeout())
						fmt.Println("Errno:", int(errno))
					}
					if try++; try > retrys {
						fmt.Println("too many retrys")
						os.Exit(1)
					}
					continue
				case *aws.Error:
					awsError := err.(*aws.Error)
					fmt.Println("caught aws.Error")
					if awsError.Message == "Request timed out." {
						if try++; try > retrys {
							fmt.Println("too many retrys")
							os.Exit(1)
						}
						continue
					}
					if awsError.Message == "An error has occurred and the request cannot be processed." {
						if try++; try > retrys {
							fmt.Println("too many retrys")
							os.Exit(1)
						}
						continue
					}
					fmt.Println(err)
					os.Exit(1)
				default:
					fmt.Println(reflect.TypeOf(err))
					fmt.Println(err)
					os.Exit(1)
				}
			}

			try = 0

			data.Parts[index].Uploaded = true
			gobFile, err = os.Create(fileName + ".gob.new")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			enc := gob.NewEncoder(gobFile)
			err = enc.Encode(data)
			gobFile.Close()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			err = os.Remove(fileName + ".gob")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			err = os.Rename(fileName+".gob.new", fileName+".gob")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			start += uint64(data.PartSize)
			index++
		}

		done := true
		for _, v := range data.Parts {
			if !v.Uploaded {
				done = false
				break
			}
		}

		if done {
			archiveId, err := connection.CompleteMultipart(data.Vault, data.UploadId, data.TreeHash, data.Size)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(archiveId)

			err = os.Remove(fileName + ".gob")
			if err != nil {
				fmt.Println(err)
			}
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

			parts, err := connection.ListMultipartParts(data.Vault, data.UploadId, "", 0)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Printf("%+v\n", *parts)

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

	default:
		fmt.Println("unknown multipart command:", command)
		os.Exit(1)
	}
}
