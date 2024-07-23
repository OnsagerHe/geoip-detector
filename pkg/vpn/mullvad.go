package vpn

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type Mullvad struct{}

func parseOutput(output string) map[string][]string {
	scanner := bufio.NewScanner(strings.NewReader(output))
	relays := make(map[string][]string)

	var currentCountryCode string

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(line, "\t\t") {
			parts := strings.Fields(trimmedLine)
			if len(parts) > 0 {
				relayName := parts[0]
				relays[currentCountryCode] = append(relays[currentCountryCode], relayName)
			}
		} else if strings.HasPrefix(line, "\t") {
			continue
		} else {
			if idx := strings.Index(trimmedLine, " ("); idx != -1 {
				countryCode := trimmedLine[idx+2 : len(trimmedLine)-1]
				currentCountryCode = countryCode
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading output: %v\n", err)
	}

	return relays
}

func (m Mullvad) ExtractIPAddresses(status string) (ipv4, ipv6 string) {
	ipv4Pattern := `IPv4:\s*([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)`
	ipv6Pattern := `IPv6:\s*([a-fA-F0-9:]+)`

	reIPv4 := regexp.MustCompile(ipv4Pattern)
	reIPv6 := regexp.MustCompile(ipv6Pattern)

	ipv4Match := reIPv4.FindStringSubmatch(status)
	ipv6Match := reIPv6.FindStringSubmatch(status)

	if len(ipv4Match) > 1 {
		ipv4 = ipv4Match[1]
	}

	if len(ipv6Match) > 1 {
		ipv6 = ipv6Match[1]
	}

	return
}

func (m Mullvad) ListVPN() map[string][]string {
	cmd := exec.Command("mullvad", "relay", "list")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Printf("Error executing command: %v\n", err)
		return nil
	}

	relays := parseOutput(out.String())

	// Print the parsed relays
	//for countryCode, names := range relays {
	//	for _, name := range names {
	//		fmt.Printf("Country Code: %s, Name: %s\n", countryCode, name)
	//	}
	//}

	return relays
}

func (m Mullvad) ConnectVPN() error {
	cmd := exec.Command("mullvad", "connect")
	err := cmd.Run()
	if err != nil {
		log.Printf("Error executing command: %v\n", err)
		return err
	}
	return nil
}

func (m Mullvad) SetLocationVPN(countryCode string) ([]string, error) {
	cmd := exec.Command("mullvad", "relay", "set", "location", countryCode)
	err := cmd.Run()
	if err != nil {
		log.Printf("Error executing command: %v\n", err)
		return nil, err
	}

	ips, err := m.CheckVPNStatus(countryCode)
	log.Println("ips", ips)
	if err != nil {
		log.Println("Error:", err)
		return nil, err
	}
	log.Println("Connected to the correct country code.")

	return ips, err
}

func (m Mullvad) SetCustomDNSResolver(ip string) error {
	cmd := exec.Command("mullvad", "dns", "set", "custom", ip)
	err := cmd.Run()
	if err != nil {
		log.Printf("Error executing command: %v\n", err)
		return err
	}

	return nil
}

func (m Mullvad) SetDefaultDNSResolver() error {
	cmd := exec.Command("mullvad", "dns", "set", "default")
	err := cmd.Run()
	if err != nil {
		log.Printf("Error executing command: %v\n", err)
		return err
	}

	return nil
}

func (m Mullvad) CheckVPNStatus(expectedCountryCode string) ([]string, error) {
	timeout := time.After(1 * time.Minute)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timeout reached while waiting for correct country code")
		case <-tick:
			res, ips, err := getMullvadCountryCode(expectedCountryCode)
			if err != nil {
				//log.Printf("Error getting Mullvad status: %v\n", err)
				continue
			}

			if res {
				return ips, nil
			} else {
				log.Printf("Current country code does not match expected (%s), retrying...\n", expectedCountryCode)
			}
		}
	}
}

func getMullvadCountryCode(expectedCountryCode string) (bool, []string, error) {
	cmd := exec.Command("mullvad", "status")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false, nil, fmt.Errorf("error executing command: %v", err)
	}

	output := out.String()
	res, ips := printRelayIdentifierIfContains(output, expectedCountryCode)
	if res != true {
		return res, nil, fmt.Errorf("country code not found: %v", err)
	}

	return res, ips, nil
}

func printRelayIdentifierIfContains(input, countryCode string) (bool, []string) {
	if strings.Contains(input, fmt.Sprintf("Connected to %s-", countryCode)) {
		ipv4Regex := regexp.MustCompile(`IPv4:\s*([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)`)
		ipv6Regex := regexp.MustCompile(`IPv6:\s*([a-fA-F0-9:]+)`)

		ipv4Match := ipv4Regex.FindStringSubmatch(input)
		ipv6Match := ipv6Regex.FindStringSubmatch(input)

		var ipAddresses []string
		if len(ipv4Match) > 1 {
			ipAddresses = append(ipAddresses, ipv4Match[1])
		}
		if len(ipv6Match) > 1 {
			ipAddresses = append(ipAddresses, ipv6Match[1])
		}

		start := strings.Index(input, "Connected to ") + len("Connected to ")
		end := strings.Index(input[start:], " ")
		if end != -1 && ipAddresses != nil {
			log.Println(input[start : start+end])
			return true, ipAddresses
		} else {
			log.Println("Relay identifier not found")
		}
	} else {
		log.Printf("The input string does not contain the country code: %s\n", countryCode)
	}
	return false, nil
}
