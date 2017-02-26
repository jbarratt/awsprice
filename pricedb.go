package awsprice

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
)

// PriceAttr captures all the supported variables that impact a price
type PriceAttr struct {
	Region string
}

// PriceData returns information about the product
// and it's pricing
type PriceData struct {
	Price   float64
	Product EC2Attr
}

// PriceList is a slice of prices
type PriceList []PriceData

// String returns a simple string version of the pricing
func (pd PriceData) String() string {
	return fmt.Sprintf("$%0.3f /hr, $%0.2f /mo", pd.Price, pd.Price*730)
}

func (slice PriceList) Len() int {
	return len(slice)
}

func (slice PriceList) Less(i, j int) bool {
	return slice[i].Price < slice[j].Price
}

func (slice PriceList) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// PriceTable returns a stringified table of all the results
func PriceTable(pl PriceList) string {
	var b bytes.Buffer
	table := tablewriter.NewWriter(&b)
	table.SetHeader([]string{"type", "vCPU", "Mem", "$/hr", "$/mo"})

	sort.Sort(pl)

	for _, pd := range pl {
		table.Append([]string{pd.Product.InstanceType, pd.Product.VCPU, pd.Product.Memory, fmt.Sprintf("$%0.3f", pd.Price), fmt.Sprintf("$%0.2f", pd.Price*730)})
	}
	table.Render()
	return b.String()
}

// Pricer is a standard interface for price lookups. Given a name and
// defined attributes, (such as Region), it will return a floating point
// hourly price
type Pricer interface {
	Set(name string, product EC2Attr, attr PriceAttr, price float64) error
	Get(name string, attr PriceAttr) (PriceData, error)
	Search(name string, attr PriceAttr) []PriceData
}

type productNode struct {
	Product EC2Attr
	Prices  map[PriceAttr]float64
}

// SimplePrices is an implementation of the Pricer interface
// It's not very smart, but it works.
type SimplePrices map[string]productNode

const summaryDBFile = "_SummaryDB.gob"

// Set sets a value (with optional attributes) to a given hourly price
func (ep *SimplePrices) Set(name string, product EC2Attr, attr PriceAttr, price float64) error {
	if val, ok := (*ep)[name]; ok {
		val.Product = product
		val.Prices[attr] = price
	} else {
		pn := productNode{Product: product, Prices: make(map[PriceAttr]float64)}
		pn.Prices[attr] = price
		(*ep)[name] = pn
	}
	return nil
}

// Get returns an hourly price (or an error, if such a thing happens)
// when given a name and optional attributes
func (ep *SimplePrices) Get(name string, attr PriceAttr) (PriceData, error) {
	if val, ok := (*ep)[name]; ok {
		if price, ok := val.Prices[attr]; ok {
			return PriceData{Price: price, Product: val.Product}, nil
		}
	}
	return PriceData{}, errors.New("Pricing data not found")
}

// Search returns a slice of all matching PriceData
func (ep *SimplePrices) Search(name string, attr PriceAttr) []PriceData {
	results := make([]PriceData, 0, 6)
	for key, val := range *ep {
		if strings.Contains(key, name) {
			if price, ok := val.Prices[attr]; ok {
				results = append(results, PriceData{Price: price, Product: val.Product})
			}
		}
	}
	return results
}

func (ep SimplePrices) save() error {
	file, err := os.Create(filepath.Join(cacheDir, summaryDBFile))
	defer func() {
		err = file.Close()
	}()
	if err != nil {
		return err
	}

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(ep)
	return err
}

// LoadSimplePrices loads the simple pricing "database" into memory
// The current implementation is a GOB file
func LoadSimplePrices() (*SimplePrices, error) {
	file, err := os.Open(filepath.Join(cacheDir, summaryDBFile))
	defer func() {
		err = file.Close()
	}()
	if err != nil {
		return nil, err
	}
	decoder := gob.NewDecoder(file)
	ep := NewSimplePrices()
	err = decoder.Decode(ep)
	if err != nil {
		return nil, err
	}
	return ep, nil
}

// NewSimplePrices creates a new SimplePrices data structure
func NewSimplePrices() *SimplePrices {
	ep := make(SimplePrices)
	return &ep
}
