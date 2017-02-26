package awsprice

// ParseInput takes a pricer and the input string and returns
// a string representation of the price
func ParseInput(pricer Pricer, input string) (string, error) {
	price, err := pricer.Get(input, PriceAttr{})
	if err != nil {
		prices := pricer.Search(input, PriceAttr{})
		if len(prices) == 0 {
			return "", err
		}
		return PriceTable(prices), nil

	}
	return price.String(), nil
}
