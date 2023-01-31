//usr/bin/go run $0 $@ ; exit
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/namedotcom/go/namecom"
)

func main() {
	username := os.Getenv("NAMECOM_USERNAME")
	apiToken := os.Getenv("NAMECOM_API_TOKEN")
	domainName := os.Getenv("NAMECOM_DOMAIN_NAME")
	host := os.Getenv("NAMECOM_DOMAIN_HOST")

	if len(username) == 0 || len(apiToken) == 0 || len(domainName) == 0 || len(host) == 0 {
		fmt.Println("NAMECOM environment variables missing")
		os.Exit(1)
	}

	nc := namecom.New(username, apiToken)

	listRecordsRequest := namecom.ListRecordsRequest{
		DomainName: domainName,
		PerPage:    100,
		Page:       1,
	}
	response, err := nc.ListRecords(&listRecordsRequest)
	if err != nil {
		log.Fatal(err)
	}

	var ourRecord namecom.Record
	for _, record := range response.Records {
		if record.Type == "A" && record.Host == host {
			ourRecord = *record
			break
		}
	}

	ip, err := getIP()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("IP", ip)

	if ourRecord.Answer == ip {
		fmt.Println("No update required")
		os.Exit(0)
	}

	ourRecord.Answer = ip
	_, err = nc.UpdateRecord(&ourRecord)
	if err != nil {
		log.Fatal(err)
	}
}

type ipAddress struct {
	IP string `json:"ip"`
}

func getIP() (string, error) {
	url := "https://api.ipify.org?format=json"
	httpClient := http.Client{
		Timeout: time.Second * 20, // Maximum of 20 secs
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	res, getErr := httpClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	ip := ipAddress{}
	jsonErr := json.Unmarshal(body, &ip)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return ip.IP, nil
}
