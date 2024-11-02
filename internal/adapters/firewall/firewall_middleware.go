// the firewall middleware enables communication with trusted subnets only
package firewall

import (
	"fmt"
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
			logger.Error(fmt.Sprint("firewall: agent belongs to untrusted network, ip=", ipStr))
			http.Error(rw, "access denied", http.StatusForbidden)
			return
		}

		logger.Info(fmt.Sprint("firewall: agent is within trusted network, ip=", ipStr))

		next.ServeHTTP(rw, r)
	})
}
