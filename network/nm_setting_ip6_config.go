/**
 * Copyright (C) 2014 Deepin Technology Co., Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 **/

package network

import (
	"fmt"
	"pkg.deepin.io/dde/daemon/network/nm"
	. "pkg.deepin.io/lib/gettext"
)

func initSettingSectionIpv6(data connectionData) {
	addSetting(data, nm.NM_SETTING_IP6_CONFIG_SETTING_NAME)
	setSettingIP6ConfigMethod(data, nm.NM_SETTING_IP6_CONFIG_METHOD_AUTO)
}

// Initialize available values
var availableValuesIp6ConfigMethod = make(availableValues)

func initAvailableValuesIp6() {
	availableValuesIp6ConfigMethod[nm.NM_SETTING_IP6_CONFIG_METHOD_IGNORE] = kvalue{nm.NM_SETTING_IP6_CONFIG_METHOD_IGNORE, Tr("Ignore")}
	availableValuesIp6ConfigMethod[nm.NM_SETTING_IP6_CONFIG_METHOD_AUTO] = kvalue{nm.NM_SETTING_IP6_CONFIG_METHOD_AUTO, Tr("Auto")}
	availableValuesIp6ConfigMethod[nm.NM_SETTING_IP6_CONFIG_METHOD_DHCP] = kvalue{nm.NM_SETTING_IP6_CONFIG_METHOD_DHCP, Tr("DHCP")}
	availableValuesIp6ConfigMethod[nm.NM_SETTING_IP6_CONFIG_METHOD_LINK_LOCAL] = kvalue{nm.NM_SETTING_IP6_CONFIG_METHOD_LINK_LOCAL, Tr("Link-Local Only")}
	availableValuesIp6ConfigMethod[nm.NM_SETTING_IP6_CONFIG_METHOD_MANUAL] = kvalue{nm.NM_SETTING_IP6_CONFIG_METHOD_MANUAL, Tr("Manual")}
	availableValuesIp6ConfigMethod[nm.NM_SETTING_IP6_CONFIG_METHOD_SHARED] = kvalue{nm.NM_SETTING_IP6_CONFIG_METHOD_SHARED, Tr("Shared")}
}

// Get available keys
func getSettingIP6ConfigAvailableKeys(data connectionData) (keys []string) {
	keys = appendAvailableKeys(data, keys, nm.NM_SETTING_IP6_CONFIG_SETTING_NAME, nm.NM_SETTING_IP_CONFIG_METHOD)
	method := getSettingIP6ConfigMethod(data)
	switch method {
	default:
		logger.Error("ip6 config method is invalid:", method)
	case nm.NM_SETTING_IP6_CONFIG_METHOD_IGNORE:
	case nm.NM_SETTING_IP6_CONFIG_METHOD_AUTO:
		keys = appendAvailableKeys(data, keys, nm.NM_SETTING_IP6_CONFIG_SETTING_NAME, nm.NM_SETTING_IP_CONFIG_DNS)
	case nm.NM_SETTING_IP6_CONFIG_METHOD_DHCP: // ignore
		keys = appendAvailableKeys(data, keys, nm.NM_SETTING_IP6_CONFIG_SETTING_NAME, nm.NM_SETTING_IP_CONFIG_DNS)
	case nm.NM_SETTING_IP6_CONFIG_METHOD_LINK_LOCAL: // ignore
	case nm.NM_SETTING_IP6_CONFIG_METHOD_MANUAL:
		keys = appendAvailableKeys(data, keys, nm.NM_SETTING_IP6_CONFIG_SETTING_NAME, nm.NM_SETTING_IP_CONFIG_DNS)
		keys = appendAvailableKeys(data, keys, nm.NM_SETTING_IP6_CONFIG_SETTING_NAME, nm.NM_SETTING_IP_CONFIG_ADDRESSES)
	case nm.NM_SETTING_IP6_CONFIG_METHOD_SHARED:
	}
	return
}

// Get available values
func getSettingIP6ConfigAvailableValues(data connectionData, key string) (values []kvalue) {
	switch key {
	case nm.NM_SETTING_IP_CONFIG_METHOD:
		// values = []string{
		// 	// nm.NM_SETTING_IP6_CONFIG_METHOD_IGNORE, // ignore
		// 	nm.NM_SETTING_IP6_CONFIG_METHOD_AUTO,
		// 	// nm.NM_SETTING_IP6_CONFIG_METHOD_DHCP, // ignore
		// 	// nm.NM_SETTING_IP6_CONFIG_METHOD_LINK_LOCAL, // ignore
		// 	nm.NM_SETTING_IP6_CONFIG_METHOD_MANUAL,
		// 	// nm.NM_SETTING_IP6_CONFIG_METHOD_SHARED,// ignore
		// }
		if getSettingConnectionType(data) == nm.NM_SETTING_VPN_SETTING_NAME {
			values = []kvalue{
				availableValuesIp6ConfigMethod[nm.NM_SETTING_IP6_CONFIG_METHOD_AUTO],
				availableValuesIp6ConfigMethod[nm.NM_SETTING_IP6_CONFIG_METHOD_IGNORE],
			}
		} else {
			values = []kvalue{
				availableValuesIp6ConfigMethod[nm.NM_SETTING_IP6_CONFIG_METHOD_AUTO],
				availableValuesIp6ConfigMethod[nm.NM_SETTING_IP6_CONFIG_METHOD_MANUAL],
				availableValuesIp6ConfigMethod[nm.NM_SETTING_IP6_CONFIG_METHOD_IGNORE],
			}
		}
	}
	return
}

