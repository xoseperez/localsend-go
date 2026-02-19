package discovery

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/meowrain/localsend-go/internal/utils/logger"

	probing "github.com/prometheus-community/pro-bing"
)

// GetLocalIP returns all local IPv4 addresses
func GetLocalIP() ([]net.IP, error) {
	ips := make([]net.IP, 0)
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if v.IP.To4() != nil && !v.IP.IsLoopback() {
					ips = append(ips, v.IP)
				}
			}
		}
	}
	return ips, nil
}

// pingScan uses ICMP ping to scan all active devices on the local network
func pingScan() ([]string, error) {
	var ips []string
	ipGroup, err := GetLocalIP()
	if err != nil {
		return nil, err
	}
	for _, i := range ipGroup {
		ip := i.Mask(net.IPv4Mask(255, 255, 255, 0)) // Assumes /24 subnet mask
		ip4 := ip.To4()
		if ip4 == nil {
			return nil, fmt.Errorf("invalid IPv4 address")
		}

		var wg sync.WaitGroup
		var mu sync.Mutex

		for i := 1; i < 255; i++ {
			ip4[3] = byte(i)
			targetIP := ip4.String()

			wg.Add(1)
			go func(ip string) {
				defer wg.Done()
				pinger, err := probing.NewPinger(ip)
				if err != nil {
					logger.Errorf("Failed to create pinger:", err)
					return
				}
				pinger.SetPrivileged(true)
				pinger.Count = 1
				pinger.Timeout = time.Second * 1

				pinger.OnRecv = func(pkt *probing.Packet) {
					mu.Lock()
					ips = append(ips, ip)
					mu.Unlock()
				}
				err = pinger.Run()
				if err != nil {
					// Ignore ping send failures
					return
				}
			}(targetIP)
		}

		wg.Wait()
	}
	return ips, nil
}
