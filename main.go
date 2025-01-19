package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/OnsagerHe/geoip-detector/internal/api"

	"github.com/OnsagerHe/geoip-detector/pkg/retriever"
	"github.com/OnsagerHe/geoip-detector/pkg/utils"
	"github.com/OnsagerHe/geoip-detector/pkg/utils/logger"
	"github.com/OnsagerHe/geoip-detector/pkg/vpn"
)

var endpoint *string
var loop *uint
var server *bool

func init() {
	endpoint = flag.String("endpoint", "http://onsager.net", "endpoint to test")
	loop = flag.Uint("loop", 3, "number of localizations you wish to use")
	utils.Source = flag.Bool("source", false, "download source code endpoint")
	utils.Screenshot = flag.Bool("screenshot", true, "get screenshot endpoint")
	utils.FolderPath = flag.String("folder", "downloads", "path folder for images webpage")
	utils.BrowserPath = flag.String("browser", "", "path to binary browser")
	utils.Prd = flag.Bool("prod", true, "don't print debug log") // set var env\
	server = flag.Bool("server", true, "run api server")
	flag.Parse()
}

func initGeoIP() *utils.GeoIP {
	res := &utils.GeoIP{
		Resource:    utils.EndpointMetadata{Endpoint: *endpoint},
		Analyzes:    nil,
		VPNProvider: vpn.Mullvad{},
		Logger:      logger.CreateLogger(*utils.Prd),
	}

	return res
}

func run() {
	res := initGeoIP()
	vpnProvider := vpn.Mullvad{}
	if err := connectToVPN(&vpnProvider); err != nil {
		log.Printf("VPN connection error: %v\n", err)
		return
	}

	rtr := retriever.Init(res, uint8(*loop))

	if *server {
		frontend := api.InitServer(rtr)
		if err := api.LaunchServer(frontend); err != nil {
			log.Fatal(err)
		}
		return
	}

	rtr.Process.Logger.Debug("value for endpoint and loop:" + *endpoint)
	err := rtr.CheckEndpoint()
	if err != nil {
		log.Fatal(err)
	}

}

func connectToVPN(vpnProvider *vpn.Mullvad) error {
	if err := vpnProvider.ConnectVPN(); err != nil {
		return errors.New("cannot connect to Mullvad VPN")
	}
	time.Sleep(3 * time.Second)
	fmt.Print("VPN connection")
	for i := 0; i < 3; i++ {
		fmt.Print(".")
		time.Sleep(1 * time.Second)
	}
	fmt.Println("")
	return nil
}

func main() {
	run()
}
