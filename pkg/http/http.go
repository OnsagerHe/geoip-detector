package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/chromedp/chromedp"
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	log.Println("browser not found", path)
	return !os.IsNotExist(err)
}

func setBrowserBinaryPath() string {
	const bravePath = "/usr/bin/brave"
	const chromePath = "/usr/bin/chrome"
	// TODO: add another browser
	// firefox does not work with chromedp https://github.com/chromedp/chromedp/issues/837

	if utils.BrowserPath != nil && *utils.BrowserPath != "" {
		if fileExists(*utils.BrowserPath) {
			return *utils.BrowserPath
		}
	}

	switch {
	case fileExists(bravePath):
		return bravePath
	case fileExists(chromePath):
		return chromePath
	default:
		return ""
	}
}

// TakeScreenshot captures a screenshot of the given URL and saves it to the specified folder.
func takeScreenshot(resource *utils.EndpointMetadata, analyze *utils.Analyze) error {
	*utils.BrowserPath = setBrowserBinaryPath()
	if *utils.BrowserPath == "" {
		return fmt.Errorf("browser path unknown")
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(*utils.BrowserPath),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("new-instance", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var buf []byte
	if err := chromedp.Run(ctx, fullScreenshot(resource.Endpoint, &buf)); err != nil {
		return fmt.Errorf("failed to capture screenshot: %w", err)
	}

	if err := os.MkdirAll(*utils.FolderPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}

	// TODO: maybe add encoded information to avoid filename to long
	fileName := fmt.Sprintf("%s_%s_%x.%s", resource.Host, analyze.CountryCode, analyze.Hash, "png")
	analyze.Filename = fileName
	filePath := filepath.Join(*utils.FolderPath, fileName)
	if err := os.WriteFile(filePath, buf, 0644); err != nil {
		return fmt.Errorf("failed to save screenshot: %w", err)
	}

	return nil
}

// fullScreenshot is a helper function to capture a full-page screenshot.
func fullScreenshot(urlstr string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.FullScreenshot(res, 100),
	}
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
