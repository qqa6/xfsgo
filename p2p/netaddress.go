// Copyright 2018 The xfsgo Authors
// This file is part of the xfsgo library.
//
// The xfsgo library is free software: you can redistribute it and/or modify
// it under the terms of the MIT Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The xfsgo library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// MIT Lesser General Public License for more details.
//
// You should have received a copy of the MIT Lesser General Public License
// along with the xfsgo library. If not, see <https://mit-license.org/>.

package p2p

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/btcsuite/go-socks/socks"
)

// NetAddress defines information about a peer on the network
// including its IP address, and port.
type NetAddress struct {
	IP    net.IP
	Port  uint16
	str   string
	isLAN bool
}

func NewNetAddress(addr net.Addr) *NetAddress {
	tcpAddr, ok := addr.(*net.TCPAddr)
	if !ok {
		if flag.Lookup("test.v") == nil { // normal run
			fmt.Sprintf("Only TCPAddrs are supported. Got: %v", addr)
		} else { // in testing
			return NewNetAddressIPPort(net.IP("0.0.0.0"), 0)
		}
	}
	ip := tcpAddr.IP
	port := uint16(tcpAddr.Port)
	return NewNetAddressIPPort(ip, port)
}

func NewNetAddressString(addr string) (*NetAddress, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	ip := net.ParseIP(host)
	if ip == nil {
		if len(host) > 0 {
			ips, err := net.LookupIP(host)
			if err != nil {
				return nil, err
			}
			ip = ips[0]
		}
	}

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, err
	}

	na := NewNetAddressIPPort(ip, uint16(port))
	return na, nil
}

func NewNetAddressStrings(addrs []string) ([]*NetAddress, error) {
	netAddrs := make([]*NetAddress, len(addrs))
	for i, addr := range addrs {
		netAddr, err := NewNetAddressString(addr)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Error in address %s: %v", addr, err))
		}
		netAddrs[i] = netAddr
	}
	return netAddrs, nil
}

func NewNetAddressIPPort(ip net.IP, port uint16) *NetAddress {
	return &NetAddress{
		IP:   ip,
		Port: port,
		str: net.JoinHostPort(
			ip.String(),
			strconv.FormatUint(uint64(port), 10),
		),
	}
}

func NewLANNetAddressIPPort(ip net.IP, port uint16) *NetAddress {
	return &NetAddress{
		IP:   ip,
		Port: port,
		str: net.JoinHostPort(
			ip.String(),
			strconv.FormatUint(uint64(port), 10),
		),
		isLAN: true,
	}
}

func (na *NetAddress) Equals(other interface{}) bool {
	if o, ok := other.(*NetAddress); ok {
		return na.String() == o.String()
	}
	return false
}

func (na *NetAddress) String() string {
	if na.str == "" {
		na.str = net.JoinHostPort(
			na.IP.String(),
			strconv.FormatUint(uint64(na.Port), 10),
		)
	}
	return na.str
}

func (na *NetAddress) DialString() string {
	return net.JoinHostPort(
		na.IP.String(),
		strconv.FormatUint(uint64(na.Port), 10),
	)
}

