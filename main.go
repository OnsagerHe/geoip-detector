package main

import (
	dnsutils "awesomeProject4/pkg/dns"
	httputils "awesomeProject4/pkg/http"
	"awesomeProject4/pkg/utils"
	"awesomeProject4/pkg/vpn"
	"errors"
	"flag"
	"log"
	"time"
)

var endpoint *string
var loop *uint

func init() {
	endpoint = flag.String("endpoint", "http://onsager.net", "a string")
	loop = flag.Uint("loop", 3, "a integer")
	flag.Parse()
}

func initializeResources() (utils.EndpointMetadata, []utils.Analyze, error) {
	resource := utils.EndpointMetadata{Endpoint: *endpoint}
	var analyzes []utils.Analyze

	if err := httputils.InitHTTPInformation(&resource); err != nil {
		return resource, analyzes, err
	}

	if err := dnsutils.InitNameserversInformation(&resource); err != nil {
		return resource, analyzes, err
	}

	return resource, analyzes, nil
}

func processRelaysAndDNS(vpnProvider *vpn.Mullvad, resource *utils.EndpointMetadata, analyzes *[]utils.Analyze) {
	relays := vpnProvider.ListVPN()
	count := uint(0)

	for countryCode := range relays {
		if count >= *loop {
			break
		}
		ips, err := vpnProvider.SetLocationVPN(countryCode)
		if err != nil {
			log.Printf("Error setting location VPN: %v\n", err)
			return
		}

		count++

		for _, ns := range resource.Nameservers {
			if err := dnsutils.GetIPsNameserver(&ns); err != nil {
				log.Printf("Error getting IPs for nameserver: %v\n", err)
				return
			}

			dnsutils.FilterIPv6(&ns.IPs)

			for _, ip := range ns.IPs {
				if err := vpnProvider.SetCustomDNSResolver(ip.String()); err != nil {
					log.Println("Error setting custom DNS resolver:", err)
					return
				}

				dnsutils.ProcessDNSRecords(resource, countryCode, ips, ns, ip, analyzes)
			}

			if err := vpnProvider.SetDefaultDNSResolver(); err != nil {
				log.Println("Error setting default DNS resolver:", err)
				return
			}

			*analyzes = utils.RemoveAnalyzeDuplicates(*analyzes)
			httputils.RequestEndpoints(resource, analyzes)
		}
	}
}

func run() {
	resource, analyzes, err := initializeResources()
	if err != nil {
		log.Printf("Initialization error: %v\n", err)
		return
	}

	vpnProvider := vpn.Mullvad{}
	if err := connectToVPN(&vpnProvider); err != nil {
		log.Printf("VPN connection error: %v\n", err)
		return
	}

	processRelaysAndDNS(&vpnProvider, &resource, &analyzes)
	utils.CompareHash(analyzes)
}

func connectToVPN(vpnProvider *vpn.Mullvad) error {
	if err := vpnProvider.ConnectVPN(); err != nil {
		return errors.New("cannot connect to Mullvad VPN")
	}
	time.Sleep(3 * time.Second)
	log.Println("Sleep during 3 sec ...")
	return nil
}

func main() {
	run()
}
