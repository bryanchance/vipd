package server

import (
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

func (s *Server) validateVIPs() error {
	// validate interfaces on the host
	for _, vip := range s.config.VirtualIPs {
		logrus.Debugf("checking interface: %s", vip.Interface)
		if _, err := netlink.LinkByName(vip.Interface); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) updateVIPs() error {
	for _, vip := range s.config.VirtualIPs {
		iface, err := netlink.LinkByName(vip.Interface)
		if err != nil {
			return err
		}
		addrs, err := netlink.AddrList(iface, netlink.FAMILY_V4)
		if err != nil {
			return err
		}

		vipAddr, err := netlink.ParseAddr(vip.IPAddress)
		if err != nil {
			return err
		}

		exists := false
		for _, addr := range addrs {
			if vipAddr.Equal(addr) {
				logrus.Debugf("vip %s already assigned to %s", vip.IPAddress, vip.Interface)
				exists = true
				break
			}
		}

		if !exists {
			logrus.Debugf("activating vip %s on %s", vip.IPAddress, vip.Interface)
			if err := netlink.AddrAdd(iface, vipAddr); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Server) removeVIPs() error {
	for _, vip := range s.config.VirtualIPs {
		iface, err := netlink.LinkByName(vip.Interface)
		if err != nil {
			return err
		}
		addrs, err := netlink.AddrList(iface, netlink.FAMILY_V4)
		if err != nil {
			return err
		}
		vipAddr, err := netlink.ParseAddr(vip.IPAddress)
		if err != nil {
			return err
		}

		for _, addr := range addrs {
			if vipAddr.Equal(addr) {
				logrus.Debugf("removing VIP %s on %s", vip.IPAddress, vip.Interface)
				if err := netlink.AddrDel(iface, &addr); err != nil {
					return err
				}
				break
			}
		}
	}

	return nil
}
