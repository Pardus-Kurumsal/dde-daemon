/**
 * Copyright (C) 2014 Deepin Technology Co., Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 **/

package screenedge

import (
	gio "gir/gio-2.0"
)

type Settings struct {
	gsettings *gio.Settings
}

func NewSettings() *Settings {
	s := new(Settings)
	s.gsettings = gio.NewSettings("com.deepin.dde.zone")
	return s
}

func (s *Settings) ConnectChanged(handler func(string)) {
	s.gsettings.Connect("changed", func(s *gio.Settings, key string) {
		handler(key)
	})
}

func (s *Settings) GetDelay() int32 {
	return s.gsettings.GetInt("delay")
}

func (s *Settings) SetEdgeAction(name, value string) {
	s.gsettings.SetString(name, value)
}

func (s *Settings) GetEdgeAction(name string) string {
	return s.gsettings.GetString(name)
}

func (s *Settings) GetWhiteList() []string {
	return s.gsettings.GetStrv("white-list")
}

func (s *Settings) GetBlackList() []string {
	return s.gsettings.GetStrv("black-list")
}
