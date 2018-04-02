package utils

import (
	"net"
	"reflect"
	"strings"
)

func HasElem(s interface{}, elem interface{}) bool {
	arrV := reflect.ValueOf(s)

	if arrV.Kind() == reflect.Slice {
		for i := 0; i < arrV.Len(); i++ {
			if arrV.Index(i).Interface() == elem {
				return true
			}
		}
	}

	return false
}

func Split2(s, sep string) (string, string, bool) {
	spl := strings.SplitN(s, sep, 2)
	if len(spl) < 2 {
		return "", "", false
	}
	return spl[0], spl[1], true
}

func ParseSchemeAddress(addr string) net.Addr {
	network, endpoint, ok := Split2(addr, "://")
	if !ok {
		return nil
	}
	switch strings.ToLower(network) {
	case "tcp", "tcp4", "tcp6":
		retAddr, err := net.ResolveTCPAddr(network, endpoint)
		if err != nil {
			return nil
		}
		return retAddr
	case "unix":
		retAddr, err := net.ResolveUnixAddr(network, endpoint)
		if err != nil {
			return nil
		}
		return retAddr
	}

	return nil
}


func CovertToSchemeAddress(addr net.Addr) string {
	if tcpAddr, ok := addr.(*net.TCPAddr); ok {
		if len(tcpAddr.IP) == net.IPv4len {
			return "tcp4://" + tcpAddr.String()
		} else if len(tcpAddr.IP) == net.IPv6len {
			return "tcp6://" + tcpAddr.String()
		}
		return ""
	} else if unixAddr, ok := addr.(*net.UnixAddr); ok {
		return "unix://" + unixAddr.String()
	}
	return ""
}