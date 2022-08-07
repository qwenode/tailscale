// Copyright (c) 2020 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"errors"
	"fmt"
	"net/netip"
	"os/exec"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/qwenode/tailscale/envknob"
	"github.com/qwenode/tailscale/types/logger"
	"github.com/qwenode/tailscale/util/dnsname"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"
)

const (
	ipv4RegBase = `SYSTEM\CurrentControlSet\Services\Tcpip\Parameters`
	ipv6RegBase = `SYSTEM\CurrentControlSet\Services\Tcpip6\Parameters`

	versionKey = `SOFTWARE\Microsoft\Windows NT\CurrentVersion`
)

var configureWSL = envknob.Bool("TS_DEBUG_CONFIGURE_WSL")

type windowsManager struct {
	logf       logger.Logf
	guid       string
	nrptDB     *nrptRuleDatabase
	wslManager *wslManager
}

func NewOSConfigurator(logf logger.Logf, interfaceName string) (OSConfigurator, error) {
	ret := windowsManager{
		logf:       logf,
		guid:       interfaceName,
		wslManager: newWSLManager(logf),
	}

	if isWindows10OrBetter() {
		ret.nrptDB = newNRPTRuleDatabase(logf)
	}

	// Log WSL status once at startup.
	if distros, err := wslDistros(); err != nil {
		logf("WSL: could not list distributions: %v", err)
	} else {
		logf("WSL: found %d distributions", len(distros))
	}

	return ret, nil
}

// keyOpenTimeout is how long we wait for a registry key to
// appear. For some reason, registry keys tied to ephemeral interfaces
// can take a long while to appear after interface creation, and we
// can end up racing with that.
const keyOpenTimeout = 20 * time.Second

func (m windowsManager) openKey(path string) (registry.Key, error) {
	key, err := openKeyWait(registry.LOCAL_MACHINE, path, registry.SET_VALUE, keyOpenTimeout)
	if err != nil {
		return 0, fmt.Errorf("opening %s: %w", path, err)
	}
	return key, nil
}

func (m windowsManager) ifPath(basePath string) string {
	return fmt.Sprintf(`%s\Interfaces\%s`, basePath, m.guid)
}

func delValue(key registry.Key, name string) error {
	if err := key.DeleteValue(name); err != nil && err != registry.ErrNotExist {
		return err
	}
	return nil
}

// setSplitDNS configures one or more NRPT (Name Resolution Policy Table) rules
// to resolve queries for domains using resolvers, rather than the
// system's "primary" resolver.
//
// If no resolvers are provided, the Tailscale NRPT rules are deleted.
func (m windowsManager) setSplitDNS(resolvers []netip.Addr, domains []dnsname.FQDN) error {
	if m.nrptDB == nil {
		if resolvers == nil {
			// Just a no-op in this case.
			return nil
		}
		return fmt.Errorf("Split DNS unsupported on this Windows version")
	}

	defer m.nrptDB.Refresh()
	if len(resolvers) == 0 {
		return m.nrptDB.DelAllRuleKeys()
	}

	servers := make([]string, 0, len(resolvers))
	for _, resolver := range resolvers {
		servers = append(servers, resolver.String())
	}

	return m.nrptDB.WriteSplitDNSConfig(servers, domains)
}

// setPrimaryDNS sets the given resolvers and domains as the Tailscale
// interface's DNS configuration.
// If resolvers is non-empty, those resolvers become the system's
// "primary" resolvers.
// domains can be set without resolvers, which just contributes new
// paths to the global DNS search list.
func (m windowsManager) setPrimaryDNS(resolvers []netip.Addr, domains []dnsname.FQDN) error {
	var ipsv4 []string
	var ipsv6 []string

	for _, ip := range resolvers {
		if ip.Is4() {
			ipsv4 = append(ipsv4, ip.String())
		} else {
			ipsv6 = append(ipsv6, ip.String())
		}
	}

	domStrs := make([]string, 0, len(domains))
	for _, dom := range domains {
		domStrs = append(domStrs, dom.WithoutTrailingDot())
	}

	key4, err := m.openKey(m.ifPath(ipv4RegBase))
	if err != nil {
		return err
	}
	defer key4.Close()

	if len(ipsv4) == 0 {
		if err := delValue(key4, "NameServer"); err != nil {
			return err
		}
	} else if err := key4.SetStringValue("NameServer", strings.Join(ipsv4, ",")); err != nil {
		return err
	}

	if len(domains) == 0 {
		if err := delValue(key4, "SearchList"); err != nil {
			return err
		}
	} else if err := key4.SetStringValue("SearchList", strings.Join(domStrs, ",")); err != nil {
		return err
	}

	key6, err := m.openKey(m.ifPath(ipv6RegBase))
	if err != nil {
		return err
	}
	defer key6.Close()

	if len(ipsv6) == 0 {
		if err := delValue(key6, "NameServer"); err != nil {
			return err
		}
	} else if err := key6.SetStringValue("NameServer", strings.Join(ipsv6, ",")); err != nil {
		return err
	}

	if len(domains) == 0 {
		if err := delValue(key6, "SearchList"); err != nil {
			return err
		}
	} else if err := key6.SetStringValue("SearchList", strings.Join(domStrs, ",")); err != nil {
		return err
	}

	// Disable LLMNR on the Tailscale interface. We don't do
	// multicast, and we certainly don't do LLMNR, so it's pointless
	// to make Windows try it.
	if err := key4.SetDWordValue("EnableMulticast", 0); err != nil {
		return err
	}
	if err := key6.SetDWordValue("EnableMulticast", 0); err != nil {
		return err
	}

	return nil
}