// Check whether the values are correct
func checkSettingIP6ConfigValues(data connectionData) (errs sectionErrors) {
	errs = make(map[string]string)

	// check method
	ensureSettingIP6ConfigMethodNoEmpty(data, errs)
	switch getSettingIP6ConfigMethod(data) {
	default:
		rememberError(errs, nm.NM_SETTING_IP6_CONFIG_SETTING_NAME, nm.NM_SETTING_IP_CONFIG_METHOD, nmKeyErrorInvalidValue)
		return
	case nm.NM_SETTING_IP6_CONFIG_METHOD_IGNORE:
		checkSettingIP6MethodConflict(data, errs)
	case nm.NM_SETTING_IP6_CONFIG_METHOD_AUTO:
	case nm.NM_SETTING_IP6_CONFIG_METHOD_DHCP: // ignore
	case nm.NM_SETTING_IP6_CONFIG_METHOD_LINK_LOCAL: // ignore
		checkSettingIP6MethodConflict(data, errs)
	case nm.NM_SETTING_IP6_CONFIG_METHOD_MANUAL:
		ensureSettingIP6ConfigAddressesNoEmpty(data, errs)
	case nm.NM_SETTING_IP6_CONFIG_METHOD_SHARED:
		checkSettingIP6MethodConflict(data, errs)
	}

	// check value of dns
	checkSettingIP6ConfigDns(data, errs)

	// check value of address
	checkSettingIP6ConfigAddresses(data, errs)

	return
}

func checkSettingIP6MethodConflict(data connectionData, errs sectionErrors) {
	// check dns
	if isSettingIP6ConfigDnsExists(data) && len(getSettingIP6ConfigDns(data)) > 0 {
		rememberError(errs, nm.NM_SETTING_IP6_CONFIG_SETTING_NAME, nm.NM_SETTING_IP_CONFIG_DNS, fmt.Sprintf(nmKeyErrorIp6MethodConflict, nm.NM_SETTING_IP_CONFIG_DNS))
	}
	// check dns search
	if isSettingIP6ConfigDnsSearchExists(data) && len(getSettingIP6ConfigDnsSearch(data)) > 0 {
		rememberError(errs, nm.NM_SETTING_IP6_CONFIG_SETTING_NAME, nm.NM_SETTING_IP_CONFIG_DNS_SEARCH, fmt.Sprintf(nmKeyErrorIp6MethodConflict, nm.NM_SETTING_IP_CONFIG_DNS_SEARCH))
	}
	// check address
	if isSettingIP6ConfigAddressesExists(data) && len(getSettingIP6ConfigAddresses(data)) > 0 {
		rememberError(errs, nm.NM_SETTING_IP6_CONFIG_SETTING_NAME, nm.NM_SETTING_IP_CONFIG_ADDRESSES, fmt.Sprintf(nmKeyErrorIp6MethodConflict, nm.NM_SETTING_IP_CONFIG_ADDRESSES))
	}
	// check route
	if isSettingIP6ConfigRoutesExists(data) && len(getSettingIP6ConfigRoutes(data)) > 0 {
		rememberError(errs, nm.NM_SETTING_IP6_CONFIG_SETTING_NAME, nm.NM_SETTING_IP_CONFIG_ROUTES, fmt.Sprintf(nmKeyErrorIp6MethodConflict, nm.NM_SETTING_IP_CONFIG_ROUTES))
	}
}

func checkSettingIP6ConfigDns(data connectionData, errs sectionErrors) {
	if !isSettingIP6ConfigDnsExists(data) {
		return
	}
	dnses := getSettingIP6ConfigDns(data)
	for _, dns := range dnses {
		if !isIpv6AddressValid(dns) {
			rememberError(errs, nm.NM_SETTING_IP6_CONFIG_SETTING_NAME, nm.NM_SETTING_VK_IP6_CONFIG_DNS, nmKeyErrorInvalidValue)
			return
		}
		if isIpv6AddressZero(dns) {
			rememberError(errs, nm.NM_SETTING_IP6_CONFIG_SETTING_NAME, nm.NM_SETTING_VK_IP6_CONFIG_DNS, nmKeyErrorInvalidValue)
			return
		}
	}
}

