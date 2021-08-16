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
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
	"xfsgo/p2p/upnp"

	"github.com/sirupsen/logrus"
	cmn "github.com/tendermint/tmlibs/common"
)

const (
	logModule              = "p2p"
	numBufferedConnections = 10
	defaultExternalPort    = 8770
	tryListenTimes         = 5
)

//Listener subset of the methods of DefaultListener
type Listener interface {
	Connections() <-chan net.Conn
	InternalAddress() *NetAddress
	ExternalAddress() *NetAddress
	String() string
	// Stop() bool
}

// Defaults to tcp
func protocolAndAddress(listenAddr string) (string, string) {
	p, address := "tcp", listenAddr
	parts := strings.SplitN(address, "://", 2)
	if len(parts) == 2 {
		p, address = parts[0], parts[1]
	}
	return p, address
}

// GetListener get listener and listen address.
func GetListener(ListenAddress string) (Listener, string) {
	p, address := protocolAndAddress(ListenAddress)
	l, listenerStatus := NewDefaultListener(p, address, true)

	if listenerStatus {
		return l, cmn.Fmt("%v:%v", l.ExternalAddress().IP.String(), l.ExternalAddress().Port)
	}
	return l, cmn.Fmt("%v:%v", l.InternalAddress().IP.String(), l.InternalAddress().Port)
}

//getUPNPExternalAddress UPNP external address discovery & port mapping
func getUPNPExternalAddress(externalPort, internalPort int) (*NetAddress, error) {
	nat, err := upnp.Discover()
	if err != nil {
		return nil, errors.New("could not perform UPNP discover")
	}

	ext, err := nat.GetExternalAddress()
	if err != nil {
		return nil, errors.New("could not perform UPNP external address")
	}

	if externalPort == 0 {
		externalPort = defaultExternalPort
	}
	externalPort, err = nat.AddPortMapping("tcp", externalPort, internalPort, "bytomd tcp", 0)
	if err != nil {
		return nil, errors.New("could not add tcp UPNP port mapping")
	}
	externalPort, err = nat.AddPortMapping("udp", externalPort, internalPort, "bytomd udp", 0)
	if err != nil {
		return nil, errors.New("could not add udp UPNP port mapping")
	}
	return NewNetAddressIPPort(ext, uint16(externalPort)), nil
}

func splitHostPort(addr string) (host string, port int) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		cmn.PanicSanity(err)
	}
	port, err = strconv.Atoi(portStr)
	if err != nil {
		cmn.PanicSanity(err)
	}
	return host, port
}

//DefaultListener Implements bytomd server Listener
type DefaultListener struct {
	cmn.BaseService

	listener    net.Listener
	intAddr     *NetAddress
	extAddr     *NetAddress
	connections chan net.Conn
}

func ExternalIPv4() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		// interface down
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		// loopback interface
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "127.0.0.1", nil
}

//NewDefaultListener create a default listener
func NewDefaultListener(protocol string, lAddr string, skipUPNP bool) (Listener, bool) {
	// Local listen IP & port
	lAddrIP, lAddrPort := splitHostPort(lAddr)

	listener, err := net.Listen(protocol, lAddr)
	for i := 0; i < tryListenTimes && err != nil; i++ {
		time.Sleep(time.Second * 1)
		listener, err = net.Listen(protocol, lAddr)
	}
	if err != nil {
		logrus.Panic(err)
	}

	intAddr, err := NewNetAddressString(lAddr)
	if err != nil {
		logrus.Panic(err)
	}

	// Actual listener local IP & port
	listenerIP, listenerPort := splitHostPort(listener.Addr().String())
	logrus.Info("Local listener", " ip:", listenerIP, " port:", listenerPort)

	// Determine external address...
	var extAddr *NetAddress
	var upnpMap bool = false

	if !skipUPNP && (lAddrIP == "" || lAddrIP == "0.0.0.0") {
		extAddr, err = getUPNPExternalAddress(lAddrPort, listenerPort)
		if err == nil {
			upnpMap = true
		}
		logrus.WithFields(logrus.Fields{"module": logModule, "err": err}).Info("get UPNP external address")
	}

	// Get the IPv4 available
	if extAddr == nil {
		if ip, err := ExternalIPv4(); err != nil {
			logrus.WithFields(logrus.Fields{"module": logModule, "err": err}).Warning("get ipv4 external address")
			logrus.Panic("get ipv4 external address fail!")
		} else {
			extAddr = NewNetAddressIPPort(net.ParseIP(ip), uint16(lAddrPort))
			logrus.WithFields(logrus.Fields{"module": logModule, "addr": extAddr}).Info("get ipv4 external address success")
		}
	}

	dl := &DefaultListener{
		listener:    listener,
		intAddr:     intAddr,
		extAddr:     extAddr,
		connections: make(chan net.Conn, numBufferedConnections),
	}

	dl.BaseService = *cmn.NewBaseService(nil, "DefaultListener", dl)
	dl.Start() // Started upon construction
	if upnpMap {
		return dl, true
	}

	conn, err := net.DialTimeout("tcp", extAddr.String(), 3*time.Second)
	if err != nil {
		return dl, false
	}
	conn.Close()

	return dl, true
}

//OnStart start listener
func (l *DefaultListener) OnStart() error {
	l.BaseService.OnStart()
	go l.listenRoutine()
	return nil
}

//OnStop stop listener
func (l *DefaultListener) OnStop() {
	l.BaseService.OnStop()
	l.listener.Close()
}

// func (l *DefaultListener) Stop() bool {
// 	if l.listener.Close() != nil {
// 		return true
// 	}
// 	return false
// }

//listenRoutine Accept connections and pass on the channel
func (l *DefaultListener) listenRoutine() {
	for {
		conn, err := l.listener.Accept()
		if !l.IsRunning() {
			break // Go to cleanup
		}
		// listener wasn't stopped,
		// yet we encountered an error.
		if err != nil {
			logrus.Panic(err)
		}
		l.connections <- conn
	}
	// Cleanup
	close(l.connections)
}

//Connections a channel of inbound connections. It gets closed when the listener closes.
func (l *DefaultListener) Connections() <-chan net.Conn {
	return l.connections
}

//InternalAddress listener internal address
func (l *DefaultListener) InternalAddress() *NetAddress {
	return l.intAddr
}

//ExternalAddress listener external address for remote peer dial
func (l *DefaultListener) ExternalAddress() *NetAddress {
	return l.extAddr
}

// NetListener the returned listener is already Accept()'ing. So it's not suitable to pass into http.Serve().
func (l *DefaultListener) NetListener() net.Listener {
	return l.listener
}

//String string of default listener
func (l *DefaultListener) String() string {
	return fmt.Sprintf("Listener(@%v)", l.extAddr)
}