func (m windowsManager) SetDNS(cfg OSConfig) error {
	// We can configure Windows DNS in one of two ways:
	//
	//  - In primary DNS mode, we set the NameServer and SearchList
	//    registry keys on our interface. Because our interface metric
	//    is very low, this turns us into the one and only "primary"
	//    resolver for the OS, i.e. all queries flow to the
	//    resolver(s) we specify.
	//  - In split DNS mode, we set the Domain registry key on our
	//    interface (which adds that domain to the global search list,
	//    but does not contribute other DNS configuration from the
	//    interface), and configure an NRPT (Name Resolution Policy
	//    Table) rule to route queries for our suffixes to the
	//    provided resolver.
	//
	// When switching modes, we delete all the configuration related
	// to the other mode, so these two are an XOR.
	//
	// Windows actually supports much more advanced configurations as
	// well, with arbitrary routing of hosts and suffixes to arbitrary
	// resolvers. However, we use it in a "simple" split domain
	// configuration only, routing one set of things to the "split"
	// resolver and the rest to the primary.

	// Unconditionally disable dynamic DNS updates on our interfaces.
	if err := m.disableDynamicUpdates(); err != nil {
		m.logf("disableDynamicUpdates error: %v\n", err)
	}

	if len(cfg.MatchDomains) == 0 {
		if err := m.setSplitDNS(nil, nil); err != nil {
			return err
		}
		if err := m.setPrimaryDNS(cfg.Nameservers, cfg.SearchDomains); err != nil {
			return err
		}
	} else if m.nrptDB == nil {
		return errors.New("cannot set per-domain resolvers on Windows 7")
	} else {
		if err := m.setSplitDNS(cfg.Nameservers, cfg.MatchDomains); err != nil {
			return err
		}
		// Still set search domains on the interface, since NRPT only
		// handles query routing and not search domain expansion.
		if err := m.setPrimaryDNS(nil, cfg.SearchDomains); err != nil {
			return err
		}
	}

	// Force DNS re-registration in Active Directory. What we actually
	// care about is that this command invokes the undocumented hidden
	// function that forces Windows to notice that adapter settings
	// have changed, which makes the DNS settings actually take
	// effect.
	//
	// This command can take a few seconds to run, so run it async, best effort.
	//
	// After re-registering DNS, also flush the DNS cache to clear out
	// any cached split-horizon queries that are no longer the correct
	// answer.
	go func() {
		t0 := time.Now()
		m.logf("running ipconfig /registerdns ...")
		cmd := exec.Command("ipconfig", "/registerdns")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		err := cmd.Run()
		d := time.Since(t0).Round(time.Millisecond)
		if err != nil {
			m.logf("error running ipconfig /registerdns after %v: %v", d, err)
		} else {
			m.logf("ran ipconfig /registerdns in %v", d)
		}

		t0 = time.Now()
		m.logf("running ipconfig /flushdns ...")
		cmd = exec.Command("ipconfig", "/flushdns")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		err = cmd.Run()
		d = time.Since(t0).Round(time.Millisecond)
		if err != nil {
			m.logf("error running ipconfig /flushdns after %v: %v", d, err)
		} else {
			m.logf("ran ipconfig /flushdns in %v", d)
		}
	}()

	// On initial setup of WSL, the restart caused by --shutdown is slow,
	// so we do it out-of-line.
	if configureWSL {
		go func() {
			if err := m.wslManager.SetDNS(cfg); err != nil {
				m.logf("WSL SetDNS: %v", err) // continue
			} else {
				m.logf("WSL SetDNS: success")
			}
		}()
	}

	return nil
}

