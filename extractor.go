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
	// for now, pick simple region/attributes
	instanceTypes := make(map[string]EC2Product)
	for _, p := range offerIndex.Products {
		if p.Attr.Location == "US West (Oregon)" && p.Attr.OperatingSystem == "Linux" && p.Attr.Tenancy == "Shared" {
			if val, ok := instanceTypes[p.Attr.InstanceType]; ok {
				log.Printf("Duplicate instance type: %+v\n %+v\n", val, p)
			} else {
				instanceTypes[p.Attr.InstanceType] = p
			}
		}
	}
	// for each instance, find a simple hourly price
	instancePrice := NewSimplePrices()
	for instance, p := range instanceTypes {
		terms, ok := offerIndex.Terms.OnDemand[p.SKU]
		if !ok {
			log.Printf("No offers found for %s @ SKU=%s\n", instance, p.SKU)
			continue
		}
		price, err := simplePrice(terms)
		if err != nil {
			log.Printf("Unable to get price for %s: %s\n", instance, err)
			continue
		}
		err = instancePrice.Set(instance, p.Attr, PriceAttr{}, price)
		if err != nil {
			log.Printf("Unable to store instance price: %v\n", err)
			continue
		}
	}
	// fmt.Printf("%s: $%.3f/hr, $%.2f/mo\n", instance, instancePrice[instance], instancePrice[instance]*730)
	err = instancePrice.save()
	if err != nil {
		log.Printf("Unable to save summary DB: %s\n", err)
	}
}
