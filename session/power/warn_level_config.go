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
	"gir/gio-2.0"
	"pkg.deepin.io/lib/dbus"
	"time"
)

type WarnLevelConfig struct {
	UsePercentageForPolicy                              bool
	LowTime, CriticalTime, ActionTime                   uint64
	LowPercentage, CriticalPercentage, ActionPercentage float64

	settings    *gio.Settings
	changeTimer *time.Timer
	changeCb    func()
}

func NewWarnLevelConfig() *WarnLevelConfig {
	c := &WarnLevelConfig{}
	return c
}

func (c *WarnLevelConfig) setChangeCallback(fn func()) {
	c.changeCb = fn
}

func (c *WarnLevelConfig) connectSettings(s *gio.Settings) {
	c.settings = s
	c.UsePercentageForPolicy = s.GetBoolean(settingKeyUsePercentageForPolicy)

	c.LowPercentage = float64(s.GetInt(settingKeyLowPercentage))
	c.CriticalPercentage = float64(s.GetInt(settingKeyCriticalPercentage))
	c.ActionPercentage = float64(s.GetInt(settingKeyActionPercentage))

	c.LowTime = uint64(s.GetInt(settingKeyLowTime))
	c.CriticalTime = uint64(s.GetInt(settingKeyCriticalTime))
	c.ActionTime = uint64(s.GetInt(settingKeyActionTime))
	c.connectSettingsChange()
}
func (c *WarnLevelConfig) connectSettingsKeyChange(key string, handler func(*gio.Settings, string)) {
	logger.Debug("connect change", key)
	c.settings.Connect("changed::"+key, handler)
}

func (c *WarnLevelConfig) delayCheckValid() {
	logger.Debug("delayCheckValid")
	if c.changeTimer != nil {
		c.changeTimer.Stop()
	}
	c.changeTimer = time.AfterFunc(20*time.Second, func() {
		logger.Debug("checkValid")
		if !c.isValid() {
			logger.Info("Warn level config is invalid, reset")
			c.Reset()
		}
	})
}

func (c *WarnLevelConfig) notifyChange(propName string) {
	if c.changeCb != nil {
		logger.Debug("WarnLevelConfig change")
		c.changeCb()
	}
	c.delayCheckValid()
	dbus.NotifyChange(c, propName)
}

func (c *WarnLevelConfig) getChangeHandlerFloat64(propName string, propRef *float64) func(*gio.Settings, string) {
	return func(s *gio.Settings, key string) {
		logger.Debug("change key", key)
		newVal := float64(s.GetInt(key))
		if newVal != *propRef {
			*propRef = newVal
			c.notifyChange(propName)
		}
	}
}

func (c *WarnLevelConfig) getChangeHandlerUInt64(propName string, propRef *uint64) func(*gio.Settings, string) {
	return func(s *gio.Settings, key string) {
		logger.Debug("change key", key)
		newVal := uint64(s.GetInt(key))
		if newVal != *propRef {
			*propRef = newVal
			c.notifyChange(propName)
		}
	}
}

func (c *WarnLevelConfig) getChangeHandlerBoolean(propName string, propRef *bool) func(*gio.Settings, string) {
	return func(s *gio.Settings, key string) {
		logger.Debug("change key", key)
		newVal := s.GetBoolean(key)
		if newVal != *propRef {
			*propRef = newVal
			c.notifyChange(propName)
		}
	}
}

func (c *WarnLevelConfig) connectSettingsChange() {
	c.connectSettingsKeyChange(settingKeyUsePercentageForPolicy, c.getChangeHandlerBoolean("UsePercentageForPolicy", &c.UsePercentageForPolicy))
	c.connectSettingsKeyChange(settingKeyLowPercentage, c.getChangeHandlerFloat64("LowPercentage", &c.LowPercentage))
	c.connectSettingsKeyChange(settingKeyCriticalPercentage, c.getChangeHandlerFloat64("CriticalPercentage", &c.CriticalPercentage))
	c.connectSettingsKeyChange(settingKeyActionPercentage, c.getChangeHandlerFloat64("ActionPercentage", &c.ActionPercentage))

	c.connectSettingsKeyChange(settingKeyLowTime, c.getChangeHandlerUInt64("LowTime", &c.LowTime))
	c.connectSettingsKeyChange(settingKeyCriticalTime, c.getChangeHandlerUInt64("CriticalTime", &c.CriticalTime))
	c.connectSettingsKeyChange(settingKeyActionTime, c.getChangeHandlerUInt64("ActionTime", &c.ActionTime))
}

func (c *WarnLevelConfig) GetDBusInfo() dbus.DBusInfo {
	return dbus.DBusInfo{
		Dest:       dbusDest,
		ObjectPath: dbusPath,
		Interface:  dbusIFC + ".WarnLevelConfig",
	}
}

func (c *WarnLevelConfig) Reset() {
	s := c.settings
	s.Reset(settingKeyUsePercentageForPolicy)
	s.Reset(settingKeyLowPercentage)
	s.Reset(settingKeyCriticalPercentage)
	s.Reset(settingKeyActionPercentage)
	s.Reset(settingKeyLowTime)
	s.Reset(settingKeyCriticalTime)
	s.Reset(settingKeyActionTime)
}

func (c *WarnLevelConfig) isValid() bool {
	if c.LowTime > c.CriticalTime &&
		c.CriticalTime > c.ActionTime &&

		c.LowPercentage > c.CriticalPercentage &&
		c.CriticalPercentage > c.ActionPercentage {
		return true
	}
	return false
}
