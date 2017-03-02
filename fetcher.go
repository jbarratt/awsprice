package awsprice

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/cavaliercoder/grab"
)

const uriBase = "https://pricing.us-east-1.amazonaws.com"
const offerPath = "/offers/v1.0/aws/index.json"

var cacheDir string

func init() {
	cacheDir = makeCacheDir()
}

// OfferIndex contains the top level information about the various Offers
// aka Amazon Service families
type OfferIndex struct {
	FormatVersion   string               `json:"formatVersion"`
	Disclaimer      string               `json:"disclaimer"`
	PublicationDate string               `json:"publicationDate"`
	Offers          map[string]JSONOffer `json:"offers"`
}

// JSONOffer identifies a singular Offer File for a given service
type JSONOffer struct {
	OfferCode         string `json:"offerCode"`
	VersionIndexURL   string `json:"versionIndexUrl"`
	CurrentVersionURL string `json:"currentVersionUrl"`
}

func makeCacheDir() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	cachedir := filepath.Join(usr.HomeDir, ".awsprice_cache")
	err = os.MkdirAll(cachedir, os.ModePerm)
	if err != nil {
		log.Fatalf("Unable to create cache directory")
	}
	return cachedir
}

func updateOfferJSON() {
	err := fetchOfferFile(offerPath, "offer.json")
	if err != nil {
		panic(err)
	}
}

func fetchOfferFile(relativeURL string, outputName string) error {
	client := grab.NewClient()
	client.UserAgent = "AWS Price Grammar Bot"
	url := uriBase + relativeURL
	fmt.Printf("Downloading %s...\n", url)
	req, err := grab.NewRequest(url)
	if err != nil {
		return err
	}
	req.SkipExisting = true
	req.Filename = filepath.Join(cacheDir, outputName)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Issue downloading %s: %v\n", url, err)
		// if file exists anyway, move on
		if _, err := os.Stat(resp.Filename); err != nil {
			return err
		}
	}
	fmt.Printf("Downloaded to %s\n", resp.Filename)
	return nil
}

func processOfferJSON() {

	file, err := ioutil.ReadFile(filepath.Join(cacheDir, "offer.json"))
	if err != nil {
		log.Printf("Error loading offer: %v\n", err)
		os.Exit(1)
	}
	var offerIndex OfferIndex
	err = json.Unmarshal(file, &offerIndex)
	if err != nil {
		log.Printf("Unable to parse offer file: %v\n", err)
		os.Exit(1)
	}

	// TODO make this somewhere DRY
	err = fetchOfferFile(offerIndex.Offers["AmazonEC2"].CurrentVersionURL, "AmazonEC2.json")
	if err != nil {
		log.Printf("Failed to download EC2 offer: %v\n", err)
		os.Exit(1)
	}
}

// FetchJSON downloads all the AWS Pricing JSON files that
// the tool is aware of how to utilize.
func FetchJSON() {
	updateOfferJSON()
	processOfferJSON()
}
