# AWSPrice: AWS Pricing via a simple grammar

## Installation

You should be able to

	go get -u github.com/jbarratt/awsprice/...

and end up with an `awsprice` binary. (Which, as of now, has very limited functionality.)

## Goals

Make it quick and easy to figure out prices for AWS configurations.

This is done via a simple grammar, like

```
$ awsprice '2 * m4.xlarge + db.t2.medium(engine=mariadb) + elb(transfer=500GB)'
```

(Warning, not yet implemented.)

The goal is to have a common engine power potentially a few different interfaces:

* A command line tool `$ awsprice c3.xlarge`
* A slackbot: `@awsprice ebs(500G)`
* A single page webapp, with pretty/shareable HTML results

## Implementation

The tools are built on the AWS Price List API.

Internally, it goes through a few phases:

1. Fetching raw JSON offer data
2. Processing that data into a format optimized for lookups
3. Querying the data via the DSL/Grammar
4. Presenting the results in various ways

## Feature Set (Planned Delivery Order)

* Initial CLI framework ✔ 
	* awsprice fetch ✔
	* awsprice process (optional -o 'db file') ✔
* Simple lookup of EC2 information (only single region, no options) ✔
	* via CLI ✔
	* With wildcard matching ✔
	* initial slackbot deploy (with cached data) ✔ ([awspricebot|http://github.com/jbaratt/awspricebot])
	* 'help' ✔
* Additional EC2 dimensions (region) ✔
* Basic RDS (region, multi-az, engine) ✔
* Basic calculator support (+, -, parenthesis grouping)
* EBS support
* ELB support (including data transfer)
* S3 support (GB)
* RDS support (region, multi-az, engine, storage dimensions)
* Cloudfront Support (transfer, price class)
* EC2 Transit support
* S3 transit support
* 'vs' operator (comparing 2 stacks with each other)
* EC2 OS


# Internal Architecture

At the top level, the parser will identify an 'Offer' token, with a set of optional k=v arguments.

	db.r3.xlarge(engine=mariadb)

There is a high level dispatch table:
	- given an identifier name
	- return the offer type

The dispatch table will then call the proper `New{}Offer` method, passing the arguments in.
It will 'mix in' any of the global arguments, as well. For example

	2 x m4.xlarge(os=Windows)  region=us-west-1

would 

* Look up m4.xlarge and discover it's an EC2 type
* Construct a new EC2Offer, with {'os': 'Windows', 'region': 'us-west-1'} as arguments

	type OfferType int
	const (
		EC2 OfferType = iota
		RDS
	)

	// when a new one is added link it in here
	map[string]OfferType

	// and construct a NewEC2OfferParam(k_v)
	EC2OfferParam {
		Name
		Region
		Os
	}
	map[EC2OfferParam]EC2Offer
	EC2Offer


The Offers all implement the Offer interface, which has some standard methods

	Type() string
	Description() string
	Hourly() float64

This allows them to be displayed & totaled as needed.

	StoreEC2(name string, params map[string]string, EC2Offer)
	StoreRDS(name string, params map[string]string, RDSOffer)
	Load(name string, params map[string]string) Offer
	Search(name string, params map[string]string) []Offer




