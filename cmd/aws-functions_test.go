package main

// snippet-start:[ec2.go.allocate_address.import]
import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
)

func TestGetAddressesForIP(t *testing.T) {
	if os.Getenv("AWS_PROFILE") == "" {
		t.Log("AWS_PROFILE not set!")
		os.Setenv("AWS_PROFILE", "brivo-int-account")
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	result, err := GetAddressesForIP(sess, []string{"3.208.198.91"})
	if err != nil {
		t.Error("Got an error retrieving the Elastic IP addresses")
	} else {

		for _, addr := range result.Addresses {
			t.Log("IP address:   ", *addr.PublicIp)
			t.Log("Allocation ID:", *addr.AllocationId)
			if addr.InstanceId != nil {
				t.Log("Instance ID:  ", *addr.InstanceId)
			}
		}
	}

	// Empty result
	result2, err2 := GetAddressesForIP(sess, []string{"3.208.198.99"})
	if err2 != nil {
		t.Error("Got an error retrieving the Elastic IP addresses")
		t.Error(err2)
		return
	}
	t.Log("Result2: ", result2)

	if result2.Addresses == nil {
		t.Log("Result2 empty!!")
	} else {
		t.Error("Expected empty result for test2")
	}
}

// func TestAllocateIP(t *testing.T) {
// 	if os.Getenv("AWS_PROFILE") == "" {
// 		t.Log("AWS_PROFILE not set!")
// 		os.Setenv("AWS_PROFILE", "brivo-int-account")
// 	}
// 	sess := session.Must(session.NewSessionWithOptions(session.Options{
// 		SharedConfigState: session.SharedConfigEnable,
// 	}))
//
// 	// BYOIP Cidr for int account: 64.35.172.0/24
// 	result, err := AllocateIP(sess, "64.35.172.5")
// 	if err != nil {
// 		t.Error("Got an error retrieving the Elastic IP addresses")
// 		if aerr, ok := err.(awserr.Error); ok {
// 			switch aerr.Code() {
// 			default:
// 				glog.Error(aerr.Error())
// 			}
// 		} else {
// 			// Print the error, cast err to awserr.Error to get the Code and
// 			// Message from an error.
// 			glog.Error(err.Error())
// 		}
// 	} else {
// 		if result.PublicIp != nil {
// 			t.Logf("Successful IP Assignment: %s", *result.PublicIp)
// 		}
// 		t.Log(result)
// 	}
// }

func TestGetAddressOrAllocate(t *testing.T) {
	if os.Getenv("AWS_PROFILE") == "" {
		t.Log("AWS_PROFILE not set!")
		os.Setenv("AWS_PROFILE", "sandbox")
	}

	// BYOIP Cidr for int account: 64.35.172.0/24
	// for SDI account: 64.35.174.0/24
	result, err := GetAddressOrAllocate([]string{"64.35.174.6", "64.35.174.7"})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				t.Error(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			t.Error(err.Error())
		}
	} else {
		t.Log(result)
	}
}
