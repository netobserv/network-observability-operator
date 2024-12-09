package agent

import (
	"fmt"
	"net"
	"net/netip"
	"regexp"
	"strings"
)

type InterfaceFilter interface {
	Allowed(iface string) (bool, error)
}

type ipInterfaceFilter struct {
	allowedIPs []netip.Prefix
	// Almost always going to be a wrapper around getting
	// the interface from net.InterfaceByName and then calling
	// .Addrs() on the interface
	ipsFromIface func(ifaceName string) ([]netip.Addr, error)
}

// Default function for getting the list of IPs configured
// for a specific network interface
func IPsFromInterface(ifaceName string) ([]netip.Addr, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return []netip.Addr{}, fmt.Errorf("error retrieving interface by name: %w", err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return []netip.Addr{}, fmt.Errorf("error retrieving addresses from interface: %w", err)
	}

	interfaceAddrs := []netip.Addr{}
	for _, addr := range addrs {
		prefix, err := netip.ParsePrefix(addr.String())
		if err != nil {
			return []netip.Addr{}, fmt.Errorf("parsing given ip to netip.Addr: %w", err)
		}
		interfaceAddrs = append(interfaceAddrs, prefix.Addr())
	}
	return interfaceAddrs, nil
}

// initIPInterfaceFilter allows filtering network interfaces that are accepted/excluded by the user,
// according to the provided INTERFACE_IPS from the configuration. It allows interfaces where at least
// one of the provided CIDRs are associated with it.
func initIPInterfaceFilter(ips []string, ipsFromIface func(ifaceName string) ([]netip.Addr, error)) (ipInterfaceFilter, error) {
	ipIfaceFilter := ipInterfaceFilter{}
	ipIfaceFilter.ipsFromIface = ipsFromIface

	for _, ip := range ips {
		prefix, err := netip.ParsePrefix(ip)
		if err != nil {
			return ipInterfaceFilter{}, fmt.Errorf("error parsing given ip: %s: %w", ip, err)
		}
		ipIfaceFilter.allowedIPs = append(ipIfaceFilter.allowedIPs, prefix)
	}

	return ipIfaceFilter, nil
}

func (f *ipInterfaceFilter) Allowed(iface string) (bool, error) {
	ifaceAddrs, err := f.ipsFromIface(iface)
	if err != nil {
		return false, fmt.Errorf("error calling ipsFromIface(): %w", err)
	}

	for _, ifaceAddr := range ifaceAddrs {
		for _, allowedPrefix := range f.allowedIPs {
			if allowedPrefix.Contains(ifaceAddr) {
				return true, nil
			}
		}
	}
	return false, nil
}

type regexpInterfaceFilter struct {
	allowedRegexpes  []*regexp.Regexp
	allowedMatches   []string
	excludedRegexpes []*regexp.Regexp
	excludedMatches  []string
}

// initRegexpInterfaceFilter allows filtering network interfaces that are accepted/excluded by the user,
// according to the provided allowed and excluded interfaces from the configuration. It allows
// matching by exact string or by regular expression
func initRegexpInterfaceFilter(allowed, excluded []string) (regexpInterfaceFilter, error) {
	var isRegexp = regexp.MustCompile("^/(.*)/$")

	itf := regexpInterfaceFilter{}
	for _, definition := range allowed {
		definition = strings.Trim(definition, " ")
		// the user defined a /regexp/ between slashes: compile and store it as regular expression
		if sm := isRegexp.FindStringSubmatch(definition); len(sm) > 1 {
			re, err := regexp.Compile(sm[1])
			if err != nil {
				return itf, fmt.Errorf("wrong interface regexp %q: %w", definition, err)
			}
			itf.allowedRegexpes = append(itf.allowedRegexpes, re)
		} else {
			// otherwise, store it as exact match definition
			itf.allowedMatches = append(itf.allowedMatches, definition)
		}
	}

	for _, definition := range excluded {
		definition = strings.Trim(definition, " ")
		// the user defined a /regexp/ between slashes: compile and store it as regexp
		if sm := isRegexp.FindStringSubmatch(definition); len(sm) > 1 {
			re, err := regexp.Compile(sm[1])
			if err != nil {
				return itf, fmt.Errorf("wrong excluded interface regexp %q: %w", definition, err)
			}
			itf.excludedRegexpes = append(itf.excludedRegexpes, re)
		} else {
			// otherwise, store it as exact match definition
			itf.excludedMatches = append(itf.excludedMatches, definition)
		}
	}

	return itf, nil
}

func (itf *regexpInterfaceFilter) Allowed(name string) (bool, error) {
	// if the allowed list is empty, any interface is allowed except if it matches the exclusion list
	allowed := len(itf.allowedRegexpes)+len(itf.allowedMatches) == 0
	// otherwise, we check if it appears in the allowed lists (both exact match and regexp)
	for i := 0; !allowed && i < len(itf.allowedMatches); i++ {
		allowed = allowed || name == itf.allowedMatches[i]
	}
	for i := 0; !allowed && i < len(itf.allowedRegexpes); i++ {
		allowed = allowed || itf.allowedRegexpes[i].MatchString(string(name))
	}
	if !allowed {
		return false, nil
	}
	// if the interface matches the allow lists, we still need to check that is not excluded
	for _, match := range itf.excludedMatches {
		if name == match {
			return false, nil
		}
	}
	for _, re := range itf.excludedRegexpes {
		if re.MatchString(string(name)) {
			return false, nil
		}
	}
	return true, nil
}
