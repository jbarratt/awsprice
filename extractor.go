package awsprice

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

/* This package processes AWS pricing JSON files and compiles into a local
 * cache database
 */

func simplePrice(terms map[string]EC2TermItem) (float64, error) {
	for _, term := range terms {
		for _, dimension := range term.PriceDimensions {
			hourly, err := strconv.ParseFloat(dimension.PricePerUnit["USD"], 64)
			if err == nil {
				return hourly, nil
			}
		}
	}
	return 0.0, fmt.Errorf("Error getting pricing from %+v", terms)
}

// ProcessJSON does the top level dispatching of processing all the AWS
// pricing JSON files and distilling them.
func ProcessJSON() {
	ec2path := filepath.Join(cacheDir, "AmazonEC2.json")
	file, err := ioutil.ReadFile(ec2path)
	if err != nil {
		log.Printf("Error loading EC2 JSON: %v\n", err)
		os.Exit(1)
	}
	var offerIndex EC2OfferIndex
	err = json.Unmarshal(file, &offerIndex)
	if err != nil {
		log.Printf("Unable to parse EC2 offer file: %v\n", err)
		os.Exit(1)
	}
	// Right now, locked to Linux/Shared
	instancePrice := NewSimplePrices()
	for _, p := range offerIndex.Products {
		if !(p.Attr.OperatingSystem != "Linux" && p.Attr.Tenancy == "Shared") {
			continue
		}
		region, err := NewRegion(p.Attr.Location)
		if err != nil {
			log.Printf("Region %s is unknown\n", p.Attr.Location)
			continue
		}
		// p.Attr.InstanceType (c4.xlarge)
		terms, ok := offerIndex.Terms.OnDemand[p.SKU]
		if !ok {
			log.Printf("No offers found for %s @ SKU=%s\n", p.Attr.InstanceType, p.SKU)
			continue
		}
		price, err := simplePrice(terms)
		if err != nil {
			log.Printf("Unable to get price for %s: %s\n", p.Attr.InstanceType, err)
			continue
		}
		err = instancePrice.Set(p.Attr.InstanceType, p.Attr, PriceAttr{Region: region}, price)
		if err != nil {
			log.Printf("Unable to store instance price: %v\n", err)
			continue
		}
	}

	err = instancePrice.save()
	if err != nil {
		log.Printf("Unable to save summary DB: %s\n", err)
	}
}
