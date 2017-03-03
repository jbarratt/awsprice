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

func simpleEC2Price(terms map[string]EC2TermItem) (float64, error) {
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

func simpleRDSPrice(terms map[string]RDSTermItem) (float64, error) {
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

func extractEC2(priceDB *PriceDB) {
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
	for _, p := range offerIndex.Products {
		if !(p.Attr.OperatingSystem != "Linux" && p.Attr.Tenancy == "Shared") {
			continue
		}
		if p.Attr.Location == "AWS GovCloud (US)" {
			continue
		}
		terms, ok := offerIndex.Terms.OnDemand[p.SKU]
		if !ok {
			log.Printf("No offers found for %s @ SKU=%s\n", p.Attr.InstanceType, p.SKU)
			continue
		}
		price, err := simpleEC2Price(terms)
		if err != nil {
			log.Printf("Unable to get price for %s: %s\n", p.Attr.InstanceType, err)
			continue
		}

		offer := EC2Offer{Price: price, Product: p.Attr}
		err = priceDB.StoreEC2(p.Attr.InstanceType, map[string]string{"region": p.Attr.Location}, offer)
		if err != nil {
			log.Printf("Unable to store instance price: %v\n", err)
			continue
		}
	}
}

func extractRDS(priceDB *PriceDB) {
	rdsPath := filepath.Join(cacheDir, "AmazonRDS.json")
	file, err := ioutil.ReadFile(rdsPath)
	if err != nil {
		log.Printf("Error loading RDS JSON: %v\n", err)
		os.Exit(1)
	}
	var offerIndex RDSOfferIndex
	err = json.Unmarshal(file, &offerIndex)
	if err != nil {
		log.Printf("Unable to parse RDS offer file: %v\n", err)
		os.Exit(1)
	}
	for _, p := range offerIndex.Products {
		if p.Attr.Location == "AWS GovCloud (US)" {
			continue
		}
		if p.Attr.ServiceCode == "AWSDataTransfer" {
			continue
		}
		terms, ok := offerIndex.Terms.OnDemand[p.SKU]
		if !ok {
			log.Printf("No offers found for %s @ SKU=%s\n", p.Attr.InstanceType, p.SKU)
			continue
		}
		price, err := simpleRDSPrice(terms)
		if err != nil {
			log.Printf("Unable to get price for %s: %s\n", p.Attr.InstanceType, err)
			continue
		}

		offer := RDSOffer{Price: price, Product: p.Attr}
		err = priceDB.StoreRDS(p.Attr.InstanceType, map[string]string{"region": p.Attr.Location,
			"engine": p.Attr.DatabaseEngine, "deployment": p.Attr.DeploymentOption}, offer)
		if err != nil {
			log.Printf("Unable to store RDS instance price: %v\n", err)
			log.Printf("%+v\n", p.Attr)
			continue
		}
	}
}

// ProcessJSON does the top level dispatching of processing all the AWS
// pricing JSON files and distilling them.
func ProcessJSON() {
	priceDB := NewPriceDB()
	extractEC2(priceDB)
	extractRDS(priceDB)
	err := priceDB.save()
	if err != nil {
		log.Printf("Unable to save summary DB: %s\n", err)
	}
}
