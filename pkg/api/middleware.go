package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	c "github.com/gp42/pf-demo/pkg/context"
	"github.com/gp42/pf-demo/pkg/db"

	"github.com/go-logr/logr"
)

// Middleware object contains configuration for the RequestMiddleware
type Middleware struct {
	DB  *db.DBConnection
	Log *logr.Logger
}

// RequestMiddleware assigns a unique ID to each request and verifies that IP address
// is not blacklisted.
func (a *Middleware) RequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := c.WithRandomRequestID(r.Context())
		log := a.Log.WithValues("request_id", (ctx).Value(c.REQ_KEY_ID))

		ip, err := RequestToIP(r)
		if err != nil {
			log.Info("Failed to get client IP. Blocking request.")
			// We might want to have some other logic here
			http.Error(w, "unable to identify client ip", http.StatusForbidden)
		}

		// Check if IP is blocked
		if blacklisted, err := a.DB.IsIPBlacklisted(&ctx, ip); blacklisted {
			log.V(1).Info("Blocking request from blacklisted ip", "ip", ip)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		} else if err != nil {
			log.Error(err, "Failed to handle HTTP request while checking if IP is blacklisted")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ctx = context.WithValue(logr.NewContext(ctx, log), c.CLIENT_IP_KEY_ID, ip)

		log.V(1).Info("Processing HTTP request", "method", r.Method, "URL", r.URL.String(), "headers", r.Header)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestToIP finds client IP address from an HTTP request.
// HTTP Request RemoteAddr provides us with the address of the client or last
// proxy, therefore we apply some logic to provide more reliable results
//
// 1. First we try to get IP from the 'X-Forwarded-For' header
// 2. If the header is set, then we get the first IP in this list:
//    https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For
// 3. If the header is not set - fall back to RemoteAddr
//
// Potential improvements depending on requirements are:
// - Add configuration to ignore provided ip ranges
// - Add configuration to ignore 'X-Forwarded-For' header
// 	 (probably we dont want to trust 'X-Forwarded-For' from public internet)
func RequestToIP(r *http.Request) (string, error) {
	var addr string
	addrs := r.Header.Get("X-Forwarded-For")
	if addrs != "" {
		addrSlice := strings.Split(addrs, ",")
		// Addr could be a comma-separated array, first address is the client
		//    https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For
		addr = addrSlice[0]
	} else {
		addr = r.RemoteAddr
	}

	// Try to remove port from ip if exists
	var ip string
	ip, _, err := net.SplitHostPort(addr)
	if _, ok := err.(*net.AddrError); ok {
		// Skip AddrError, addr may have no port defined
		ip = addr
	} else if err != nil {
		// Some other error
		return "", fmt.Errorf("failed to split port from IP: %s", err.Error())
	}

	// Try to parse IP address to verify
	ipBytes := net.ParseIP(ip)
	if ipBytes == nil {
		return "", fmt.Errorf(`failed to parse IP address: "%s"`, ip)
	}

	return ipBytes.String(), nil
}
