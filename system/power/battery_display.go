/**
 * Copyright (C) 2016 Deepin Technology Co., Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 **/

package power

import (
	"pkg.deepin.io/dde/api/powersupply/battery"
	"pkg.deepin.io/lib/dbus"
	"time"
)

func (m *Manager) refreshBatteryDisplay() {
	logger.Debug("refreshBatteryDisplay")
	defer dbus.Emit(m, "BatteryDisplayUpdate", time.Now().Unix())

	var percentage float64
	var status battery.Status
	var timeToEmpty, timeToFull uint64

	batteryCount := len(m.batteries)
	if batteryCount == 0 {
		m.resetBatteryDisplay()
		return
	} else if batteryCount == 1 {
		var bat0 *Battery
		for _, bat := range m.batteries {
			bat0 = bat
			break
		}

		// copy from bat0
		percentage = bat0.Percentage
		status = bat0.Status
		timeToEmpty = bat0.TimeToEmpty
		timeToFull = bat0.TimeToFull
	} else {
		var energyTotal, energyFullTotal, energyRateTotal float64
		for _, bat := range m.batteries {
			energyTotal += bat.Energy
			energyFullTotal += bat.EnergyFull
			energyRateTotal += bat.EnergyRate
		}
		logger.Debugf("energyTotal: %v", energyTotal)
		logger.Debugf("energyFullTotal: %v", energyFullTotal)
		logger.Debugf("energyRateTotal: %v", energyRateTotal)

		percentage = rightPercentage(energyTotal / energyFullTotal * 100.0)
		status = m.getBatteryDisplayStatus()

		if energyRateTotal > 0 {
			if status == battery.StatusDischarging {
				timeToEmpty = uint64(3600 * (energyTotal / energyRateTotal))
			} else if status == battery.StatusCharging {
				timeToFull = uint64(3600 * ((energyFullTotal - energyTotal) / energyRateTotal))
			}
		}

		/* check the remaining thime is under a set limit, to deal with broken
		primary batteries rate */
		if timeToEmpty > 240*60*60 { /* ten days for discharging */
			timeToEmpty = 0
		}
		if timeToFull > 20*60*60 { /* 20 hours for charging */
			timeToFull = 0
		}
	}

	// report
	m.setPropHasBattery(true)
	m.setPropBatteryPercentage(percentage)
	m.setPropBatteryStatus(status)
	m.setPropBatteryTimeToEmpty(timeToEmpty)
	m.setPropBatteryTimeToFull(timeToFull)

	logger.Debugf("percentage: %.1f%%", percentage)
	logger.Debug("status:", status, uint32(status))
	logger.Debugf("timeToEmpty %v (%vs), timeToFull %v (%vs)",
		time.Duration(timeToEmpty)*time.Second,
		timeToEmpty,
		time.Duration(timeToFull)*time.Second,
		timeToFull)
}

func _getBatteryDisplayStatus(batteries []*Battery) battery.Status {
	var statusSlice []battery.Status
	for _, bat := range batteries {
		statusSlice = append(statusSlice, bat.Status)
	}
	return battery.GetDisplayStatus(statusSlice)
}

func (m *Manager) getBatteryDisplayStatus() battery.Status {
	return _getBatteryDisplayStatus(m.GetBatteries())
}

func (m *Manager) resetBatteryDisplay() {
	m.setPropHasBattery(false)
	m.setPropBatteryPercentage(0)
	m.setPropBatteryTimeToFull(0)
	m.setPropBatteryTimeToEmpty(0)
	m.setPropBatteryStatus(battery.StatusUnknown)
}

func rightPercentage(val float64) float64 {
	if val < 0.0 {
		val = 0.0
	} else if val > 100.0 {
		val = 100.0
	}
	return val
}
