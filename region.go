package awsprice

import "fmt"

// Region is a type identifying AWS Regions.
type Region string

var codeToRegion map[string]string
var regionToCode map[string]string

func init() {
	regionToCode = map[string]string{
		"Asia Pacific (Mumbai)":     "ap-south-1",
		"Asia Pacific (Seoul)":      "ap-northeast-2",
		"Asia Pacific (Singapore)":  "ap-southeast-1",
		"Asia Pacific (Sydney)":     "ap-southeast-2",
		"Asia Pacific (Tokyo)":      "ap-northeast-1",
		"Canada (Central)":          "ca-central-1",
		"EU (Frankfurt)":            "eu-central-1",
		"EU (Ireland)":              "eu-west-1",
		"EU (London)":               "eu-west-2",
		"South America (Sao Paulo)": "sa-east-1",
		"US East (N. Virginia)":     "us-east-1",
		"US East (Ohio)":            "us-east-2",
		"US West (N. California)":   "us-west-1",
		"US West (Oregon)":          "us-west-2",
	}
	codeToRegion = make(map[string]string)
	for region, code := range regionToCode {
		codeToRegion[code] = region
	}
}

// NewRegion returns a region type for any valid AWS region identifier
func NewRegion(given string) (Region, error) {
	if _, ok := regionToCode[given]; ok {
		return Region(given), nil
	}

	if region, ok := codeToRegion[given]; ok {
		return Region(region), nil
	}

	return Region(""), fmt.Errorf("Invalid Region")
}
