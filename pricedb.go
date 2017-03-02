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

// EC2Offer The product/price details for a given EC2 Offering
type EC2Offer struct {
	Product EC2Attr
	Price   float64
}

// Name returns the EC2 instance type
func (eo EC2Offer) Name() string {
	return eo.Product.InstanceType
}

// HourlyPrice returns the fractional dollars per hour
func (eo EC2Offer) HourlyPrice() float64 {
	return eo.Price
}

// Type always returns EC2
func (eo EC2Offer) Type() OfferType {
	return EC2
}

// EC2OfferParam stores the unique factors that determine an EC2 Offer
type EC2OfferParam struct {
	Region Region
	Name   string
}

// NewEC2OfferParam constructs an EC2 offer from a name & attributes
func NewEC2OfferParam(name string, attr map[string]string) (EC2OfferParam, error) {
	offerParams := &EC2OfferParam{Name: name}
	if region, ok := attr["region"]; ok {
		reg, err := NewRegion(region)
		if err != nil {
			return *offerParams, err
		}
		offerParams.Region = reg
	} else {
		offerParams.Region = defaultRegion
	}
	return *offerParams, nil
}

// String returns a simple string version of the pricing
func (eo EC2Offer) String() string {
	return fmt.Sprintf("$%0.3f /hr, $%0.2f /mo", eo.HourlyPrice(), eo.HourlyPrice()*HoursPerMonth)
}

// Columns returns a slice of the column names for this type
func (eo EC2Offer) Columns() []string {
	return []string{"type", "vCPU", "Mem", "$/hr", "$/mo"}
}

// RowData returns data for this item for tablular presentation
// Should be used in concert with Columns
func (eo EC2Offer) RowData() []string {
	return []string{eo.Product.InstanceType, eo.Product.VCPU, eo.Product.Memory, fmt.Sprintf("$%0.3f", eo.Price), fmt.Sprintf("$%0.2f", eo.Price*HoursPerMonth)}
}

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
}

const summaryDBFile = "_SummaryDB.gob"

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
		return nil, fmt.Errorf("No matching records found")
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
	return &db
}
