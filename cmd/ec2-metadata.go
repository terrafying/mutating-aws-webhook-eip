package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const (
	addr = "http://169.254.169.254/latest/meta-data/"
)

// GetMAC returns the mac address list
func GetMAC() (string, error) {
	return get("network/interfaces/macs")
}

// GetVPCID - LocalIPAddress returns the local IP address of the running instance.
func GetVPCID() (string, error) {
	if os.Getenv("VPC_ID") != "" {
		fmt.Println("Found VPC_ID in environment.")
		return os.Getenv("VPC_ID"), nil
	}

	mac, err := GetMAC()
	if err != nil {
		return "", err
	}
	mac0 := strings.Split(mac, "\n")[0]
	s, err2 := get("network/interfaces/macs/" + mac0 + "/vpc-id")
	if err2 != nil {
		return "", err
	}
	return s, err2
}

// PublicIPAddress returns the public IP address of the running instance.
func PublicIPAddress() (string, error) {
	return get("public-ipv4")
}

func get(part string) (string, error) {
	resp, err := http.Get(addr + part)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
