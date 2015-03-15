package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rdwilliamson/aws/glacier"
)

func policy(args []string) {
	if len(args) < 1 {
		fmt.Println("no policy command")
		os.Exit(1)
	}
	command := args[0]
	args = args[1:]

	switch command {
	case "data-retrieval":
		dataRetrieval(args)

	default:
		fmt.Println("unknown policy command:", command)
		os.Exit(1)
	}
}

func dataRetrieval(args []string) {
	if len(args) < 1 {
		fmt.Println("no data-retrieval command")
		os.Exit(1)
	}
	command := args[0]
	args = args[1:]

	switch command {
	case "get":
		args = getConnection(args)
		policy, bytesPerHour, err := connection.GetDataRetrievalPolicy()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Policy:", policy)
		if bytesPerHour != 0 {
			fmt.Printf("Bytes per hour: %s (%d)\n", prettySize(int64(bytesPerHour)), bytesPerHour)
		}

	case "set":
		args = getConnection(args)
		if len(args) < 1 {
			fmt.Println("no data-retrieval policy")
		}
		policy := glacier.ToDataRetrievalPolicy(args[0])
		var bytesPerHour int
		if len(args) >= 2 {
			var err error
			bytesPerHour, err = strconv.Atoi(args[1])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		err := connection.SetRetrievalPolicy(policy, bytesPerHour)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	default:
		fmt.Println("unknown data-retrieval command:", command)
		os.Exit(1)
	}
}
