package main

import (
	"flag"
	"fmt"
	// "github.com/rdwilliamson/aws/glacier"
	"../aws/glacier"
	"os"
)

// $ glacier us-east-1 vault create <name>
// $ glacier us-east-1 vault delete <name>
// $ glacier us-east-1 vault describe <name>
// $ glacier us-east-1 vault list
// $ glacier us-east-1 vault notification set <name> <topic>
// $ glacier us-east-1 vault notification get <name>
// $ glacier us-east-1 vault notification delete <name>

func vault() {
	switch flag.Arg(2) {
	case "create":
		err := connection.CreateVault(flag.Arg(3))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "delete":
		err := connection.DeleteVault(flag.Arg(3))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "describe":
		vault, err := connection.DescribeVault(flag.Arg(3))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("%+v\n", vault)
	case "list":
		_, vaults, err := connection.ListVaults(0, "")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("%+v\n", vaults)
	case "notifications":
		if flag.NArg() < 5 {
			fmt.Println("no notification command")
			os.Exit(1)
		}
		switch flag.Arg(3) {
		case "set":
			notifications := glacier.Notifications{[]string{
				"ArchiveRetrievalCompleted", "InventoryRetrievalCompleted"},
				flag.Arg(5)}
			err := connection.SetVaultNotifications(flag.Arg(4), notifications)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		case "get":
			notifications, err := connection.GetVaultNotifications(flag.Arg(4))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Printf("%+v\n", notifications)
		case "delete":
			err := connection.DeleteVaultNotifications(flag.Arg(4))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		default:
			fmt.Println("unknown notification command:", flag.Arg(3))
			os.Exit(1)
		}
	default:
		fmt.Println("unknown vault command:", flag.Arg(2))
		os.Exit(1)
	}
}
