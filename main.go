package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

var (
	hostname     = flag.String("hostname", "", "hostname (A record) to update")
	hostedZoneID = flag.String("hosted-zone-id", "", "hosted zone id")
)

func main() {
	flag.Parse()

	if *hostname == "" || *hostedZoneID == "" {
		log.Fatalf("usage: %s -hostname <A RECORD> -hosted-zone-id <ZONE ID>", os.Args[0])
	}

	sess, err := session.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	svc := route53.New(sess)

	resp, err := http.Get("https://checkip.amazonaws.com")
	if err != nil {
		panic(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	ip := strings.TrimSpace(string(body))

	fmt.Println(ip)

	ips, err := net.LookupIP(*hostname)
	if err != nil {
		panic(err)
	}

	for _, existingIP := range ips {
		fmt.Println(existingIP)
		if existingIP.String() == ip {
			fmt.Println("IP hasn't changed")
			os.Exit(0)
		}
	}

	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: hostname,
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(ip),
							},
						},
						TTL:  aws.Int64(60),
						Type: aws.String("A"),
					},
				},
			},
		},
		HostedZoneId: hostedZoneID,
	}
	result, err := svc.ChangeResourceRecordSets(input)
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
}
