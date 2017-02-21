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
	} else {
		pricer, err := awsprice.LoadSimplePrices()
		if err != nil {
			// just in case, try to fetch & process
			if strings.Contains(fmt.Sprintf("%s", err), "no such file") {
				awsprice.FetchJSON()
				awsprice.ProcessJSON()
			}
			pricer, err = awsprice.LoadSimplePrices()
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
