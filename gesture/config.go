/**
 * Copyright (C) 2016 Deepin Technology Co., Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 **/

package gesture

import (
	"encoding/json"
	"fmt"
	"gir/gio-2.0"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	dutils "pkg.deepin.io/lib/utils"
	"sync"
)

const (
	ActionTypeShortcut    string = "shortcut"
	ActionTypeCommandline        = "commandline"
)

var (
	configSystemPath = "/usr/share/dde-daemon/gesture.json"
	configUserPath   = os.Getenv("HOME") + "/.config/deepin/dde-daemon/gesture.json"

	gestureSchemaId = "com.deepin.dde.gesture"
	gsKeyEnabled    = "enabled"
)

type ActionInfo struct {
	Type   string
	Action string
}

type gestureInfo struct {
	Name      string
	Direction string
	Fingers   int32
	Action    ActionInfo
}
type gestureInfos []*gestureInfo

type gestureManager struct {
	locker   sync.RWMutex
	userFile string
	Infos    gestureInfos

	setting *gio.Settings
	enabled bool
}

func newGestureManager() (*gestureManager, error) {
	var filename = configUserPath
	if !dutils.IsFileExist(configUserPath) {
		filename = configSystemPath
	}

	infos, err := newGestureInfosFromFile(filename)
	if err != nil {
		return nil, err
	}

	setting, err := dutils.CheckAndNewGSettings(gestureSchemaId)
	if err != nil {
		return nil, err
	}

	return &gestureManager{
		userFile: configUserPath,
		Infos:    infos,
		setting:  setting,
		enabled:  setting.GetBoolean(gsKeyEnabled),
	}, nil
}

func (m *gestureManager) Exec(name, direction string, fingers int32) error {
	m.locker.RLock()
	defer m.locker.RUnlock()

	if !m.enabled {
		logger.Debug("Gesture had been disabled")
		return nil
	}

	info := m.Infos.Get(name, direction, fingers)
	if info == nil {
		return fmt.Errorf("Not found gesture info for: %s, %s, %d", name, direction, fingers)
	}

	var cmd = info.Action.Action
	if info.Action.Type == ActionTypeShortcut {
		cmd = fmt.Sprintf("xdotool key %s", cmd)
	}
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(out))
	}
	return nil
}

func (m *gestureManager) Write() error {
	m.locker.Lock()
	defer m.locker.Unlock()
	err := os.MkdirAll(path.Dir(m.userFile), 0755)
	if err != nil {
		return err
	}
	data, err := json.Marshal(m.Infos)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(m.userFile, data, 0644)
}

func (m *gestureManager) handleGSettingsChanged() {
	m.setting.Connect("changed", func(s *gio.Settings, key string) {
		switch key {
		case gsKeyEnabled:
			m.enabled = m.setting.GetBoolean(key)
		}
	})
}

func (infos gestureInfos) Get(name, direction string, fingers int32) *gestureInfo {
	for _, info := range infos {
		if info.Name == name && info.Direction == direction &&
			info.Fingers == fingers {
			return info
		}
	}
	return nil
}

func (infos gestureInfos) Set(name, direction string, fingers int32, action ActionInfo) error {
	info := infos.Get(name, direction, fingers)
	if info == nil {
		return fmt.Errorf("Not found gesture info for: %s, %s, %d", name, direction, fingers)
	}
	info.Action = action
	return nil
}

func newGestureInfosFromFile(filename string) (gestureInfos, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if len(content) == 0 {
		return nil, fmt.Errorf("File '%s' is empty", filename)
	}

	var infos gestureInfos
	err = json.Unmarshal(content, &infos)
	if err != nil {
		return nil, err
	}
	return infos, nil
}
