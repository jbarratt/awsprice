package awsprice

import "fmt"

// RDSOfferIndex is at the root of the RDS Offer JSON document
type RDSOfferIndex struct {
	FormatVersion   string                `json:"formatVersion"`
	Disclaimer      string                `json:"disclaimer"`
	PublicationDate string                `json:"publicationDate"`
	Products        map[string]RDSProduct `json:"products"`
	Terms           RDSTerms              `json:"terms"`
}

// RDSProduct identifies a single product 'leaf' in the JSON document
type RDSProduct struct {
	SKU           string  `json:"sku"`
	ProductFamily string  `json:"productFamily"`
	Attr          RDSAttr `json:"attributes"`
}

// RDSAttr identifies a selected list of useful attributes
type RDSAttr struct {
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
	DatabaseEngine    string `json:"databaseEngine"`
	DeploymentOption  string `json:"deploymentOption"`
}

// RDSTerms tracks the various terms. For now only OnDemand (not prepaid/spot/etc) is used.
type RDSTerms struct {
	OnDemand map[string]map[string]RDSTermItem
}

// RDSTermItem is a given pricing term
type RDSTermItem struct {
	OfferTermCode   string                        `json:"offerTermCode"`
	SKU             string                        `json:"sku"`
	PriceDimensions map[string]RDSPriceDimensions `json:"priceDimensions"`
}

// RDSPriceDimensions stores various combinations of billing duration
// and currency
type RDSPriceDimensions struct {
	RateCode     string            `json:"rateCode"`
	Description  string            `json:"description"`
	Unit         string            `json:"unit"`
	PricePerUnit map[string]string `json:"pricePerUnit"`
}

// RDSOffer The product/price details for a given RDS Offering
type RDSOffer struct {
	Product RDSAttr
	Price   float64
}

// Name returns the RDS instance type
func (ro RDSOffer) Name() string {
	return ro.Product.InstanceType
}

// HourlyPrice returns the fractional dollars per hour
func (ro RDSOffer) HourlyPrice() float64 {
	return ro.Price
}

// Type always returns RDS
func (ro RDSOffer) Type() OfferType {
	return RDS
}

// RDSOfferParam stores the unique factors that determine an RDS Offer
type RDSOfferParam struct {
	DatabaseEngine   string
	DeploymentOption string
	Region           Region
	Name             string
}

// NewRDSOfferParam constructs an RDS offer from a name & attributes
func NewRDSOfferParam(name string, attr map[string]string) (RDSOfferParam, error) {
	offerParams := &RDSOfferParam{Name: name}
	if region, ok := attr["region"]; ok {
		reg, err := NewRegion(region)
		if err != nil {
			return *offerParams, err
		}
		offerParams.Region = reg
	} else {
		offerParams.Region = defaultRegion
	}
	if engine, ok := attr["engine"]; ok {
		offerParams.DatabaseEngine = engine
	} else {
		offerParams.DatabaseEngine = "MySQL"
	}
	if deployment, ok := attr["deployment"]; ok {
		offerParams.DeploymentOption = deployment
	} else {
		offerParams.DeploymentOption = "Multi-AZ"
	}
	return *offerParams, nil
}

// String returns a simple string version of the pricing
func (ro RDSOffer) String() string {
	return fmt.Sprintf("$%0.3f /hr, $%0.2f /mo", ro.HourlyPrice(), ro.HourlyPrice()*HoursPerMonth)
}

// Columns returns a slice of the column names for this type
func (ro RDSOffer) Columns() []string {
	return []string{"type", "vCPU", "Mem", "Engine", "Deployment", "$/hr", "$/mo"}
}

// RowData returns data for this item for tablular presentation
// Should be used in concert with Columns
func (ro RDSOffer) RowData() []string {
	return []string{ro.Product.InstanceType, ro.Product.VCPU, ro.Product.Memory, ro.Product.DatabaseEngine, ro.Product.DeploymentOption, fmt.Sprintf("$%0.3f", ro.Price), fmt.Sprintf("$%0.2f", ro.Price*HoursPerMonth)}
}
