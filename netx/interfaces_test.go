package netx

import (
	"net"
	"reflect"
	"testing"
)

func TestIsVirtualInterfaceName(t *testing.T) {
	virtualNames := []string{
		"utun0",
		"docker0",
		"veth123",
		"br-abc",
		"vmnet8",
		"vboxnet0",
		"tun0",
		"tap0",
		"tailscale0",
		"wg0",
		"zt0",
		"meta0",
		"clash0",
		"anyconnect0",
		"VMware Network Adapter VMnet1",
		"VMware Network Adapter VMnet8",
		"VirtualBox Host-Only Ethernet Adapter",
		"vEthernet (Default Switch)",
		"Hyper-V Virtual Ethernet Adapter",
	}
	for _, name := range virtualNames {
		if !IsVirtualInterfaceName(name) {
			t.Fatalf("expected %s to be virtual", name)
		}
	}

	if IsVirtualInterfaceName("en0") {
		t.Fatal("expected en0 to be physical")
	}
}

func TestVirtualCandidatesKeepIndependentFilters(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		flags    net.Flags
		ordinary bool
		included bool
	}{
		{"Wi-Fi", "192.168.1.20", net.FlagUp, true, true},
		{"vEthernet（无线虚拟交换机）", "192.168.43.117", net.FlagUp, false, true},
		{"vEthernet (Default Switch)", "172.20.0.1", net.FlagUp, false, true},
		{"  VETHERNET (External)  ", "192.168.1.21", net.FlagUp, false, true},
		{"Tailscale", "100.64.0.1", net.FlagUp, false, true},
		{"docker0", "172.17.0.1", net.FlagUp, false, true},
		{"lo", "127.0.0.1", net.FlagUp | net.FlagLoopback, false, false},
		{"vEthernet (Inactive)", "192.168.1.22", 0, false, false},
		{"vEthernet (Fake IP)", "198.18.0.1", net.FlagUp, false, false},
		{"Wi-Fi", "198.19.255.254", net.FlagUp, false, false},
		{"vEthernet (IPv6)", "2001:db8::1", net.FlagUp, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name+tt.ip, func(t *testing.T) {
			candidate := localInterfaceCandidate{name: tt.name, ip: net.ParseIP(tt.ip), flags: tt.flags, mask: net.CIDRMask(24, 32)}
			for _, include := range []bool{false, true} {
				want := tt.ordinary
				if include {
					want = tt.included
				}
				got := filterLocalInterfaceCandidates([]localInterfaceCandidate{candidate}, LocalInterfaceOptions{IncludeVirtual: include})
				if (len(got) == 1) != want {
					t.Fatalf("includeVirtual=%v: got %+v, want retained=%v", include, got, want)
				}
				if want && (got[0].Name != tt.name || got[0].IP != tt.ip) {
					t.Fatalf("candidate changed: %+v", got[0])
				}
			}
		})
	}
}

func TestFilterLocalInterfaceCandidates(t *testing.T) {
	candidates := []localInterfaceCandidate{
		{
			name:  "en0",
			flags: net.FlagUp,
			ip:    net.IPv4(192, 168, 1, 10),
			mask:  net.IPv4Mask(255, 255, 255, 0),
		},
		{
			name:  "docker0",
			flags: net.FlagUp,
			ip:    net.IPv4(172, 17, 0, 1),
			mask:  net.IPv4Mask(255, 255, 0, 0),
		},
		{
			name:  "lo0",
			flags: net.FlagUp | net.FlagLoopback,
			ip:    net.IPv4(127, 0, 0, 1),
			mask:  net.IPv4Mask(255, 0, 0, 0),
		},
		{
			name:  "en1",
			flags: net.Flags(0),
			ip:    net.IPv4(10, 0, 0, 2),
			mask:  net.IPv4Mask(255, 0, 0, 0),
		},
		{
			name:  "en2",
			flags: net.FlagUp,
			ip:    net.IPv4(198, 18, 1, 1),
			mask:  net.IPv4Mask(255, 254, 0, 0),
		},
	}

	got := filterLocalInterfaceCandidates(candidates, LocalInterfaceOptions{})
	expected := []NetworkInterface{
		{
			Name:    "en0",
			IP:      "192.168.1.10",
			Netmask: "255.255.255.0",
			Subnet:  "192.168.1.0/24",
		},
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %+v, got %+v", expected, got)
	}
}

