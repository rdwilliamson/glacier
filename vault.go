package main

import (
	"fmt"
	"github.com/rdwilliamson/aws/glacier"
	"os"
)

// $ glacier us-east-1 vault create <name>
// $ glacier us-east-1 vault delete <name>
// $ glacier us-east-1 vault describe <name>
// $ glacier us-east-1 vault list
// $ glacier us-east-1 vault notifications set <name> <topic>
// $ glacier us-east-1 vault notifications get <name>
// $ glacier us-east-1 vault notifications delete <name>

func vault(args []string) {
	if len(args) < 1 {
		fmt.Println("no vault command")
		os.Exit(1)
	}
	command := args[0]
	args = args[1:]
	switch command {
	case "create", "delete", "describe":
		if len(args) < 1 {
			fmt.Println("no vault name")
			os.Exit(1)
		}
		name := args[0]

		switch command {
		case "create":
			err := connection.CreateVault(name)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

		case "delete":
			err := connection.DeleteVault(name)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

		case "describe":
			vault, err := connection.DescribeVault(name)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Printf("%+v\n", vault)
		}

	case "list":
		_, vaults, err := connection.ListVaults(0, "")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("%+v\n", vaults)

	case "notifications":
		args = args[1:]
		if len(args) < 2 {
			fmt.Println("no notification command or no vault name")
			os.Exit(1)
		}
		subCommand := args[0]
		name := args[1]
		args = args[2:]

		switch subCommand {
		case "set":
			if len(args) < 1 {
				fmt.Println("no notification topic")
				os.Exit(1)
			}
			topic := args[0]
			notifications := glacier.Notifications{[]string{
				"ArchiveRetrievalCompleted", "InventoryRetrievalCompleted"},
				topic}
			err := connection.SetVaultNotifications(name, notifications)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

		case "get":
			notifications, err := connection.GetVaultNotifications(name)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Printf("%+v\n", notifications)

		case "delete":
			err := connection.DeleteVaultNotifications(name)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		default:
			fmt.Println("unknown notification command:", subCommand)
			os.Exit(1)
		}
	default:
		fmt.Println("unknown vault command:", command)
		os.Exit(1)
	}
}
