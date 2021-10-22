// Package handler implements handlers for HTTP routes
package handler

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/go-logr/logr"

	c "github.com/gp42/pf-demo/pkg/context"
	"github.com/gp42/pf-demo/pkg/db"
	m "github.com/gp42/pf-demo/pkg/mailer"
)

var (
	maxN = int(math.Sqrt(math.MaxInt32))
)

// Handler object contains handler configuration
type Handler struct {
	DB     *db.DBConnection
	Mailer *m.Mailer
}

// Handler for Healthcheck endpoint
// This check should have some logic to check if the application is healthy
func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	log := logr.FromContextOrDiscard(r.Context()).WithValues("handler", "status")

	w.Header().Add("Content-Type", "text/plain")
	_, err := w.Write([]byte("OK"))
	if err != nil {
		log.Error(err, "error writing response body")
	}
}

// Handler for calculating square of n parameter
// It relies on the 'n' query parameter and returns n*n
func (h *Handler) Square(w http.ResponseWriter, r *http.Request) {
	log := logr.FromContextOrDiscard(r.Context()).WithValues("handler", "square")
	log.V(1).Info("Running handler")

	params := r.URL.Query()
	n, err := getSquareNumberFromString(params.Get("n"))
	if err != nil {
		log.Error(err, "failed to get square number")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "text/plain")
	_, err = w.Write([]byte(*n))
	if err != nil {
		log.Error(err, "error writing response body")
	}
}

// Calculate square root from a string value and return square with necessary sanity checks
func getSquareNumberFromString(nStr string) (*string, error) {
	n, err := strconv.Atoi(nStr)
	if err != nil {
		return nil, err
	}

	if n > maxN || n < -maxN {
		return nil, fmt.Errorf("Provided n is too big. Must be in range: %d >= n <= %d, got %d", -maxN, maxN, n)
	}
	r := strconv.Itoa(n * n)
	return &r, nil
}

// Handler for blacklisting a client
// It will skip localhost IPv4 and IPv6 addresses, and perform the following
// operations
// - Add db record about blacklisted IP
// - Send an email about blacklisted IP
func (h *Handler) Blacklisted(w http.ResponseWriter, r *http.Request) {
	log := logr.FromContextOrDiscard(r.Context()).WithValues("handler", "blacklisted")
	log.V(1).Info("Running handler")

	ip := (r.Context()).Value(c.CLIENT_IP_KEY_ID).(string)

	// Do not block localhost
	if ip != "127.0.0.1" && ip != "::1" {
		log.V(1).Info("Blacklisting client", "ip", ip)
		t := time.Now().UTC()
		ctx := r.Context()
		h.DB.UpsertBlacklistRecord(&ctx, ip, t)
		h.Mailer.Send(&ctx, "Blacklist "+ip, "Blacklisted ip: "+ip+" at "+t.String())
	} else {
		log.V(1).Info("Not blocking localhost", "ip", ip)
	}

	w.WriteHeader(444)
}
