package awsprice

// ParseInput takes a pricer and the input string and returns
// a string representation of the price
func ParseInput(pricedb PriceDB, input string) (string, error) {

	attr := make(map[string]string)
	offer, err := pricedb.Get(input, attr)
	if err != nil {
		prices := pricedb.Search(input, attr)
		if len(prices) == 0 {
			return "", err
		}
		if len(prices) == 1 {
			return prices[0].String(), nil
		}
		return PriceTable(prices), nil

	}
	return offer.String(), nil
}
