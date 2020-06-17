// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0
// snippet-start:[ec2.go.allocate_address]
package main

// snippet-start:[ec2.go.allocate_address.import]
import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

// snippet-end:[ec2.go.allocate_address.import]

// AllocateAndAssociate allocates a VPC Elastic IP address with an instance.
// Inputs:
//     svc is an Amazon EC2 service client
//     instanceID is the ID of the instance
// Output:
//     If success, information about the allocation and association and nil
//     Otherwise, two nils and an error from the call to AllocateAddress or AssociateAddress
func AllocateAndAssociate(svc ec2iface.EC2API, instanceID *string) (*ec2.AllocateAddressOutput, *ec2.AssociateAddressOutput, error) {
	// snippet-start:[ec2.go.allocate_address.allocate]
	allocRes, err := svc.AllocateAddress(&ec2.AllocateAddressInput{
		Domain: aws.String("vpc"),
	})
	// snippet-end:[ec2.go.allocate_address.allocate]
	if err != nil {
		return nil, nil, err
	}

	// snippet-start:[ec2.go.allocate_address.associate]
	assocRes, err := svc.AssociateAddress(&ec2.AssociateAddressInput{
		AllocationId: allocRes.AllocationId,
		InstanceId:   instanceID,
	})
	// snippet-end:[ec2.go.allocate_address.associate]
	if err != nil {
		return nil, nil, err
	}

	return allocRes, assocRes, nil
}

func GetAddresses(sess *session.Session) (*ec2.DescribeAddressesOutput, error) {
	// snippet-start:[ec2.go.describe_addresses.call]
	svc := ec2.New(sess)

	result, err := svc.DescribeAddresses(&ec2.DescribeAddressesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("domain"),
				Values: aws.StringSlice([]string{"vpc"}),
			},
		},
	})
	// snippet-end:[ec2.go.describe_addresses.call]
	if err != nil {
		return nil, err
	}

	return result, nil
}

func GetAddressesForIP(sess *session.Session, ip []string) (*ec2.DescribeAddressesOutput, error) {
	// snippet-start:[ec2.go.describe_addresses.call]
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
	// snippet-end:[ec2.go.describe_addresses.call]
	if err != nil {
		return nil, err
	}

	return result, nil
}

// func main() {
//     // snippet-start:[ec2.go.describe_addresses.session]
//     sess := session.Must(session.NewSessionWithOptions(session.Options{
//         SharedConfigState: session.SharedConfigEnable,
//     }))
//     // snippet-end:[ec2.go.describe_addresses.session]
//
//     result, err := GetAddresses(sess)
//     if err != nil {
//         fmt.Println("Got an error retrieving the Elastic IP addresses")
//         return
//     }
//
//     // snippet-start:[ec2.go.describe_addresses.display]
//     for _, addr := range result.Addresses {
//         fmt.Println("IP address:   ", *addr.PublicIp)
//         fmt.Println("Allocation ID:", *addr.AllocationId)
//         if addr.InstanceId != nil {
//             fmt.Println("Instance ID:  ", *addr.InstanceId)
//         }
//     }
//     // snippet-end:[ec2.go.describe_addresses.display]
// }
