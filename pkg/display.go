package pkg

import (
	"fmt"
	"sort"

	"github.com/OnsagerHe/geoip-detector/pkg/utils"
	pb "github.com/OnsagerHe/geoip-detector/proto/gen"
	"github.com/fatih/color"
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
	_ = sortByHashFrequency(data, frequencyMap)
}

func DisplayInformation(data []utils.Analyze) *[]pb.PutEndpointResponse {
	metadata := []pb.PutEndpointResponse{}
	var status string

	for _, entry := range data {
		var statusMsg string
		if entry.Online {
			status = "online"
			statusMsg = color.GreenString("[+] Status: online")
		} else {
			status = "offline"
			statusMsg = color.RedString("[-] Status: offline")
		}
		metadata = append(metadata, pb.PutEndpointResponse{
			Ip:       entry.IpDest,
			Status:   status,
			Filname:  entry.Filename,
			HashFile: string(entry.Hash),
		})

		fmt.Printf("%s\n", statusMsg)
		fmt.Printf("IP Source: %v\n", entry.IpSource)
		fmt.Printf("IP Dest: %s\n", entry.IpDest)
		fmt.Printf("Hash: %x\n", entry.Hash)
		// fmt.Printf("Hour UTC: %s\n", "N/A") // TODO: Maybe add timestamp
		fmt.Printf("Country Code IP Source: %s\n", entry.CountryCode)
		fmt.Printf("Filename screenshot: %s\n", entry.Filename)
		fmt.Printf("Nameserver requested: %s\n\n", entry.Nameserver.IPs)
	}

	return &metadata
}
