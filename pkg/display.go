package pkg

import (
	"fmt"
	"github.com/OnsagerHe/geoip-detector/pkg/utils"
	"github.com/fatih/color"
	"sort"
)

func countHashFrequencies(data []utils.Analyze) map[string]int {
	frequencyMap := make(map[string]int)
	for _, entry := range data {
		hashStr := string(entry.Hash)
		frequencyMap[hashStr]++
	}
	return frequencyMap
}

func sortByHashFrequency(data []utils.Analyze, frequencyMap map[string]int) []utils.Analyze {
	sort.Slice(data, func(i, j int) bool {
		hashI := string(data[i].Hash)
		hashJ := string(data[j].Hash)
		return frequencyMap[hashI] < frequencyMap[hashJ]
	})
	return data
}

func SortResult(data []utils.Analyze) {
	frequencyMap := countHashFrequencies(data)
	sortedData := sortByHashFrequency(data, frequencyMap)
	displayInformation(sortedData)
}

func displayInformation(data []utils.Analyze) {
	for _, entry := range data {
		var statusMsg string
		if entry.Online {
			statusMsg = color.GreenString("[+] Status: online")
		} else {
			statusMsg = color.RedString("[-] Status: offline")
		}

		fmt.Printf("%s\n", statusMsg)
		fmt.Printf("IP Source: %v\n", entry.IpSource)
		fmt.Printf("IP Dest: %s\n", entry.IpDest)
		fmt.Printf("Hash: %x\n", entry.Hash)
		// fmt.Printf("Hour UTC: %s\n", "N/A") // TODO: Maybe add timestamp
		fmt.Printf("Country Code IP Source: %s\n", entry.CountryCode)
		fmt.Printf("Filename screenshot: %s\n", entry.Filename)
		fmt.Printf("Nameserver requested: %s\n\n", entry.Nameserver.IPs)
	}
}
