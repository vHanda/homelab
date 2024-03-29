package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go"
)

func main() {
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	if apiToken == "" {
		log.Fatal("CLOUDFLARE_API_TOKEN not set")
	}

	domainName := os.Getenv("HOMELAB_DOMAIN_NAME")
	if domainName == "" {
		log.Fatal("HOMELAB_DOMAIN_NAME not set")
	}

	api, err := cloudflare.NewWithAPIToken(apiToken)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	zoneName := strings.Join(strings.Split(domainName, ".")[1:], ".")
	zoneID, err := api.ZoneIDByName(zoneName)
	if err != nil {
		log.Fatal(err)
	}
	zoneC := cloudflare.ZoneIdentifier(zoneID)

	records, _, err := api.ListDNSRecords(ctx, zoneC, cloudflare.ListDNSRecordsParams{})
	if err != nil {
		fmt.Println(err)
		return
	}

	dnsRecord := dnsRecordForName(records, domainName)
	if dnsRecord == nil {
		fmt.Println("DNS record not found")
		os.Exit(1)
	}

	ip, err := getIP()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("IP", ip)

	if dnsRecord.Content == ip {
		fmt.Println("No update required")
		os.Exit(0)
	}

	err = api.UpdateDNSRecord(ctx, zoneC, cloudflare.UpdateDNSRecordParams{
		ID:      dnsRecord.ID,
		Type:    dnsRecord.Type,
		Name:    dnsRecord.Name,
		Content: ip,
		TTL:     dnsRecord.TTL,
	})
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

func dnsRecordForName(records []cloudflare.DNSRecord, name string) *cloudflare.DNSRecord {
	for _, r := range records {
		if r.Name == name {
			return &r
		}
	}
	return nil
}
