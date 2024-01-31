package network

import (
	"fmt"
	"github.com/densify-dev/retry-config/consts"
	"net"
	"strings"
)

// ParseAddress parses the input string to validate the following:
//  1. It has a mandatory IP address component in IPv4 dotted decimal, IPv6 or IPv4-mapped IPv6 form
//     (see also net.ParseIP())
//  2. It has an optional valid TCP/UDP port number (no limitation of port type or type range),
//     separated from the address component by ':'
//  3. If the port exists and the address is in IPv6 or IPv4-mapped IPv6 form, the address component MUST
//     be enclosed by square brackets ('[' and ']'), e.g. "[2001:0db8:85a3::8a2e:0370:7334]:80";
//     in all other cases, the address component MAY be enclosed by square brackets
//
// If all validations pass, the function returns the address component as a string and the Port; otherwise,
// an error is returned
func ParseAddress(s string) (string, Port, error) {
	return ParseAddressForPortTypeRange(s, All)
}

// ParseAddressForPortType behaves like ParseAddress, only that the port validation (#2) is limited
// to the specified port type
func ParseAddressForPortType(s string, pt portType) (string, Port, error) {
	return ParseAddressForPortTypeRange(s, rangeOfSame(pt))
}

// ParseAddressForPortTypeRange behaves like ParseAddress, only that the port validation (#2) is limited
// to the specified port type range
func ParseAddressForPortTypeRange(s string, ptr *portTypeRange) (address string, p Port, err error) {
	addr, po, hasPort := parseAddressPort(s)
	if ip := net.ParseIP(addr); ip == nil {
		err = fmt.Errorf("invalid IP address '%s'", addr)
		return
	}
	if hasPort {
		p, err = NewPortForTypeRange(po, ptr)
	}
	if err == nil {
		address = addr
	}
	return
}

func parseAddressPort(s string) (addr, p string, hasPort bool) {
	elems := strings.Split(s, consts.Colon)
	if l := len(elems); l < 2 {
		addr = s
	} else {
		n := l - 2
		if n == 0 || (strings.HasPrefix(elems[0], consts.LeftSquareBracket) && strings.HasSuffix(elems[n], consts.RightSquareBracket)) {
			p = elems[n+1]
			hasPort = true
		} else {
			n++
		}
		addr = strings.Join(elems[:n+1], consts.Colon)
	}
	addr = strings.TrimSuffix(strings.TrimPrefix(addr, consts.LeftSquareBracket), consts.RightSquareBracket)
	return
}