func (m windowsManager) SupportsSplitDNS() bool {
	return m.nrptDB != nil
}

func (m windowsManager) Close() error {
	err := m.SetDNS(OSConfig{})
	if m.nrptDB != nil {
		m.nrptDB.Close()
	}
	return err
}

// disableDynamicUpdates sets the appropriate registry values to prevent the
// Windows DHCP client from sending dynamic DNS updates for our interface to
// AD domain controllers.
func (m windowsManager) disableDynamicUpdates() error {
	setRegValue := func(regBase string) error {
		key, err := m.openKey(m.ifPath(regBase))
		if err != nil {
			return err
		}
		defer key.Close()

		return key.SetDWordValue("DisableDynamicUpdate", 1)
	}

	for _, regBase := range []string{ipv4RegBase, ipv6RegBase} {
		if err := setRegValue(regBase); err != nil {
			return err
		}
	}

	return nil
}

func (m windowsManager) GetBaseConfig() (OSConfig, error) {
	resolvers, err := m.getBasePrimaryResolver()
	if err != nil {
		return OSConfig{}, err
	}
	return OSConfig{
		Nameservers: resolvers,
		// Don't return any search domains here, because even Windows
		// 7 correctly handles blending search domains from multiple
		// sources, and any search domains we add here will get tacked
		// onto the Tailscale config unnecessarily.
	}, nil
}

// getBasePrimaryResolver returns a guess of the non-Tailscale primary
// resolver on the system.
// It's used on Windows 7 to emulate split DNS by trying to figure out
// what the "previous" primary resolver was. It might be wrong, or
// incomplete.
func (m windowsManager) getBasePrimaryResolver() (resolvers []netip.Addr, err error) {
	tsGUID, err := windows.GUIDFromString(m.guid)
	if err != nil {
		return nil, err
	}
	tsLUID, err := winipcfg.LUIDFromGUID(&tsGUID)
	if err != nil {
		return nil, err
	}
	ifrows, err := winipcfg.GetIPInterfaceTable(windows.AF_INET)
	if err == windows.ERROR_NOT_FOUND {
		// IPv4 seems disabled, try to get interface metrics from IPv6 instead.
		ifrows, err = winipcfg.GetIPInterfaceTable(windows.AF_INET6)
	}
	if err != nil {
		return nil, err
	}

	type candidate struct {
		id     winipcfg.LUID
		metric uint32
	}
	var candidates []candidate
	for _, row := range ifrows {
		if !row.Connected {
			continue
		}
		if row.InterfaceLUID == tsLUID {
			continue
		}
		candidates = append(candidates, candidate{row.InterfaceLUID, row.Metric})
	}
	if len(candidates) == 0 {
		// No resolvers set outside of Tailscale.
		return nil, nil
	}

	sort.Slice(candidates, func(i, j int) bool { return candidates[i].metric < candidates[j].metric })

	for _, candidate := range candidates {
		ips, err := candidate.id.DNS()
		if err != nil {
			return nil, err
		}

	ipLoop:
		for _, stdip := range ips {
			ip, ok := netip.AddrFromSlice(stdip)
			if !ok {
				continue
			}
			ip = ip.Unmap()
			// Skip IPv6 site-local resolvers. These are an ancient
			// and obsolete IPv6 RFC, which Windows still faithfully
			// implements. The net result is that some low-metric
			// interfaces can "have" DNS resolvers, but they're just
			// site-local resolver IPs that don't go anywhere. So, we
			// skip the site-local resolvers in order to find the
			// first interface that has real DNS servers configured.
			for _, sl := range siteLocalResolvers {
				if ip.WithZone("") == sl {
					continue ipLoop
				}
			}
			resolvers = append(resolvers, ip)
		}

		if len(resolvers) > 0 {
			// Found some resolvers, we're done.
			break
		}
	}

	return resolvers, nil
}

var siteLocalResolvers = []netip.Addr{
	netip.MustParseAddr("fec0:0:0:ffff::1"),
	netip.MustParseAddr("fec0:0:0:ffff::2"),
	netip.MustParseAddr("fec0:0:0:ffff::3"),
}

func isWindows10OrBetter() bool {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, versionKey, registry.READ)
	if err != nil {
		// Fail safe, assume old Windows.
		return false
	}
	// This key above only exists in Windows 10 and above. Its mere
	// presence is good enough.
	if _, _, err := key.GetIntegerValue("CurrentMajorVersionNumber"); err != nil {
		return false
	}
	return true
}
