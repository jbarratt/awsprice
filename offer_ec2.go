package awsprice

import "fmt"

// EC2OfferIndex is at the root of the EC2 Offer JSON document
type EC2OfferIndex struct {
	FormatVersion   string                `json:"formatVersion"`
	Disclaimer      string                `json:"disclaimer"`
	PublicationDate string                `json:"publicationDate"`
	Products        map[string]EC2Product `json:"products"`
	Terms           EC2Terms              `json:"terms"`
}

// EC2Product identifies a single product 'leaf' in the JSON document
type EC2Product struct {
	SKU           string  `json:"sku"`
	ProductFamily string  `json:"productFamily"`
	Attr          EC2Attr `json:"attributes"`
}

// EC2Attr identifies a selected list of useful attributes
type EC2Attr struct {
	ServiceCode       string `json:"servicecode"`
	Location          string `json:"location"`
	LocationType      string `json:"locationType"`
	InstanceType      string `json:"instanceType"`
	CurrentGeneration string `json:"currentGeneration"`
	InstanceFamily    string `json:"instanceFamily"`
	VCPU              string `json:"vcpu"`
	Memory            string `json:"memory"`
	OperatingSystem   string `json:"operatingSystem"`
	Tenancy           string `json:"tenancy"`
}

// EC2Terms tracks the various terms. For now only OnDemand (not prepaid/spot/etc) is used.
type EC2Terms struct {
	OnDemand map[string]map[string]EC2TermItem
}

// EC2TermItem is a given pricing term
type EC2TermItem struct {
	OfferTermCode   string                        `json:"offerTermCode"`
	SKU             string                        `json:"sku"`
	PriceDimensions map[string]EC2PriceDimensions `json:"priceDimensions"`
}

// EC2PriceDimensions stores various combinations of billing duration
// and currency
type EC2PriceDimensions struct {
	RateCode     string            `json:"rateCode"`
	Description  string            `json:"description"`
	Unit         string            `json:"unit"`
	PricePerUnit map[string]string `json:"pricePerUnit"`
}

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
