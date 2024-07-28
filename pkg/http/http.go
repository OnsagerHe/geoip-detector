package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/chromedp/chromedp"
	"golang.org/x/exp/rand"
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

// fullScreenshot is a helper function to capture a full-page screenshot.
func fullScreenshot(urlstr string, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.FullScreenshot(res, quality),
	}
}

func TakeScreenshot(res *utils.GeoIP) {
	for i := range res.Analyzes {
		err := takeScreenshot(&res.Resource, &(res.Analyzes)[i])
		if err != nil {
			log.Println("error", err)
			return
		}
	}
}

func TakeScreenshotByCountryCode(res *utils.GeoIP, analyzes []*utils.Analyze) {
	for i := range analyzes {
		err := takeScreenshot(&res.Resource, (analyzes)[i])
		if err != nil {
			log.Println("error", err)
			return
		}
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
)

func RandStringBytesMask(n int) string {
	b := make([]byte, n)
	for i := 0; i < n; {
		if idx := int(rand.Int63() & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i++
		}
	}
	return string(b)
}

// TakeScreenshot captures a screenshot of the given URL and saves it to the specified folder.
func takeScreenshot(resource *utils.EndpointMetadata, analyze *utils.Analyze) error {
	bravePath := "/usr/bin/brave"

	log.Println("value", analyze.Nameserver.IPs[0].String())
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(bravePath),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("new-instance", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Run task list
	var buf []byte
	if err := chromedp.Run(ctx, screenshotTasks(resource.Endpoint, &buf)); err != nil {
		return fmt.Errorf("failed to capture screenshot: %w", err)
	}

	// Ensure the folder exists
	if err := os.MkdirAll("downloads", os.ModePerm); err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}

	//s := RandStringBytesMask(3)
	fileName := fmt.Sprintf("%s_%s_%x.%s", resource.Host, analyze.CountryCode, analyze.Hash, "png")
	analyze.Filename = fileName
	filePath := filepath.Join("downloads", fileName)
	if err := os.WriteFile(filePath, buf, 0644); err != nil {
		return fmt.Errorf("failed to save screenshot: %w", err)
	}

	return nil
}

// screenshotTasks returns a chromedp.Tasks to capture a screenshot of a webpage.
func screenshotTasks(url string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Screenshot(`body`, res, chromedp.NodeVisible, chromedp.ByQuery),
	}
}

func RequestEndpoints(res *utils.GeoIP) {
	for i := range res.Analyzes {
		RequestEndpoint(&res.Resource, &(res.Analyzes)[i])
	}
}

func RequestSpecificEndpoints(res *utils.GeoIP, analyzes []*utils.Analyze) {
	for i := range analyzes {
		RequestEndpoint(&res.Resource, &*(analyzes)[i])
	}
}

func RequestEndpoint(resource *utils.EndpointMetadata, analyze *utils.Analyze) {
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
		log.Println("Value append", analyze.Online, analyze.Hash)
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
