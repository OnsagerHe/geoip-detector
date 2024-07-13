package utils

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strings"

	"golang.org/x/crypto/sha3"
)

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
}

type Analyze struct {
	IpDest      string
	IpSource    []string
	CountryCode string
	Hash        []byte
	Online      bool
	Nameserver  Nameserver
}

// Key Method to convert the struct to a comparable string key
func (a Analyze) Key() string {
	return a.IpDest + ":" + strings.Join(a.IpSource, ",") + ":" + a.CountryCode + ":" + string(a.Hash)
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
	// Get the first hash value
	firstHash := analyzes[0].Hash

	// Iterate over the rest of the analyses array
	for i := range analyzes {
		log.Printf("\tip %s: %x\n", analyzes[i].IpDest, analyzes[i].Hash)
		if !bytes.Equal(analyzes[i].Hash, firstHash) {
			fmt.Printf("%s has a different hash: %x\n", analyzes[i].IpDest, analyzes[i].Hash)

		}
	}
}
