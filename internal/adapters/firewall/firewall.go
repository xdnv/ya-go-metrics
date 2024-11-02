package firewall

import (
	"net"
)

// main cryptor object to store security configuration
type FirewallObject struct {
	useFirewall bool      // enables or disables firewall
	Subnet      net.IPNet // trusted IP and subnet mask
	SubnetStr   string    // secret key to en/decode payload
}

var firewall *FirewallObject

func init() {
	firewall = new(FirewallObject)
}

func IsFirewallEnabled() bool {
	return firewall.useFirewall
}

// set trusted network mask in CIDR format
func SetSubnetMask(subnet string) error {
	firewall.SubnetStr = subnet
	firewall.useFirewall = (firewall.SubnetStr != "")
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return err
	}
	firewall.Subnet = *ipNet
	return nil
}

func IsTrustedIP(ip net.IP) bool {
	if !firewall.useFirewall {
		return true
	}
	return firewall.Subnet.Contains(ip)
}

// get HTTP request header used to subnet filtering
func GetFirewallToken() string {
	return "X-Real-IP"
}
