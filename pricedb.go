package awsprice

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
)

// PriceAttr captures all the supported variables that impact a price
type PriceAttr struct {
	Region string
}

// Pricer is a standard interface for price lookups. Given a name and
// defined attributes, (such as Region), it will return a floating point
// hourly price
type Pricer interface {
	Set(name string, attr PriceAttr, price float64) error
	Get(name string, attr PriceAttr) (float64, error)
}

// SimplePrices is an implementation of the Pricer interface
// It's not very smart, and doesn't even use attributes (yet), but it works.
type SimplePrices map[string]float64

const summaryDBFile = "_SummaryDB.json"

// Set sets a value (with optional attributes) to a given hourly price
func (ep *SimplePrices) Set(name string, attr PriceAttr, price float64) error {
	// simple pricer ignores attributes. Lazy.
	(*ep)[name] = price
	return nil
}

// Get returns an hourly price (or an error, if such a thing happens)
// when given a name and optional attributes
func (ep *SimplePrices) Get(name string, attr PriceAttr) (float64, error) {
	if val, ok := (*ep)[name]; ok {
		return val, nil
	}
	return 0.0, errors.New("Pricing data not found")
}

func (ep SimplePrices) save() error {
	b, err := json.Marshal(ep)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(cacheDir, summaryDBFile), b, 0644)
	return err
}

// LoadSimplePrices loads the simple pricing "database" into memory
// The current implementation is a JSON file
func LoadSimplePrices() (*SimplePrices, error) {
	file, err := ioutil.ReadFile(filepath.Join(cacheDir, summaryDBFile))
	if err != nil {
		return nil, err
	}
	ep := NewSimplePrices()
	err = json.Unmarshal(file, ep)
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
