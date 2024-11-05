// the firewall middleware enables communication with trusted subnets only
package firewall

import (
	"net"
	"net/http"
	"strings"

	"internal/adapters/logger"
)

// provides message security check by kicking agents from untrusted networks
func HandleTrustedNetworkRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

		if !IsFirewallEnabled() {
			next.ServeHTTP(rw, r)
			return
		}

		// parse X-Real-IP header
		ipStr := r.Header.Get(GetFirewallToken())
		ip := net.ParseIP(ipStr)
		// attempt to parse X-Forwarded-For IP chain
		if ip == nil {
			ips := r.Header.Get("X-Forwarded-For")
			ipStrs := strings.Split(ips, ",")
			ipStr = ipStrs[0]
			ip = net.ParseIP(ipStr)
		}
		if ip == nil {
			logger.Error("firewall: failed to parse ip from http header")
			http.Error(rw, "failed to parse ip from http header", http.StatusBadRequest)
			return
		}

		if !IsTrustedIP(ip) {
			logger.Errorf("firewall: agent belongs to untrusted network, ip=%s", ipStr)
			http.Error(rw, "access denied", http.StatusForbidden)
			return
		}

		logger.Infof("firewall: agent is within trusted network, ip=%s", ipStr)

		next.ServeHTTP(rw, r)
	})
}
