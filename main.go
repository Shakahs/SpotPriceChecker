package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	//"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"sync"
	"time"
)

func parseTime(layout, value string) *time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return &t
}

var wg sync.WaitGroup

func getPrices(region string) {
	defer wg.Done()
	fmt.Println(region)
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
		//Credentials: credentials.NewStaticCredentials("ASIAI4JRMRV6S73MHELQ", "bmDL5BjEVWnRJHIY0szP7HRn0gbkgQfCKrGcWZQG", "FQoDYXdzEFYaDJ9V/+srTMy5OD4jcSKsAX26PMwtKZvHuYWEE43Kii0hTHYrIM/DEE9jOqoihpZG5caSDECFDRct9ar/+Fw7Kq8MlEtb8QKTHvNbFH0EDgZxEF9ggiDKOWFWKoI/GAC2Fu513ZWFuSS3+QbSCxNR3u5bYqFlHcghMqbQtF+lSy13JBkiYaE2IBDgvt/UPsNsy5CqRlBC7Ni+RK8Rcn70ux9IokTPCDfhMb108klrstM9v4Is09ARCvC13kso44HV0wU="),
	},
	)
	if err != nil {
		fmt.Println(err)
	}

	//creds := stscreds.NewCredentials(sess, "arn:aws:iam::130150402265:user/ec2spotpricechecker")

	svc := ec2.New(sess)
	//svc := ec2.New(sess, &aws.Config{Credentials: creds})
	//svc := ec2.New(session.New())
	input := &ec2.DescribeSpotPriceHistoryInput{
		InstanceTypes: []*string{
			aws.String("p3.2xlarge"),
		},
		ProductDescriptions: []*string{
			aws.String("Linux/UNIX (Amazon VPC)"),
		},
		//EndTime:   parseTime("2006-01-02T15:04:05Z", "2014-01-06T08:09:10Z"),
		//StartTime: parseTime("2006-01-02T15:04:05Z", "2014-01-06T07:08:09Z"),
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
	//fmt.Print(result)
	for _, val := range result.SpotPriceHistory {
		if seen[*val.AvailabilityZone] == false {
			fmt.Println(val)
			seen[*val.AvailabilityZone] = true
		}
	}

}

func main() {
	regionList := []string{"us-west-1", "us-west-2", "us-east-1", "us-east-2", "ca-central-1",
		"sa-east-1",
		"eu-west-1", "eu-west-2", "eu-west-3", "eu-central-1",
		"ap-south-1", "ap-northeast-1", "ap-northeast-2", "ap-southeast-1", "ap-southeast-2"}
	//fmt.Println(regionList)

	for _, r := range regionList {
		//fmt.Println(r)
		wg.Add(1)
		go getPrices(r)
	}
	wg.Wait()
}
