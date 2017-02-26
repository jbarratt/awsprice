package awsprice

// ParseInput takes a pricer and the input string and returns
// a string representation of the price
func ParseInput(pricer Pricer, input string) (string, error) {
	price, err := pricer.Get(input, PriceAttr{})
	// default region
	region, _ := NewRegion("us-west-2")
	attr := PriceAttr{Region: region}
	if err != nil {
		prices := pricer.Search(input, attr)
		if len(prices) == 0 {
			return "", err
		}
		return PriceTable(prices), nil

	}
	return price.String(), nil
}
