package utils

import (
	"bytes"
	"fmt"
	"github.com/OnsagerHe/geoip-detector/pkg/vpn"
	"log"
	"net"
	"strings"

	"golang.org/x/crypto/sha3"
)

var FolderPath *string
var BrowserPath *string

type GeoIP struct {
	Resource    EndpointMetadata
	Analyzes    []Analyze
	VPNProvider vpn.IProviderVPN
}

type Nameserver struct {
	Host *net.NS
	IPs  []net.IP
}

type EndpointMetadata struct {
	Endpoint    string
	Port        string
	Prefix      string
	Host        string
	Nameservers []Nameserver
	Cname       bool
	CnameHost   string
	Online      bool
}

type Analyze struct {
	IpDest      string
	IpSource    []string
	CountryCode string
	Hash        []byte
	Online      bool
	Nameserver  Nameserver
	Filename    string
}

// Key Method to convert the struct to a comparable string key
func (a Analyze) Key() string {
	return a.IpDest + ":" + strings.Join(a.IpSource, ",") + ":" + a.CountryCode + ":" + string(a.Hash)
}

func GetAnalyzesByHosts(analyzes []Analyze, countryCode string, hosts []string) []*Analyze {
	allKeys := make(map[string]bool)
	var list []*Analyze
	hostSet := make(map[string]bool)
	for _, host := range hosts {
		hostSet[host] = true
	}

	for i := range analyzes {
		item := &analyzes[i]
		// Filter out items with a different CountryCode or IpDest not in hosts
		if item.CountryCode != countryCode || !hostSet[item.IpDest] {
			continue
		}
		key := item.Key()
		if _, exists := allKeys[key]; !exists {
			allKeys[key] = true
			list = append(list, item)
		}
	}
	return list
}

func GetAnalyzesByCountryCode(analyzes []Analyze, countryCode string) []*Analyze {
	allKeys := make(map[string]bool)
	var list []*Analyze
	for i := range analyzes {
		item := &analyzes[i]
		// Filter out items with a different CountryCode
		if item.CountryCode != countryCode {
			continue
		}
		key := item.Key()
		if _, exists := allKeys[key]; !exists {
			allKeys[key] = true
			list = append(list, item)
		}
	}
	return list
}

// RemoveAnalyzeDuplicates Function to remove duplicates from a slice of Analyze structs
func RemoveAnalyzeDuplicates(analyzes []Analyze) []Analyze {
	allKeys := make(map[string]bool)
	var list []Analyze
	for _, item := range analyzes {
		key := item.Key()
		if _, value := allKeys[key]; !value {
			allKeys[key] = true
			list = append(list, item)
		}
	}
	return list
}

func HashByte(body []byte) []byte {
	hasher := sha3.New256()
	hasher.Write(body)
	hashSum := hasher.Sum(nil)

	return hashSum
}

func CompareHash(analyzes []Analyze) {
	firstHash := analyzes[0].Hash

	for i := range analyzes {
		log.Printf("\tip %s: %x\n", analyzes[i].IpDest, analyzes[i].Hash)
		if !bytes.Equal(analyzes[i].Hash, firstHash) {
			fmt.Printf("%s has a different hash: %x\n", analyzes[i].IpDest, analyzes[i].Hash)

		}
	}
}
