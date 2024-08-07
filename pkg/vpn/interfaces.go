package vpn

type IProviderVPN interface {
	ConnectVPN() error
	ListVPN() map[string][]string
	SetLocationVPN(countryCode string) ([]string, error)
	SetCustomDNSResolver(ip string) error
	SetDefaultDNSResolver() error
	CheckVPNStatus(expectedCountryCode string) ([]string, error)
	ExtractIPAddresses(status string) (ipv4, ipv6 string)
}