func checkSettingIP6ConfigAddresses(data connectionData, errs sectionErrors) {
	if !isSettingIP6ConfigAddressesExists(data) {
		return
	}
	addresses := getSettingIP6ConfigAddresses(data)
	for _, addr := range addresses {
		// check address
		if !isIpv6AddressValid(addr.Address) {
			rememberError(errs, nm.NM_SETTING_IP6_CONFIG_SETTING_NAME, nm.NM_SETTING_VK_IP6_CONFIG_ADDRESSES_ADDRESS, nmKeyErrorInvalidValue)
			logger.Warning(nmKeyErrorInvalidValue, addr.Address)
		}
		if isIpv6AddressZero(addr.Address) {
			rememberError(errs, nm.NM_SETTING_IP6_CONFIG_SETTING_NAME, nm.NM_SETTING_VK_IP6_CONFIG_ADDRESSES_ADDRESS, nmKeyErrorInvalidValue)
		}
		// check prefix
		if addr.Prefix < 1 || addr.Prefix > 128 {
			rememberError(errs, nm.NM_SETTING_IP6_CONFIG_SETTING_NAME, nm.NM_SETTING_VK_IP6_CONFIG_ADDRESSES_PREFIX, nmKeyErrorInvalidValue)
		}
		// check gateway
		if !isIpv6AddressValid(addr.Gateway) {
			rememberError(errs, nm.NM_SETTING_IP6_CONFIG_SETTING_NAME, nm.NM_SETTING_VK_IP6_CONFIG_ADDRESSES_GATEWAY, nmKeyErrorInvalidValue)
		}
	}
}

// Logic setter
func logicSetSettingIP6ConfigMethod(data connectionData, value string) (err error) {
	switch value {
	case nm.NM_SETTING_IP6_CONFIG_METHOD_IGNORE:
		removeSettingKeyBut(data, nm.NM_SETTING_IP6_CONFIG_SETTING_NAME, nm.NM_SETTING_IP_CONFIG_METHOD)
	case nm.NM_SETTING_IP6_CONFIG_METHOD_AUTO:
		removeSettingIP6ConfigAddresses(data)
	case nm.NM_SETTING_IP6_CONFIG_METHOD_DHCP: // ignore
	case nm.NM_SETTING_IP6_CONFIG_METHOD_LINK_LOCAL: // ignore
		removeSettingIP6ConfigDns(data)
		removeSettingIP6ConfigDnsSearch(data)
		removeSettingIP6ConfigAddresses(data)
		removeSettingIP6ConfigRoutes(data)
	case nm.NM_SETTING_IP6_CONFIG_METHOD_MANUAL:
	case nm.NM_SETTING_IP6_CONFIG_METHOD_SHARED:
		removeSettingIP6ConfigDns(data)
		removeSettingIP6ConfigDnsSearch(data)
		removeSettingIP6ConfigAddresses(data)
		removeSettingIP6ConfigRoutes(data)
	}
	setSettingIP6ConfigMethod(data, value)
	return
}

// Virtual key utility
func isSettingIP6ConfigAddressesEmpty(data connectionData) bool {
	addresses := getSettingIP6ConfigAddresses(data)
	if len(addresses) == 0 {
		return true
	}
	return false
}
func getOrNewSettingIP6ConfigAddresses(data connectionData) (addresses ipv6Addresses) {
	if !isSettingIP6ConfigAddressesEmpty(data) {
		addresses = getSettingIP6ConfigAddresses(data)
	} else {
		addresses = make(ipv6Addresses, 1)
		addresses[0].Gateway = make([]byte, 16)
	}
	return
}

