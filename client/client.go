package client

import (
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-ping/ping"
)

var (
	secureHTTPClient   *http.Client
	insecureHTTPClient *http.Client

	// pingTimeout is the timeout for the Ping function
	// This is mainly exposed for testing purposes
	pingTimeout = 5 * time.Second

	// httpTimeout is the timeout for secureHTTPClient and insecureHTTPClient
	httpTimeout = 10 * time.Second
)

func init() {
	// XXX: This is an undocumented feature. See https://github.com/Meldiron/gatus/issues/104.
	httpTimeoutInSecondsFromEnvironmentVariable := os.Getenv("HTTP_CLIENT_TIMEOUT_IN_SECONDS")
	if len(httpTimeoutInSecondsFromEnvironmentVariable) > 0 {
		if httpTimeoutInSeconds, err := strconv.Atoi(httpTimeoutInSecondsFromEnvironmentVariable); err == nil {
			httpTimeout = time.Duration(httpTimeoutInSeconds) * time.Second
		}
	}
}

// GetHTTPClient returns the shared HTTP client
func GetHTTPClient(insecure bool) *http.Client {
	if insecure {
		if insecureHTTPClient == nil {
			insecureHTTPClient = &http.Client{
				Timeout: httpTimeout,
				Transport: &http.Transport{
					MaxIdleConns:        100,
					MaxIdleConnsPerHost: 20,
					Proxy:               http.ProxyFromEnvironment,
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}
		}
		return insecureHTTPClient
	}
	if secureHTTPClient == nil {
		secureHTTPClient = &http.Client{
			Timeout: httpTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				Proxy:               http.ProxyFromEnvironment,
			},
		}
	}
	return secureHTTPClient
}

// CanCreateTCPConnection checks whether a connection can be established with a TCP service
func CanCreateTCPConnection(address string) bool {
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// Ping checks if an address can be pinged and returns the round-trip time if the address can be pinged
//
// Note that this function takes at least 100ms, even if the address is 127.0.0.1
func Ping(address string) (bool, time.Duration) {
	pinger, err := ping.NewPinger(address)
	if err != nil {
		return false, 0
	}
	pinger.Count = 1
	pinger.Timeout = pingTimeout
	pinger.SetNetwork("ip4")
	pinger.SetPrivileged(true)
	err = pinger.Run()
	if err != nil {
		return false, 0
	}
	if pinger.Statistics() != nil {
		// If the packet loss is 100, it means that the packet didn't reach the host
		if pinger.Statistics().PacketLoss == 100 {
			return false, pinger.Timeout
		}
		return true, pinger.Statistics().MaxRtt
	}
	return true, 0
}
