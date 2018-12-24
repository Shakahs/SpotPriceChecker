package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"strconv"
	"sync"
	"time"
)

var wg sync.WaitGroup
var resultChan = make(chan ec2.SpotPrice)
var lowestSeen ec2.SpotPrice
var searchWindows = flag.Bool("windows", false, "search for Windows pricing")
var searchString = []*string{aws.String("Linux/UNIX (Amazon VPC)")}
var verbose = flag.Bool("verbose", false, "produce verbose output")

func parseTime(layout, value string) *time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return &t
}

func determinePrice() {
	for {
		select {
		case res := <-resultChan:
			if lowestSeen.SpotPrice != nil {
				lowestSeenPrice, err := strconv.ParseFloat(*lowestSeen.SpotPrice, 32)
				if err != nil {
					fmt.Println(err)
				}

				thisPrice, err := strconv.ParseFloat(*res.SpotPrice, 32)
				if err != nil {
					fmt.Println(err)
				}
				if thisPrice < lowestSeenPrice {
					lowestSeen = res
				}
			} else {
				lowestSeen = res
			}
		}
	}
}

func getPrices(region string) {
	defer wg.Done()

	sess, err := session.NewSession(
		&aws.Config{
			Region:      aws.String(region),
			Credentials: credentials.NewSharedCredentials("", "SpotPriceChecker"),
		},
	)
	if err != nil {
		fmt.Println(err)
	}

	svc := ec2.New(sess)
	input := &ec2.DescribeSpotPriceHistoryInput{
		InstanceTypes: []*string{
			aws.String("p3.2xlarge"),
		},
		ProductDescriptions: searchString,
	}

	result, err := svc.DescribeSpotPriceHistory(input)
	if err != nil {
		fmt.Println("an error occured")
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	seen := make(map[string]bool)
	for _, val := range result.SpotPriceHistory {
		if seen[*val.AvailabilityZone] == false {
			if *verbose {
				fmt.Println(val)
			}
			resultChan <- *val
			seen[*val.AvailabilityZone] = true
		}
	}

}

func main() {
	flag.Parse()

	if *searchWindows == true {
		searchString = []*string{aws.String("Windows (Amazon VPC)")}
	}

	regionList := []string{"us-west-1", "us-west-2", "us-east-1", "us-east-2", "ca-central-1",
		"sa-east-1",
		"eu-west-1", "eu-west-2", "eu-west-3", "eu-central-1",
		"ap-south-1", "ap-northeast-1", "ap-northeast-2", "ap-southeast-1", "ap-southeast-2"}

	go determinePrice()
	for _, r := range regionList {
		wg.Add(1)
		go getPrices(r)
	}
	wg.Wait()
	fmt.Println("Lowest price:", lowestSeen)
}
