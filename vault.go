package main

import (
	"fmt"
	"github.com/rdwilliamson/aws/glacier"
	"os"
)

func vault(args []string) {
	if len(args) < 1 {
		fmt.Println("no vault command")
		os.Exit(1)
	}
	command := args[0]
	args = args[1:]
	switch command {
	case "create", "delete", "describe":
		args = getConnection(args)

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
		args = getConnection(args)

		vaults, _, err := connection.ListVaults("", 0)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, v := range vaults {
			fmt.Println("Vault Name:", v.VaultName)
			fmt.Println("Vault ARN:", v.VaultARN)
			fmt.Println("Creation Date:", v.CreationDate)
			fmt.Println("Archives;", v.NumberOfArchives)
			fmt.Println("Size:", v.SizeInBytes, prettySize(v.SizeInBytes))
			fmt.Println("Last Inventory Date:", v.LastInventoryDate)
			fmt.Println()
		}

	case "notifications":
		if len(args) < 1 {
			fmt.Println("no notification command")
			os.Exit(1)
		}
		subCommand := args[0]
		args = args[1:]

		args = getConnection(args)

		if len(args) < 1 {
			fmt.Println("no vault name")
			os.Exit(1)
		}
		name := args[0]
		args = args[1:]

		switch subCommand {
		case "set":
			if len(args) < 1 {
				fmt.Println("no notification topic")
				os.Exit(1)
			}
			topic := args[0]
			notifications := &glacier.Notifications{[]string{
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

			fmt.Printf("Events: ")
			for _, v := range notifications.Events {
				fmt.Printf("%s ", v)
			}
			fmt.Println()
			fmt.Println("SNSTopic:", notifications.SNSTopic)

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
