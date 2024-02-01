package network

import (
	"fmt"
	"strconv"
)

// portType is unexported to ensure consistency -
// use the exported consts System, Registered, Dynamic
// see also: https://en.wikipedia.org/wiki/List_of_TCP_and_UDP_port_numbers
type portType int

const (
	System     portType = iota // system or well-known ports
	Registered                 // registered ports
	Dynamic                    // dynamic, private or ephemeral ports
)

// portTypeRange is unexported to ensure consistency (min <= max) -
// use the exported variables All, NonSystem, NonDynamic
type portTypeRange struct {
	min, max portType
}

// Port is an interface to ensure consistency -
// use NewPort(), NewPortForType() or NewPortForTypeRange() functions
// to obtain a valid Port
type Port interface {
	IsSet() bool
	IsValid() bool
	IsValidForType(portType) bool
	IsValidForTypeRange(*portTypeRange) bool
	Uint64() uint64
}

type port uint64

const (
	MinSystem port = iota
)

const (
	MaxSystem port = iota + 1023
	MinRegistered
)

const (
	MaxRegistered port = iota + 49151
	MinDynamic
)

const (
	MaxDynamic port = iota + 65535
	Invalid
)

var All = rangeOf(System, Dynamic)
var NonSystem = rangeOf(Registered, Dynamic)
var NonDynamic = rangeOf(System, Registered)

type portRange struct {
	min, max port
}

var ranges = map[portType]*portRange{
	System:     {min: MinSystem, max: MaxSystem},
	Registered: {min: MinRegistered, max: MaxRegistered},
	Dynamic:    {min: MinDynamic, max: MaxDynamic},
}

func (p port) IsSet() bool {
	return p < Invalid
}

func (p port) IsValid() bool {
	return p.IsValidForTypeRange(All)
}

func (p port) IsValidForType(pt portType) bool {
	return p.IsValidForTypeRange(rangeOfSame(pt))
}

func (p port) IsValidForTypeRange(ptr *portTypeRange) bool {
	return ptr != nil &&
		p >= ranges[ptr.min].min &&
		p <= ranges[ptr.max].max
}

func (p port) Uint64() uint64 {
	return uint64(p)
}

type PortInput interface {
	~uint64 | ~string
}

// NewPort returns a Port if the argument has a valid TCP/UDP port number
// (no limitation of port type or type range), error otherwise
func NewPort[PI PortInput](pi PI) (Port, error) {
	return NewPortForTypeRange(pi, All)
}

// NewPortForType returns a Port if the argument has a valid TCP/UDP port number
// for the requested port type, error otherwise
func NewPortForType[PI PortInput](pi PI, pt portType) (Port, error) {
	return NewPortForTypeRange(pi, rangeOfSame(pt))
}

// NewPortForTypeRange returns a Port if the argument has a valid TCP/UDP port number
// for the requested port type range, error otherwise
func NewPortForTypeRange[PI PortInput](pi PI, ptr *portTypeRange) (p Port, err error) {
	var n uint64
	switch v := any(pi).(type) {
	case string:
		n, err = strconv.ParseUint(v, 10, 64)
	case uint64:
		n = v
	}
	if err == nil {
		if candidate := port(n); candidate.IsValidForTypeRange(ptr) {
			p = candidate
		} else {
			err = fmt.Errorf("invalid port %d", n)
		}
	}
	return
}

func rangeOfSame(pt portType) *portTypeRange {
	return rangeOf(pt, pt)
}

func rangeOf(min, max portType) *portTypeRange {
	return &portTypeRange{min: min, max: max}
}
