
**SpotPriceChecker** is a simple command line tool to determine the lowest spot price for an AWS p3.2xlarge GPU instance. It queries all AWS datacenters once (simultaneously), compiles the data, prints the result, and exits.

Options:
* -windows 
Search for Windows instances (defaults to Linux)
* -verbose
Print all pricing information (defaults to lowest only)

