package dns

import (
	"awesomeProject4/pkg/utils"
	"fmt"
	"github.com/miekg/dns"
	"log"
	"net"
	"strings"
)

func InitNameserversInformation(resource *utils.EndpointMetadata) error {
	err := checkCNAME(resource)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return err
	}
	// Get nameserver from domain
	err = getNs(resource)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return err
	}

	return nil
}

func getNs(resource *utils.EndpointMetadata) error {
	var nsRecords []*net.NS
	var subDomain string

	parts := strings.Split(resource.CnameHost, ".")

	for i := 0; i < len(parts)-1; i++ {
		subDomain = strings.Join(parts[i:], ".")
		nsRecords, _ = net.LookupNS(subDomain)
		if len(nsRecords) > 0 {
			break
		}
	}

	for _, ns := range nsRecords {
		resource.Nameservers = append(resource.Nameservers, utils.Nameserver{Host: ns})
	}

	//for _, ns := range resource.Nameservers {
	//	log.Println("value", ns.Host)
	//}
	return nil
}

func GetIPsNameserver(nameserver *utils.Nameserver) error {
	var err error
	nameserver.IPs, err = net.LookupIP(nameserver.Host.Host)
	if err != nil {
		return err
	}

	return nil
}

func FilterIPv6(addresses *[]net.IP) {
	var filtered []net.IP

	for _, addr := range *addresses {
		if addr.To4() != nil {
			filtered = append(filtered, addr)
		}
	}

	*addresses = filtered
}

func filterIPv6Str(addresses *[]string) {
	var filtered []string

	for _, addr := range *addresses {
		ip := net.ParseIP(addr)
		if ip == nil {
			log.Fatalln("Error: cannot parse ip addresses")
			return
		}
		if ip.To4() != nil {
			filtered = append(filtered, addr)
		}
	}

	*addresses = filtered
}

func ProcessDNSRecords(resource *utils.EndpointMetadata, countryCode string, ips []string, ns utils.Nameserver, ip net.IP, analyzes *[]utils.Analyze) {
	var analyze utils.Analyze
	host, err := net.LookupHost(resource.CnameHost)
	if err != nil {
		log.Println("Error looking up host:", err)
		return
	}
	filterIPv6Str(&host)

	for _, h := range host {
		analyze.IpDest = h
		analyze.CountryCode = countryCode
		analyze.IpSource = ips
		analyze.Nameserver = utils.Nameserver{Host: ns.Host, IPs: []net.IP{ip}}
		*analyzes = append(*analyzes, analyze)
	}
}

func InitDNSClient() *dns.Client {
	return new(dns.Client)
}

func InitDNSRequest(domain string, typeRecords uint16) *dns.Msg {
	dnsMessage := new(dns.Msg)
	dnsMessage.SetQuestion(dns.Fqdn(domain), typeRecords)
	dnsMessage.RecursionDesired = true
	return dnsMessage
}

func checkCNAME(resource *utils.EndpointMetadata) error {
	var err error

	resource.CnameHost, err = net.LookupCNAME(resource.Host)
	if err != nil {
		fmt.Println("Error looking up CNAME:", err)
		return err
	}

	return nil
}