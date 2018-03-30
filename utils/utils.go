package utils

import (
	"reflect"
	"strings"
	"net"
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
	switch strings.ToLower(network){
	case "tcp":
	case "tcp4":
	case "tcp6":
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

	return
}
