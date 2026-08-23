package ipgeo

import (
	"encoding/binary"
	"net"
	"strings"
	"sync"
)

type Location struct {
	Country string `json:"country"`
	Region  string `json:"region"`
	City    string `json:"city"`
	ISP     string `json:"isp"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Timezone string `json:"timezone"`
}

type Resolver struct {
	mu      sync.RWMutex
	cache   map[string]*Location
	enabled bool
}

func NewResolver(enabled bool) *Resolver {
	return &Resolver{
		cache:   make(map[string]*Location),
		enabled: enabled,
	}
}

var privateIPBlocks []*net.IPNet

func init() {
	for _, cidr := range []string{
		"127.0.0.0/8", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16",
		"169.254.0.0/16", "::1/128", "fc00::/7", "fe80::/10",
	} {
		_, block, _ := net.ParseCIDR(cidr)
		privateIPBlocks = append(privateIPBlocks, block)
	}
}

func IsPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func (r *Resolver) Resolve(ipStr string) *Location {
	if !r.enabled {
		return &Location{Country: "Unknown", Region: "", City: ""}
	}
	if ipStr == "" || IsPrivateIP(ipStr) || ipStr == "::1" {
		return &Location{Country: "Local", Region: "Local", City: "Localhost"}
	}

	r.mu.RLock()
	if cached, ok := r.cache[ipStr]; ok {
		r.mu.RUnlock()
		return cached
	}
	r.mu.RUnlock()

	loc := r.lookupIP(ipStr)

	r.mu.Lock()
	if len(r.cache) > 10000 {
		for k := range r.cache {
			delete(r.cache, k)
			break
		}
	}
	r.cache[ipStr] = loc
	r.mu.Unlock()

	return loc
}

func (r *Resolver) lookupIP(ipStr string) *Location {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return &Location{Country: "Unknown"}
	}

	country := r.ipToCountry(ip)
	return &Location{
		Country: country,
		Region:  "",
		City:    "",
	}
}

func (r *Resolver) ipToCountry(ip net.IP) string {
	ipv4 := ip.To4()
	if ipv4 == nil {
		return "Unknown"
	}

	ipInt := binary.BigEndian.Uint32(ipv4)

	switch {
	case ipInt >= 0x01000000 && ipInt <= 0x7E0000FF:
		if isCN(ipInt) {
			return "China"
		}
		if isUS(ipInt) {
			return "United States"
		}
		if isJP(ipInt) {
			return "Japan"
		}
		return "Asia"
	case ipInt >= 0x7F000000 && ipInt <= 0xBE0000FF:
		return "Europe"
	case ipInt >= 0xBF000000 && ipInt <= 0xC0FF00FF:
		return "North America"
	default:
		return "Unknown"
	}
}

func isCN(ip uint32) bool {
	ranges := [][2]uint32{
		{0x3E940000, 0x3E94FFFF},
		{0x58AD0000, 0x58B5FFFF},
		{0x60040000, 0x6005FFFF},
	}
	for _, r := range ranges {
		if ip >= r[0] && ip <= r[1] {
			return true
		}
	}
	return false
}

func isUS(ip uint32) bool {
	ranges := [][2]uint32{
		{0x03000000, 0x03FFFFFF},
		{0x04000000, 0x04FFFFFF},
		{0x05000000, 0x05FFFFFF},
	}
	for _, r := range ranges {
		if ip >= r[0] && ip <= r[1] {
			return true
		}
	}
	return false
}

func isJP(ip uint32) bool {
	ranges := [][2]uint32{
		{0x0D000000, 0x0DFFFFFF},
		{0x2EFF0000, 0x2EFFFFFF},
	}
	for _, r := range ranges {
		if ip >= r[0] && ip <= r[1] {
			return true
		}
	}
	return false
}

func ParseUserAgent(ua string) (device, os, browser string) {
	uaLower := strings.ToLower(ua)

	switch {
	case strings.Contains(uaLower, "iphone"):
		device = "iPhone"
	case strings.Contains(uaLower, "ipad"):
		device = "iPad"
	case strings.Contains(uaLower, "android"):
		device = "Android"
	case strings.Contains(uaLower, "macintosh"), strings.Contains(uaLower, "mac os"):
		device = "Mac"
	case strings.Contains(uaLower, "windows"):
		device = "Windows"
	case strings.Contains(uaLower, "linux"):
		device = "Linux"
	default:
		device = "Other"
	}

	switch {
	case strings.Contains(uaLower, "iphone os"), strings.Contains(uaLower, "ipad; cpu os"):
		os = "iOS"
	case strings.Contains(uaLower, "android"):
		os = "Android"
	case strings.Contains(uaLower, "mac os x"):
		os = "macOS"
	case strings.Contains(uaLower, "windows nt 10") || strings.Contains(uaLower, "windows nt 6.3"):
		os = "Windows 10/11"
	case strings.Contains(uaLower, "windows nt 6.2"):
		os = "Windows 8"
	case strings.Contains(uaLower, "windows nt 6.1"):
		os = "Windows 7"
	case strings.Contains(uaLower, "linux"):
		os = "Linux"
	default:
		os = "Unknown"
	}

	switch {
	case strings.Contains(uaLower, "safari") && strings.Contains(uaLower, "chrome") == false:
		browser = "Safari"
	case strings.Contains(uaLower, "edg/"):
		browser = "Edge"
	case strings.Contains(uaLower, "chrome"), strings.Contains(uaLower, "chromium"):
		browser = "Chrome"
	case strings.Contains(uaLower, "firefox"), strings.Contains(uaLower, "fxios"):
		browser = "Firefox"
	case strings.Contains(uaLower, "opera"), strings.Contains(uaLower, "opr/"):
		browser = "Opera"
	case strings.Contains(uaLower, "podcast"), strings.Contains(uaLower, "castbox"),
		strings.Contains(uaLower, "xiaoyuzhou"):
		browser = "Podcast App"
	default:
		browser = "Other"
	}

	return device, os, browser
}
