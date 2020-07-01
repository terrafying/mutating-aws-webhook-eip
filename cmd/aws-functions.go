package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
)

// AllocateIP allocates a VPC Elastic IP address
// Inputs:
//     svc is an Amazon EC2 service client
//     instanceID is the ID of the instance
// Output:
//     If success, information about the allocation and association and nil
//     Otherwise, two nils and an error from the call to AllocateAddress or AssociateAddress
func AllocateIP(sess *session.Session, ip string) (*ec2.AllocateAddressOutput, error) {
	svc := ec2.New(sess)

	// pools, err := svc.DescribePublicIpv4Pools(&ec2.DescribePublicIpv4PoolsInput{})
	// poolid := ""
	// if len(pools.PublicIpv4Pools) > 0 {
	// 	poolid = *pools.PublicIpv4Pools[0].PoolId
	// }
	allocRes, err := svc.AllocateAddress(&ec2.AllocateAddressInput{
		Domain:  aws.String("vpc"),
		Address: aws.String(ip),
	})

	if err != nil {
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
	}

	return allocRes, nil
}

// PrintAWSError
// Prints an aws-type error (they're special)
func PrintAWSError(err error) {
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				glog.Error(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			glog.Error(err.Error())
		}
	}
}

// GetAddresses returns all VPC Elastic IP addresses in the region.
// Inputs:
//     sess is an Amazon EC2 service client
// Output:
//     If success, information about the allocation+association and nil
func GetAddresses(sess *session.Session) (*ec2.DescribeAddressesOutput, error) {
	svc := ec2.New(sess)

	result, err := svc.DescribeAddresses(&ec2.DescribeAddressesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("domain"),
				Values: aws.StringSlice([]string{"vpc"}),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetAddressesForIP Returns IP address information for given IP.  For example result.[].AllocationID.
func GetAddressesForIP(sess *session.Session, ip []string) (*ec2.DescribeAddressesOutput, error) {
	svc := ec2.New(sess)

	result, err := svc.DescribeAddresses(&ec2.DescribeAddressesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("domain"),
				Values: aws.StringSlice([]string{"vpc"}),
			},
			{
				Name:   aws.String("public-ip"),
				Values: aws.StringSlice(ip),
			},
		},
	})
	if err != nil {
		glog.Error("Got an error retrieving the Elastic IP addresses")
		return nil, err
	}

	return result, nil
}

// GetAddressOrAllocate will return *ec2.DescribeAddressesOutput
// Either the address already exists, so we return the object
// or, it doesn't exist, so we make it.  Then return that.
func GetAddressOrAllocate(ip []string) (*ec2.DescribeAddressesOutput, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// check each ip, determine if already allocated
	// if not, allocate it
	for _, ipaddr := range ip {
		result, err := GetAddressesForIP(sess, []string{ipaddr})
		if err != nil {
			glog.Error("Got an error retrieving the Elastic IP addresses: ", err)
			return nil, err
		}
		if result.Addresses == nil {
			ipresult, iperr := AllocateIP(sess, ipaddr)
			if iperr != nil {
				PrintAWSError(iperr)
				return nil, iperr
			}
			glog.Info("AllocateIP result: ")
			glog.Info(ipresult)
		}
	}

	// Finally, get IP info and return it
	result, err := GetAddressesForIP(sess, ip)
	if err != nil {
		glog.Error("Got an error retrieving the Elastic IP addresses: ", err)
		return nil, err
	}
	return result, nil
}