// Virtual key getter
func getSettingVkIp6ConfigDns(data connectionData) (value string) {
	return getSettingCacheKeyString(data, nm.NM_SETTING_IP6_CONFIG_SETTING_NAME, nm.NM_SETTING_VK_IP6_CONFIG_DNS)
}
func getSettingVkIp6ConfigDns2(data connectionData) (value string) {
	return getSettingCacheKeyString(data, nm.NM_SETTING_IP6_CONFIG_SETTING_NAME, nm.NM_SETTING_VK_IP6_CONFIG_DNS2)
}
func getSettingVkIp6ConfigAddressesAddress(data connectionData) (value string) {
	if isSettingIP6ConfigAddressesEmpty(data) {
		return
	}
	addresses := getSettingIP6ConfigAddresses(data)
	if isIpv6AddressValid(addresses[0].Address) {
		value = convertIpv6AddressToString(addresses[0].Address)
	}
	return
}
func getSettingVkIp6ConfigAddressesPrefix(data connectionData) (value uint32) {
	if isSettingIP6ConfigAddressesEmpty(data) {
		return
	}
	addresses := getSettingIP6ConfigAddresses(data)
	value = addresses[0].Prefix
	return
}
func getSettingVkIp6ConfigAddressesGateway(data connectionData) (value string) {
	if isSettingIP6ConfigAddressesEmpty(data) {
		return
	}
	addresses := getSettingIP6ConfigAddresses(data)
	value = convertIpv6AddressToStringNoZero(addresses[0].Gateway)
	return
}
func getSettingVkIp6ConfigRoutesAddress(data connectionData) (value string) {
	// TODO value := getSettingIP6ConfigRoutesAddress(data)
	return
}
func getSettingVkIp6ConfigRoutesPrefix(data connectionData) (value string) {
	// TODO value := getSettingIP6ConfigRoutesPrefix(data)
	return
}
func getSettingVkIp6ConfigRoutesNexthop(data connectionData) (value string) {
	// TODO value := getSettingIP6ConfigRoutesNexthop(data)
	return
}
func getSettingVkIp6ConfigRoutesMetric(data connectionData) (value string) {
	// TODO value := getSettingIP6ConfigRoutesMetric(data)
	return
}

// Virtual key logic setter
func logicSetSettingVkIp6ConfigDns(data connectionData, value string) (err error) {
	setSettingCacheKey(data, nm.NM_SETTING_IP6_CONFIG_SETTING_NAME, nm.NM_SETTING_VK_IP6_CONFIG_DNS, value)
	if len(value) > 0 {
		if _, errWrap := convertIpv6AddressToArrayByteCheck(value); errWrap != nil {
			err = fmt.Errorf(nmKeyErrorInvalidValue)
		}
	}
	return
}
func logicSetSettingVkIp6ConfigDns2(data connectionData, value string) (err error) {
	setSettingCacheKey(data, nm.NM_SETTING_IP6_CONFIG_SETTING_NAME, nm.NM_SETTING_VK_IP6_CONFIG_DNS2, value)
	if len(value) > 0 {
		if _, errWrap := convertIpv6AddressToArrayByteCheck(value); errWrap != nil {
			err = fmt.Errorf(nmKeyErrorInvalidValue)
		}
	}
	return
}
func logicSetSettingVkIp6ConfigAddressesAddress(data connectionData, value string) (err error) {
	if len(value) == 0 {
		value = ipv6AddrZero
	}
	tmp, err := convertIpv6AddressToArrayByteCheck(value)
	if err != nil {
		err = fmt.Errorf(nmKeyErrorInvalidValue)
		return
	}
	addresses := getOrNewSettingIP6ConfigAddresses(data)
	addr := addresses[0]
	addr.Address = tmp
	addresses[0] = addr
	if !isIpv6AddressStructZero(addr) {
		setSettingIP6ConfigAddresses(data, addresses)
	} else {
		removeSettingIP6ConfigAddresses(data)
	}
	return
}
func logicSetSettingVkIp6ConfigAddressesPrefix(data connectionData, value uint32) (err error) {
	addresses := getOrNewSettingIP6ConfigAddresses(data)
	addr := addresses[0]
	addr.Prefix = value
	addresses[0] = addr
	if !isIpv6AddressStructZero(addr) {
		setSettingIP6ConfigAddresses(data, addresses)
	} else {
		removeSettingIP6ConfigAddresses(data)
	}
	return
}
func logicSetSettingVkIp6ConfigAddressesGateway(data connectionData, value string) (err error) {
	if len(value) == 0 {
		value = ipv6AddrZero
	}
	tmp, err := convertIpv6AddressToArrayByteCheck(value)
	if err != nil {
		err = fmt.Errorf(nmKeyErrorInvalidValue)
		return
	}
	addresses := getOrNewSettingIP6ConfigAddresses(data)
	addr := addresses[0]
	addr.Gateway = tmp
	addresses[0] = addr
	if !isIpv6AddressStructZero(addr) {
		setSettingIP6ConfigAddresses(data, addresses)
	} else {
		removeSettingIP6ConfigAddresses(data)
	}
	return
}
func logicSetSettingVkIp6ConfigRoutesAddress(data connectionData, value string) (err error) {
	// TODO setSettingIP6ConfigRoutesAddressJSON(data)
	return
}
func logicSetSettingVkIp6ConfigRoutesPrefix(data connectionData, value uint32) (err error) {
	// TODO setSettingIP6ConfigRoutesPrefixJSON(data)
	return
}
func logicSetSettingVkIp6ConfigRoutesNexthop(data connectionData, value string) (err error) {
	// TODO setSettingIP6ConfigRoutesNexthopJSON(data)
	return
}
func logicSetSettingVkIp6ConfigRoutesMetric(data connectionData, value uint32) (err error) {
	// TODO setSettingIP6ConfigRoutesMetricJSON(data)
	return
}