func (na *NetAddress) Dial() (net.Conn, error) {
	conn, err := net.Dial("tcp", na.DialString())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (na *NetAddress) DialTimeout(timeout time.Duration) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", na.DialString(), timeout)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (na *NetAddress) DialTimeoutWithProxy(proxy *socks.Proxy, timeout time.Duration) (net.Conn, error) {
	conn, err := proxy.DialTimeout("tcp", na.DialString(), timeout)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (na *NetAddress) Routable() bool {
	return na.Valid() && !(na.RFC1918() || na.RFC3927() || na.RFC4862() ||
		na.RFC4193() || na.RFC4843() || na.Local())
}

func (na *NetAddress) Valid() bool {
	return na.IP != nil && !(na.IP.IsUnspecified() || na.RFC3849() ||
		na.IP.Equal(net.IPv4bcast))
}

func (na *NetAddress) Local() bool {
	return na.IP.IsLoopback() || zero4.Contains(na.IP)
}

func (na *NetAddress) ReachabilityTo(o *NetAddress) int {
	const (
		Unreachable = 0
		Default     = iota
		Teredo
		Ipv6Weak
		Ipv4
		Ipv6Strong
	)
	if !na.Routable() {
		return Unreachable
	} else if na.RFC4380() {
		if !o.Routable() {
			return Default
		} else if o.RFC4380() {
			return Teredo
		} else if o.IP.To4() != nil {
			return Ipv4
		}
		return Ipv6Weak
	} else if na.IP.To4() != nil {
		if o.Routable() && o.IP.To4() != nil {
			return Ipv4
		}
		return Default
	}

	var tunnelled bool
	// Is our v6 is tunnelled?
	if o.RFC3964() || o.RFC6052() || o.RFC6145() {
		tunnelled = true
	}
	if !o.Routable() {
		return Default
	} else if o.RFC4380() {
		return Teredo
	} else if o.IP.To4() != nil {
		return Ipv4
	} else if tunnelled {
		// only prioritise ipv6 if we aren't tunnelling it.
		return Ipv6Weak
	}
	return Ipv6Strong
}

var rfc1918_10 = net.IPNet{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(8, 32)}
var rfc1918_192 = net.IPNet{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 32)}
var rfc1918_172 = net.IPNet{IP: net.ParseIP("172.16.0.0"), Mask: net.CIDRMask(12, 32)}
var rfc3849 = net.IPNet{IP: net.ParseIP("2001:0DB8::"), Mask: net.CIDRMask(32, 128)}
var rfc3927 = net.IPNet{IP: net.ParseIP("169.254.0.0"), Mask: net.CIDRMask(16, 32)}
var rfc3964 = net.IPNet{IP: net.ParseIP("2002::"), Mask: net.CIDRMask(16, 128)}
var rfc4193 = net.IPNet{IP: net.ParseIP("FC00::"), Mask: net.CIDRMask(7, 128)}
var rfc4380 = net.IPNet{IP: net.ParseIP("2001::"), Mask: net.CIDRMask(32, 128)}
var rfc4843 = net.IPNet{IP: net.ParseIP("2001:10::"), Mask: net.CIDRMask(28, 128)}
var rfc4862 = net.IPNet{IP: net.ParseIP("FE80::"), Mask: net.CIDRMask(64, 128)}
var rfc6052 = net.IPNet{IP: net.ParseIP("64:FF9B::"), Mask: net.CIDRMask(96, 128)}
var rfc6145 = net.IPNet{IP: net.ParseIP("::FFFF:0:0:0"), Mask: net.CIDRMask(96, 128)}
var zero4 = net.IPNet{IP: net.ParseIP("0.0.0.0"), Mask: net.CIDRMask(8, 32)}

func (na *NetAddress) RFC1918() bool {
	return rfc1918_10.Contains(na.IP) || rfc1918_192.Contains(na.IP) || rfc1918_172.Contains(na.IP)
}

func (na *NetAddress) RFC3849() bool {
	return rfc3849.Contains(na.IP)
}

func (na *NetAddress) RFC3927() bool {
	return rfc3927.Contains(na.IP)
}

func (na *NetAddress) RFC3964() bool {
	return rfc3964.Contains(na.IP)
}

func (na *NetAddress) RFC4193() bool {
	return rfc4193.Contains(na.IP)
}

func (na *NetAddress) RFC4380() bool {
	return rfc4380.Contains(na.IP)
}

func (na *NetAddress) RFC4843() bool {
	return rfc4843.Contains(na.IP)
}

func (na *NetAddress) RFC4862() bool {
	return rfc4862.Contains(na.IP)
}

func (na *NetAddress) RFC6052() bool {
	return rfc6052.Contains(na.IP)
}

func (na *NetAddress) RFC6145() bool {
	return rfc6145.Contains(na.IP)
}
