package awsprice

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
	SKU           string               `json:"sku"`
	ProductFamily string               `json:"productFamily"`
	Attr          EC2ProductAttributes `json:"attributes"`
}

// EC2ProductAttributes identifies a selected list of useful attributes
type EC2ProductAttributes struct {
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
