package awsprice

// ParseInput takes a pricer and the input string and returns
// a string representation of the price
func ParseInput(pricer Pricer, input string) (string, error) {

	attr := make(map[string]string)
	offer, err := pricer.Get(input, attr)
	if err != nil {
		prices := pricer.Search(input, attr)
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
