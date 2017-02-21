# AWSPrice: AWS Pricing via a simple grammar

## Goals

Make it quick and easy to figure out prices for AWS configurations.

This is done via a simple grammar, like

```
$ awsprice '2 * m4.xlarge + db.t2.medium(engine=mariadb) + elb(transfer=500GB)'
```

The goal is to have a common engine power potentially a few different interfaces:

* A command line tool `$ awsprice c3.xlarge`
* A slackbot slash command `/awsprice ebs(500G)`
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
	* With wildcard matching
	* initial slackbot deploy (with cached data)
	* 'help'
* Additional EC2 demensions (region, OS)
* Basic calculator support (+, -, parenthesis grouping)
* EBS support
* ELB support (including data transfer)
* S3 support (GB)
* RDS support (region, multi-az, engine, storage dimensions)
* Cloudfront Support (transfer, price class)
* EC2 Transit support
* S3 transit support
* 'vs' operator (comparing 2 stacks with each other)

# Internal Architecture

The CLI will need to have a fetch/process method to call.
Implicit command will be to consider argv[1] as string.

For the first pass, if the db is not available, run a fetch/process to make it
Check the db for a string.

This should use a 'OfferLookup' interface.

takes a String & optional map[string]string of attributes

returns something implementing the Offer interface

float hourly() # USD/hr
string Name
string OfferType

