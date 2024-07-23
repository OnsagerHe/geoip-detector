package http

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/OnsagerHe/geoip-detector/pkg/utils"
)

func getDomainFromURL(resource *utils.EndpointMetadata) error {
	u, err := url.Parse(resource.Endpoint)
	if err != nil {
		return err
	}
	resource.Host = u.Hostname()
	return nil
}

// customDialer creates a custom dialer that maps the domain to a specific IP address
func customDialer(domain, ip, port string) func(ctx context.Context, network, address string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		if address == domain+port {
			address = ip + port
		}
		dialer := &net.Dialer{
			Timeout: 5 * time.Second,
		}
		return dialer.DialContext(ctx, network, address)
	}
}

func InitHTTPInformation(resource *utils.EndpointMetadata) error {
	err := parseHTTP(resource)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return err
	}

	err = getDomainFromURL(resource)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return err
	}

	return nil
}

func RequestEndpoints(resource *utils.EndpointMetadata, analyzes *[]utils.Analyze) {
	for i := range *analyzes {
		requestEndpoint(resource, &(*analyzes)[i])
	}
}

func requestEndpoint(resource *utils.EndpointMetadata, analyze *utils.Analyze) {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: customDialer(resource.Host, analyze.IpDest, resource.Port),
		},
	}

	req, err := http.NewRequest("GET", resource.Endpoint, nil)
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error performing request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading the response body: %v\n", err)
			return
		}

		analyze.Hash = utils.HashByte(body)
		analyze.Online = true

		hashStr := hex.EncodeToString(analyze.Hash)

		fileName := fmt.Sprintf("%s_%s_%s.html", resource.Host, analyze.CountryCode, hashStr)
		filePath := filepath.Join("downloads/"+resource.Host, fileName)

		analyze.Filename = filePath
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			log.Printf("Error creating directory: %v\n", err)
			return
		}

		file, err := os.Create(filePath)
		if err != nil {
			log.Printf("Error creating file: %v\n", err)
			return
		}
		defer file.Close()

		if _, err := file.Write(body); err != nil {
			log.Printf("Error writing to file: %v\n", err)
			return
		}

		log.Printf("Content saved to: %s\n", filePath)
	} else {
		analyze.Online = false
	}
}

func parseHTTP(resource *utils.EndpointMetadata) error {
	if strings.HasPrefix(resource.Endpoint, "https://") {
		resource.Prefix = "https://"
		resource.Port = ":443"
	} else if strings.HasPrefix(resource.Endpoint, "http://") {
		resource.Prefix = "http://"
		resource.Port = ":80"
	} else {
		return errors.New("endpoint must be contain http:// or https://")
	}
	return nil

}
