package netx

import (
	"net"
	"strings"
)

// NetworkInterface 描述本机 IPv4 网络接口。
type NetworkInterface struct {
	Name    string `json:"name"`
	IP      string `json:"ip"`
	Netmask string `json:"netmask"`
	Subnet  string `json:"subnet"`
}

// LocalInterfaceOptions 控制本机网络接口过滤规则。
type LocalInterfaceOptions struct {
	IncludeLoopback bool
	IncludeInactive bool
	IncludeVirtual  bool
	IncludeRFC2544  bool
}

type localInterfaceCandidate struct {
	name  string
	flags net.Flags
	ip    net.IP
	mask  net.IPMask
}

// LocalIPv4Interfaces 返回过滤后的本机 IPv4 网络接口。
func LocalIPv4Interfaces(options LocalInterfaceOptions) ([]NetworkInterface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	candidates := make([]localInterfaceCandidate, 0, len(ifaces))
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			candidates = append(candidates, localInterfaceCandidate{
				name:  iface.Name,
				flags: iface.Flags,
				ip:    ipnet.IP,
				mask:  ipnet.Mask,
			})
		}
	}

	return filterLocalInterfaceCandidates(candidates, options), nil
}

// LocalIPsFromInterfaces 从接口列表中返回展示用 IP，并在为空时回退到 localhost。
func LocalIPsFromInterfaces(interfaces []NetworkInterface, bindAddr string, showAll bool) []string {
	ips := make([]string, 0, len(interfaces))
	for _, iface := range interfaces {
		if bindAddr != "" && !showAll && iface.IP != bindAddr {
			continue
		}
		ips = append(ips, iface.IP)
	}
	if len(ips) == 0 {
		return []string{"localhost"}
	}
	return ips
}

// InterfacesBySubnet 按子网分组接口，并可按 bindAddr 做展示过滤。
func InterfacesBySubnet(interfaces []NetworkInterface, bindAddr string, showAll bool) map[string][]NetworkInterface {
	subnets := make(map[string][]NetworkInterface)
	for _, iface := range interfaces {
		if bindAddr != "" && !showAll && iface.IP != bindAddr {
			continue
		}
		subnets[iface.Subnet] = append(subnets[iface.Subnet], iface)
	}
	return subnets
}

// IsVirtualInterfaceName 判断接口名是否匹配常见虚拟接口命名。
func IsVirtualInterfaceName(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	virtualPrefixes := []string{
		"utun",
		"docker",
		"veth",
		"br-",
		"vmnet",
		"vboxnet",
		"tun",
		"tap",
		"tailscale",
		"wg",
		"zt",
		"meta",
		"clash",
		"anyconnect",
	}
	for _, prefix := range virtualPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}

	virtualKeywords := []string{
		"vmware",
		"virtualbox",
		"hyper-v",
		"vethernet",
		"loopback",
	}
	for _, keyword := range virtualKeywords {
		if strings.Contains(name, keyword) {
			return true
		}
	}
	return false
}

func filterLocalInterfaceCandidates(candidates []localInterfaceCandidate, options LocalInterfaceOptions) []NetworkInterface {
	result := make([]NetworkInterface, 0, len(candidates))
	for _, candidate := range candidates {
		if !options.IncludeInactive && candidate.flags&net.FlagUp == 0 {
			continue
		}
		if !options.IncludeLoopback && (candidate.flags&net.FlagLoopback != 0 || candidate.ip.IsLoopback()) {
			continue
		}
		if !options.IncludeVirtual && IsVirtualInterfaceName(candidate.name) {
			continue
		}

		ip4 := candidate.ip.To4()
		if ip4 == nil {
			continue
		}
		if !options.IncludeRFC2544 && isRFC2544BenchmarkIP(ip4) {
			continue
		}

		result = append(result, NetworkInterface{
			Name:    candidate.name,
			IP:      ip4.String(),
			Netmask: net.IP(candidate.mask).String(),
			Subnet:  subnetCIDR(ip4, candidate.mask),
		})
	}
	return result
}

func isRFC2544BenchmarkIP(ip net.IP) bool {
	ip4 := ip.To4()
	return ip4 != nil && ip4[0] == 198 && (ip4[1] == 18 || ip4[1] == 19)
}

func subnetCIDR(ip net.IP, mask net.IPMask) string {
	ones, bits := mask.Size()
	if bits == 0 {
		return ip.Mask(mask).String()
	}
	return ip.Mask(mask).String() + "/" + itoa(ones)
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := [20]byte{}
	index := len(digits)
	for value > 0 {
		index--
		digits[index] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[index:])
}
