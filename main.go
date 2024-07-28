package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/OnsagerHe/geoip-detector/pkg"
	"log"
	"time"

	dnsutils "github.com/OnsagerHe/geoip-detector/pkg/dns"
	httputils "github.com/OnsagerHe/geoip-detector/pkg/http"
	"github.com/OnsagerHe/geoip-detector/pkg/utils"
	"github.com/OnsagerHe/geoip-detector/pkg/vpn"
)

var endpoint *string
var loop *uint

func init() {
	endpoint = flag.String("endpoint", "http://onsager.net", "endpoint to test")
	loop = flag.Uint("loop", 3, "number of localizations you wish to use")
	flag.Parse()
}

func initializeResources() (*utils.GeoIP, error) {
	res := utils.GeoIP{
		Resource:    utils.EndpointMetadata{Endpoint: *endpoint},
		Analyzes:    nil,
		VPNProvider: vpn.Mullvad{},
	}

	if err := httputils.InitHTTPInformation(&res.Resource); err != nil {
		return nil, err
	}

	if err := dnsutils.InitNameserversInformation(&res.Resource); err != nil {
		return nil, err
	}

	return &res, nil
}

func processRelaysAndDNS(res *utils.GeoIP) {
	relays := res.VPNProvider.ListVPN()
	count := uint(0)

	for countryCode := range relays {
		if count >= *loop {
			break
		}
		// TODO: remove debug condition
		if count == 0 {
			countryCode = "al"
		}

		ips, err := res.VPNProvider.SetLocationVPN(countryCode)
		if err != nil {
			log.Printf("Error setting location VPN: %v\n", err)
			return
		}

		count++

		for _, ns := range res.Resource.Nameservers {
			if err := dnsutils.GetIPsNameserver(&ns); err != nil {
				log.Printf("Error getting IPs for nameserver: %v\n", err)
				return
			}

			dnsutils.FilterIPv6(&ns.IPs)

			log.Println("nbr ns IPS", ns.IPs)
			for _, ip := range ns.IPs {
				if err := res.VPNProvider.SetCustomDNSResolver(ip.String()); err != nil {
					log.Println("Error setting custom DNS resolver:", err)
					return
				}
				hosts := dnsutils.ProcessDNSRecords(res, countryCode, ips, ns, ip)
				a := utils.GetAnalyzesByHosts(res.Analyzes, countryCode, hosts)
				httputils.RequestSpecificEndpoints(res, a)
				httputils.TakeScreenshotByCountryCode(res, a)
			}
		}
	}
	if err := res.VPNProvider.SetDefaultDNSResolver(); err != nil {
		log.Println("Error setting default DNS resolver:", err)
	}
	//res.Analyzes = utils.RemoveAnalyzeDuplicates(res.Analyzes)
}

func run() {
	res, err := initializeResources()
	if err != nil {
		log.Printf("Initialization error: %v\n", err)
		return
	}

	vpnProvider := vpn.Mullvad{}
	if err := connectToVPN(&vpnProvider); err != nil {
		log.Printf("VPN connection error: %v\n", err)
		return
	}

	processRelaysAndDNS(res)
	utils.CompareHash(res.Analyzes)
	pkg.SortResult(res.Analyzes)
}

func connectToVPN(vpnProvider *vpn.Mullvad) error {
	if err := vpnProvider.ConnectVPN(); err != nil {
		return errors.New("cannot connect to Mullvad VPN")
	}
	time.Sleep(3 * time.Second)
	fmt.Print("VPN connection")
	for i := 0; i < 3; i++ {
		fmt.Print(".")
		time.Sleep(1 * time.Second)
	}
	fmt.Println("")
	return nil
}

func main() {
	run()
}
