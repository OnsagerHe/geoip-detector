package retriever

import (
	"log"

	"github.com/OnsagerHe/geoip-detector/pkg"
	dnsutils "github.com/OnsagerHe/geoip-detector/pkg/dns"
	httputils "github.com/OnsagerHe/geoip-detector/pkg/http"
	"github.com/OnsagerHe/geoip-detector/pkg/utils"
	pb "github.com/OnsagerHe/geoip-detector/proto/gen"
)

type Utils struct {
	Loop uint8
}

type Retriever struct {
	Process *utils.GeoIP
	Utils   *Utils
}

func Init(p *utils.GeoIP, loopValue uint8) *Retriever {
	return &Retriever{
		Process: p,
		Utils: &Utils{
			Loop: loopValue,
		},
	}
}

func (p Retriever) CheckEndpoint() (*[]pb.PutEndpointResponse, error) {
	err := p.initializeResources()
	if err != nil {
		log.Printf("Initialization error: %v\n", err)
		return nil, nil
	}

	p.processRelaysAndDNS()
	utils.CompareHash(p.Process.Analyzes)
	pkg.SortResult(p.Process.Analyzes)
	if err := p.Process.VPNProvider.SetDefaultDNSResolver(); err != nil {
		log.Println("Error setting default DNS resolver:", err)
	}

	return pkg.DisplayInformation(p.Process.Analyzes), nil
}

func (p Retriever) initializeResources() error {
	if err := httputils.InitHTTPInformation(&p.Process.Resource); err != nil {
		return err
	}

	if err := dnsutils.InitNameserversInformation(&p.Process.Resource); err != nil {
		return err
	}

	return nil
}

func (p Retriever) processRelaysAndDNS() {
	relays := p.Process.VPNProvider.ListVPN()
	count := uint8(0)

	for countryCode := range relays {
		if count >= p.Utils.Loop {
			break
		}

		ips, err := p.Process.VPNProvider.SetLocationVPN(countryCode)
		if err != nil {
			log.Printf("Error setting location VPN: %v\n", err)
			return
		}

		count++

		for _, ns := range p.Process.Resource.Nameservers {
			if err := dnsutils.GetIPsNameserver(&ns); err != nil {
				log.Printf("Error getting IPs for nameserver: %v\n", err)
				continue
			}

			dnsutils.FilterIPv6(&ns.IPs)

			// TODO: add debug print with level log
			//log.Println("nbr ns IPS", ns.IPs)
			for _, ip := range ns.IPs {
				if err := p.Process.VPNProvider.SetCustomDNSResolver(ip.String()); err != nil {
					log.Println("Error setting custom DNS resolver:", err)
					continue
				}
				hosts := dnsutils.ProcessDNSRecords(p.Process, countryCode, ips, ns, ip)
				a := utils.GetAnalyzesByHosts(p.Process.Analyzes, countryCode, hosts)
				httputils.RequestSpecificEndpoints(p.Process, a)
				if *utils.Screenshot {
					httputils.TakeScreenshotByCountryCode(p.Process, a)
				}
			}
		}
	}
	if err := p.Process.VPNProvider.SetDefaultDNSResolver(); err != nil {
		log.Println("Error setting default DNS resolver:", err)
	}
	//res.Analyzes = utils.RemoveAnalyzeDuplicates(res.Analyzes)
}