func TestFilterLocalInterfaceCandidatesIncludeOptions(t *testing.T) {
	candidates := []localInterfaceCandidate{
		{name: "docker0", flags: net.FlagUp, ip: net.IPv4(172, 17, 0, 1), mask: net.IPv4Mask(255, 255, 0, 0)},
		{name: "lo0", flags: net.FlagUp | net.FlagLoopback, ip: net.IPv4(127, 0, 0, 1), mask: net.IPv4Mask(255, 0, 0, 0)},
		{name: "en1", flags: net.Flags(0), ip: net.IPv4(10, 0, 0, 2), mask: net.IPv4Mask(255, 0, 0, 0)},
		{name: "en2", flags: net.FlagUp, ip: net.IPv4(198, 19, 1, 1), mask: net.IPv4Mask(255, 254, 0, 0)},
		{name: "en3", flags: net.FlagUp, ip: net.ParseIP("2001:db8::1"), mask: net.CIDRMask(64, 128)},
	}

	got := filterLocalInterfaceCandidates(candidates, LocalInterfaceOptions{
		IncludeVirtual:  true,
		IncludeLoopback: true,
		IncludeInactive: true,
		IncludeRFC2544:  true,
	})

	if len(got) != 4 {
		t.Fatalf("expected 4 IPv4 interfaces, got %+v", got)
	}
}

func TestLocalIPsFromInterfaces(t *testing.T) {
	interfaces := []NetworkInterface{
		{Name: "en0", IP: "192.168.1.10", Subnet: "192.168.1.0/24"},
		{Name: "en1", IP: "10.0.0.2", Subnet: "10.0.0.0/8"},
	}

	if got := LocalIPsFromInterfaces(interfaces, "", false); !reflect.DeepEqual(got, []string{"192.168.1.10", "10.0.0.2"}) {
		t.Fatalf("unexpected IPs: %v", got)
	}
	if got := LocalIPsFromInterfaces(interfaces, "10.0.0.2", false); !reflect.DeepEqual(got, []string{"10.0.0.2"}) {
		t.Fatalf("unexpected bound IPs: %v", got)
	}
	if got := LocalIPsFromInterfaces(interfaces, "10.0.0.2", true); !reflect.DeepEqual(got, []string{"192.168.1.10", "10.0.0.2"}) {
		t.Fatalf("unexpected show all IPs: %v", got)
	}
	if got := LocalIPsFromInterfaces(nil, "", false); !reflect.DeepEqual(got, []string{"localhost"}) {
		t.Fatalf("expected localhost fallback, got %v", got)
	}
}

func TestInterfacesBySubnet(t *testing.T) {
	interfaces := []NetworkInterface{
		{Name: "en0", IP: "192.168.1.10", Subnet: "192.168.1.0/24"},
		{Name: "en1", IP: "192.168.1.11", Subnet: "192.168.1.0/24"},
		{Name: "en2", IP: "10.0.0.2", Subnet: "10.0.0.0/8"},
	}

	got := InterfacesBySubnet(interfaces, "10.0.0.2", false)
	if len(got) != 1 || len(got["10.0.0.0/8"]) != 1 {
		t.Fatalf("unexpected subnet map: %+v", got)
	}
}

func TestLocalIPv4InterfacesDoesNotError(t *testing.T) {
	if _, err := LocalIPv4Interfaces(LocalInterfaceOptions{}); err != nil {
		t.Fatalf("LocalIPv4Interfaces returned error: %v", err)
	}
}
