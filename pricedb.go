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

// HoursPerMonth is the average hours in a month. (365*24)/12 = 720
const HoursPerMonth = 730

// OfferType provides a set of constants that point to offer types
type OfferType int

// EC2 and other constants are the values of EC2 offers
const (
	EC2 OfferType = iota
	RDS
	S3
	EBS
)

// OfferList is a slice of Offers
type OfferList []Offer

func (slice OfferList) Len() int {
	return len(slice)
}

func (slice OfferList) Less(i, j int) bool {
	return slice[i].HourlyPrice() < slice[j].HourlyPrice()
}

func (slice OfferList) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// PriceTable returns a stringified table of all the results
func PriceTable(el OfferList) string {
	var b bytes.Buffer
	if len(el) == 0 {
		return b.String()
	}
	tables := make(map[OfferType]*tablewriter.Table)

	sort.Sort(el)

	for _, eo := range el {
		if writer, ok := tables[eo.Type()]; ok {
			writer.Append(eo.RowData())
		} else {
			writer = tablewriter.NewWriter(&b)
			writer.SetHeader(eo.Columns())
			tables[eo.Type()] = writer
			writer.Append(eo.RowData())
		}
	}
	for _, writer := range tables {
		writer.Render()
	}
	return b.String()
}

// Offer is the standard interface for offers.
// It is implemented by all objects which end up with prices
type Offer interface {
	HourlyPrice() float64
	Type() OfferType
	Name() string
	String() string
	Columns() []string
	RowData() []string
	// Description() string
}

// Pricer is a standard interface for price lookups. Given a name and
// defined attributes, (such as Region), it will return a floating point
// hourly price
type Pricer interface {
	StoreEC2(name string, attr map[string]string, offer EC2Offer) error
	StoreRDS(name string, attr map[string]string, offer RDSOffer) error
	Get(name string, attr map[string]string) (Offer, error)
	Search(name string, attr map[string]string) []Offer
}

// PriceDB is the high level storage container
// for all the pricing data. It has utility methods for storing,
// loading, and searching the price data.
type PriceDB struct {
	// OfferLookup maps a name (like 'm4.xlarge') to a type (EC2)
	OfferLookup map[string]OfferType
	EC2         map[EC2OfferParam]EC2Offer
	RDS         map[RDSOfferParam]RDSOffer
}

const summaryDBFile = "_SummaryDB_v0.2.gob"

// StoreEC2 sets a value (with optional attributes) to a given hourly price
func (pd *PriceDB) StoreEC2(name string, attr map[string]string, offer EC2Offer) error {

	// register this name as an EC2 type.
	pd.OfferLookup[name] = EC2
	offerParam, err := NewEC2OfferParam(name, attr)
	if err != nil {
		return err
	}
	(*pd).EC2[offerParam] = offer
	return nil
}

// StoreRDS sets a value (with optional attributes) to a given hourly price
func (pd *PriceDB) StoreRDS(name string, attr map[string]string, offer RDSOffer) error {

	// register this name as an EC2 type.
	pd.OfferLookup[name] = RDS
	offerParam, err := NewRDSOfferParam(name, attr)
	if err != nil {
		return err
	}
	(*pd).RDS[offerParam] = offer
	return nil
}

// Get returns an hourly price (or an error, if such a thing happens)
// when given a name and optional attributes
func (pd *PriceDB) Get(name string, attr map[string]string) (Offer, error) {

	offerType, ok := (*pd).OfferLookup[name]
	if !ok {
		return nil, fmt.Errorf("No known resources named %s", name)
	}
	switch offerType {
	case EC2:
		offerParam, err := NewEC2OfferParam(name, attr)
		if err != nil {
			return nil, err
		}
		if ec2Offer, ok := (*pd).EC2[offerParam]; ok {
			return ec2Offer, nil
		}
		return nil, fmt.Errorf("No matching EC2 records found")
	case RDS:
		offerParam, err := NewRDSOfferParam(name, attr)
		if err != nil {
			return nil, err
		}
		if rdsOffer, ok := (*pd).RDS[offerParam]; ok {
			return rdsOffer, nil
		}
		return nil, fmt.Errorf("No matching RDS records found")
	}
	return nil, errors.New("Pricing data not found")
}

// Search returns a slice of all matching Offers
func (pd *PriceDB) Search(name string, attr map[string]string) []Offer {
	results := make([]Offer, 0, 6)
	for key := range (*pd).OfferLookup {
		if strings.Contains(key, name) {
			offer, err := pd.Get(key, attr)
			if err == nil {
				results = append(results, offer)
			}
		}
	}
	return results
}

func (pd PriceDB) save() error {
	file, err := os.Create(filepath.Join(cacheDir, summaryDBFile))
	defer func() {
		err = file.Close()
	}()
	if err != nil {
		return err
	}

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(pd)
	return err
}

// LoadPriceDB loads the simple pricing "database" into memory
// The current implementation is a GOB file
func LoadPriceDB() (*PriceDB, error) {
	file, err := os.Open(filepath.Join(cacheDir, summaryDBFile))
	defer func() {
		err = file.Close()
	}()
	if err != nil {
		return nil, err
	}
	decoder := gob.NewDecoder(file)
	db := NewPriceDB()
	err = decoder.Decode(db)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// NewPriceDB creates a new PriceDB data structure
func NewPriceDB() *PriceDB {
	db := PriceDB{}
	db.OfferLookup = make(map[string]OfferType)
	db.EC2 = make(map[EC2OfferParam]EC2Offer)
	db.RDS = make(map[RDSOfferParam]RDSOffer)
	return &db
}
