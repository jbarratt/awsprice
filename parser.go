package awsprice

import "fmt"

// ParseInput takes a pricer and the input string and returns
// a string representation of the price
func ParseInput(pricer Pricer, input string) (string, error) {
	price, err := pricer.Get(input, PriceAttr{})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("$%0.3f /hr, $%0.2f /mo", price.Price, price.Price*730), nil
}
