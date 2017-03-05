package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jbarratt/awsprice"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Call with fetch, process, or with a pricing string")
		os.Exit(1)
	} else if os.Args[1] == "fetch" {
		awsprice.FetchJSON()
	} else if os.Args[1] == "process" {
		awsprice.ProcessJSON()
	} else if os.Args[1] == "help" {
		fmt.Printf("fetch: fetch new pricing data\nprocess: rebuild local pricing db\nhelp: you're looking at it\nAnything else: a pricing string to interpret\n")
	} else {
		pricer, err := awsprice.LoadPriceDB()
		if err != nil {
			// just in case, try to fetch & process
			if strings.Contains(fmt.Sprintf("%s", err), "no such file") {
				awsprice.FetchJSON()
				awsprice.ProcessJSON()
			}
			pricer, err = awsprice.LoadPriceDB()
			if err != nil {
				panic(err)
			}
		}
		value, err := awsprice.ParseInput(pricer, os.Args[1])
		if err != nil {
			fmt.Printf("Unable to find a price for '%s'\n", os.Args[1])
			os.Exit(1)
		}
		fmt.Println(value)
	}
}
