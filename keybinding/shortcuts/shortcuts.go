/**
 * Copyright (C) 2016 Deepin Technology Co., Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 **/

package shortcuts

import (
	"gir/gio-2.0"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"pkg.deepin.io/dde/daemon/keybinding/keybind"
	"pkg.deepin.io/dde/daemon/keybinding/xrecord"
	"pkg.deepin.io/lib/log"
	"strings"
	"sync"
)

var logger *log.Logger

const (
	SKLCtrlShift uint32 = 1 << iota
	SKLAltShift
	SKLSuperSpace
)

func SetLogger(l *log.Logger) {
	logger = l
}

type KeyEventFunc func(ev *KeyEvent)

type Shortcuts struct {
	idShortcutMap       map[string]Shortcut
	grabedKeyAccelMap   map[Key]*Accel
	xu                  *xgbutil.XUtil
	xRecordEventHandler *XRecordEventHandler
	eventCb             KeyEventFunc
	eventCbMu           sync.Mutex
}

type KeyEvent struct {
	Mods     Modifiers
	Code     Keycode
	Shortcut Shortcut
}

func NewShortcuts(xu *xgbutil.XUtil, eventCb KeyEventFunc) *Shortcuts {
	ss := &Shortcuts{
		idShortcutMap:     make(map[string]Shortcut),
		grabedKeyAccelMap: make(map[Key]*Accel),
		eventCb:           eventCb,
		xu:                xu,
	}

	ss.xRecordEventHandler = NewXRecordEventHandler(xu)
	ss.xRecordEventHandler.modKeyReleasedCb = func(code uint8, mods uint16) {
		if isKbdAlreadyGrabbed(ss.xu) {
			return
		}
		switch mods {
		case xproto.ModMaskLock, xproto.ModMask2, xproto.ModMask4:
			// caps_lock, num_lock, super
			ss.emitKeyEvent(0, Key{Code: Keycode(code)})

		case xproto.ModMaskControl | xproto.ModMaskShift:
			// ctrl-shift
			ss.emitFakeKeyEvent(&Action{Type: ActionTypeSwitchKbdLayout, Arg: SKLCtrlShift})

		case xproto.ModMask1 | xproto.ModMaskShift:
			// alt-shift
			ss.emitFakeKeyEvent(&Action{Type: ActionTypeSwitchKbdLayout, Arg: SKLAltShift})
		}
	}
	// init package xrecord
	err := xrecord.Initialize()
	if err == nil {
		xrecord.KeyEventCallback = ss.handleXRecordKeyEvent
		xrecord.ButtonEventCallback = ss.handleXRecordButtonEvent
	} else {
		logger.Warning(err)
	}
	return ss
}

func (ss *Shortcuts) Destroy() {
	xrecord.KeyEventCallback = nil
	xrecord.ButtonEventCallback = nil
	xrecord.Finalize()
}

func (ss *Shortcuts) List() (list []Shortcut) {
	// RLock ss.idShortcutMap
	for _, shortcut := range ss.idShortcutMap {
		list = append(list, shortcut)
	}
	return
}

func (ss *Shortcuts) grabAccel(shortcut Shortcut, pa ParsedAccel, dummy bool) {
	key, err := pa.QueryKey(ss.xu)
	if err != nil {
		logger.Debugf("getAccel failed shortcut: %v, pa: %v, err: %v", shortcut.GetId(), pa, err)
		return
	}
	//logger.Debugf("grabAccel shortcut: %s, pa: %s, key: %s, dummy: %v", shortcut.GetId(), pa, key, dummy)

	// RLock ss.grabedKeyAccelMap
	if confAccel, ok := ss.grabedKeyAccelMap[key]; ok {
		// conflict
		logger.Debugf("key %v is grabed by %v", key, confAccel.Shortcut.GetId())
		return
	}

	// no conflict
	if !dummy {
		err = key.Grab(ss.xu)
		if err != nil {
			logger.Debug(err)
			return
		}
	}
	accel := &Accel{
		Parsed:    pa,
		Shortcut:  shortcut,
		GrabedKey: key,
	}
	// Lock ss.grabedKeyAccelMap
	// attach key <-> accel
	ss.grabedKeyAccelMap[key] = accel
}

func (ss *Shortcuts) ungrabAccel(pa ParsedAccel, dummy bool) {
	key, err := pa.QueryKey(ss.xu)
	if err != nil {
		logger.Debug(err)
		return
	}

	// Lock ss.grabedKeyAccelMap
	delete(ss.grabedKeyAccelMap, key)
	key.Ungrab(ss.xu)
}

func (ss *Shortcuts) grabShortcut(shortcut Shortcut) {
	//logger.Debug("grabShortcut shortcut id:", shortcut.GetId())
	for _, pa := range shortcut.GetAccels() {
		dummy := dummyGrab(shortcut, pa)
		ss.grabAccel(shortcut, pa, dummy)
	}
}

func (ss *Shortcuts) ungrabShortcut(shortcut Shortcut) {

	for _, pa := range shortcut.GetAccels() {
		dummy := dummyGrab(shortcut, pa)
		ss.ungrabAccel(pa, dummy)
	}
}

func (ss *Shortcuts) ModifyShortcutAccels(shortcut Shortcut, newAccels []ParsedAccel) {
	logger.Debug("Shortcuts.ModifyShortcutAccels", shortcut, newAccels)
	ss.ungrabShortcut(shortcut)
	shortcut.setAccels(newAccels)
	ss.grabShortcut(shortcut)
}

func (ss *Shortcuts) AddShortcutAccel(shortcut Shortcut, pa ParsedAccel) {
	logger.Debug("Shortcuts.AddShortcutAccel", shortcut, pa)
	newAccels := shortcut.GetAccels()
	newAccels = append(newAccels, pa)
	shortcut.setAccels(newAccels)

	// grab accel
	dummy := dummyGrab(shortcut, pa)
	ss.grabAccel(shortcut, pa, dummy)
}

func (ss *Shortcuts) RemoveShortcutAccel(shortcut Shortcut, pa ParsedAccel) {
	logger.Debug("Shortcuts.RemoveShortcutAccel", shortcut, pa)
	logger.Debug("shortcut.GetAccel", shortcut.GetAccels())
	var newAccels []ParsedAccel
	for _, accel := range shortcut.GetAccels() {
		if !accel.Equal(ss.xu, pa) {
			newAccels = append(newAccels, accel)
		}
	}
	shortcut.setAccels(newAccels)
	logger.Debug("shortcut.setAccels", newAccels)

	// ungrab accel
	dummy := dummyGrab(shortcut, pa)
	ss.ungrabAccel(pa, dummy)
}

func dummyGrab(shortcut Shortcut, pa ParsedAccel) bool {
	if shortcut.GetType() == ShortcutTypeWM {
		return true
	}

	switch strings.ToLower(pa.Key) {
	case "super_l", "super_r", "caps_lock", "num_lock":
		return true
	}
	return false
}

func (ss *Shortcuts) UngrabAll() {
	// ungrab all grabed keys
	// Lock ss.grabedKeyAccelMap
	for grabedKey, accel := range ss.grabedKeyAccelMap {
		dummy := dummyGrab(accel.Shortcut, accel.Parsed)
		if !dummy {
			grabedKey.Ungrab(ss.xu)
		}
	}
	// new map
	count := len(ss.grabedKeyAccelMap)
	ss.grabedKeyAccelMap = make(map[Key]*Accel, count)
}

func (ss *Shortcuts) GrabAll() {
	// re-grab all shortcuts
	for _, shortcut := range ss.idShortcutMap {
		ss.grabShortcut(shortcut)
	}
}

func (ss *Shortcuts) regrabAll() {
	logger.Debug("regrabAll")
	ss.UngrabAll()
	ss.GrabAll()
}

func (ss *Shortcuts) ReloadAllShortcutAccels() []Shortcut {
	var changes []Shortcut
	for _, shortcut := range ss.idShortcutMap {
		changed := shortcut.ReloadAccels()
		if changed {
			changes = append(changes, shortcut)
		}
	}
	return changes
}

// shift, control, alt(mod1), super(mod4)
func getConcernedMods(state uint16) uint16 {
	var mods uint16
	if state&xproto.ModMaskShift > 0 {
		mods |= xproto.ModMaskShift
	}
	if state&xproto.ModMaskControl > 0 {
		mods |= xproto.ModMaskControl
	}
	if state&xproto.ModMask1 > 0 {
		mods |= xproto.ModMask1
	}
	if state&xproto.ModMask4 > 0 {
		mods |= xproto.ModMask4
	}
	return mods
}

func GetConcernedModifiers(state uint16) Modifiers {
	return Modifiers(getConcernedMods(state))
}

func combineStateCode2Key(state uint16, code uint8) Key {
	mods := GetConcernedModifiers(state)
	_code := Keycode(code)
	key := Key{
		Mods: mods,
		Code: _code,
	}
	return key
}

func (ss *Shortcuts) callEventCallback(ev *KeyEvent) {
	ss.eventCbMu.Lock()
	ss.eventCb(ev)
	ss.eventCbMu.Unlock()
}

func (ss *Shortcuts) handleKeyEvent(pressed bool, detail xproto.Keycode, state uint16) {
	key := combineStateCode2Key(state, uint8(detail))
	logger.Debug("event key:", key)

	if pressed {
		// key press
		ss.emitKeyEvent(Modifiers(state), key)
	}
}

func (ss *Shortcuts) emitFakeKeyEvent(action *Action) {
	keyEvent := &KeyEvent{
		Shortcut: NewFakeShortcut(action),
	}
	ss.callEventCallback(keyEvent)
}

func (ss *Shortcuts) emitKeyEvent(mods Modifiers, key Key) {
	// RLock ss.grabedKeyAccelMap
	accel, ok := ss.grabedKeyAccelMap[key]
	if ok {
		logger.Debugf("accel: %#v", accel)
		keyEvent := &KeyEvent{
			Mods:     mods,
			Code:     key.Code,
			Shortcut: accel.Shortcut,
		}

		ss.callEventCallback(keyEvent)
	} else {
		logger.Debug("accel not found")
	}
}

func isKbdAlreadyGrabbed(xu *xgbutil.XUtil) bool {
	var grabWin xproto.Window

	if activeWin, _ := ewmh.ActiveWindowGet(xu); activeWin == 0 {
		grabWin = xu.RootWin()
	} else {
		// check viewable
		attrs, err := xproto.GetWindowAttributes(xu.Conn(), activeWin).Reply()
		if err != nil {
			grabWin = xu.RootWin()
		} else if attrs.MapState != xproto.MapStateViewable {
			// err is nil and activeWin is not viewable
			grabWin = xu.RootWin()
		} else {
			// err is nil, activeWin is viewable
			grabWin = activeWin
		}
	}

	err := keybind.GrabKeyboard(xu, grabWin)
	if err == nil {
		// grab keyboard successful
		keybind.UngrabKeyboard(xu)
		return false
	}

	logger.Warningf("GrabKeyboard win %d failed: %v", grabWin, err)

	gkErr, ok := err.(keybind.GrabKeyboardError)
	if ok && gkErr.Status == xproto.GrabStatusAlreadyGrabbed {
		return true
	}
	return false
}

func (ss *Shortcuts) SetAllModKeysReleasedCallback(cb func()) {
	ss.xRecordEventHandler.allModKeysReleasedCb = cb
}

func (ss *Shortcuts) handleXRecordKeyEvent(pressed bool, code uint8, state uint16) {
	ss.xRecordEventHandler.handleKeyEvent(pressed, code, state)
	if pressed {
		// Special handling screenshot* shortcuts
		key := combineStateCode2Key(state, code)
		accel, ok := ss.grabedKeyAccelMap[key]
		if ok {
			shortcut := accel.Shortcut
			if shortcut != nil && shortcut.GetType() == ShortcutTypeSystem &&
				strings.HasPrefix(shortcut.GetId(), "screenshot") {
				keyEvent := &KeyEvent{
					Mods:     key.Mods,
					Code:     key.Code,
					Shortcut: shortcut,
				}
				logger.Debug("handleXRecordKeyEvent: emit key event for screenshot* shortcuts")
				ss.callEventCallback(keyEvent)
			}
		}
	}
}

func (ss *Shortcuts) handleXRecordButtonEvent(pressed bool) {
	ss.xRecordEventHandler.handleButtonEvent(pressed)
}

func (ss *Shortcuts) ListenXEvents() {
	xevent.KeyPressFun(func(xu *xgbutil.XUtil, ev xevent.KeyPressEvent) {
		logger.Debug(ev)
		ss.handleKeyEvent(true, ev.Detail, ev.State)
	}).Connect(ss.xu, ss.xu.RootWin())

	xevent.KeyReleaseFun(func(xu *xgbutil.XUtil, ev xevent.KeyReleaseEvent) {
		logger.Debug(ev)
		ss.handleKeyEvent(false, ev.Detail, ev.State)
	}).Connect(ss.xu, ss.xu.RootWin())

	keybind.KbdMappingNotifyCallback = ss.regrabAll
}

func (ss *Shortcuts) Add(shortcut Shortcut) {
	ss.AddWithoutLock(shortcut)
}

// TODO private
func (ss *Shortcuts) AddWithoutLock(shortcut Shortcut) {
	logger.Debug("AddWithoutLock", shortcut)
	uid := shortcut.GetUid()
	// Lock ss.idShortcutMap
	ss.idShortcutMap[uid] = shortcut
	ss.grabShortcut(shortcut)
}

func (ss *Shortcuts) Delete(shortcut Shortcut) {
	uid := shortcut.GetUid()
	// Lock ss.idShortcutMap
	delete(ss.idShortcutMap, uid)
	ss.ungrabShortcut(shortcut)
}

func (ss *Shortcuts) GetByIdType(id string, _type int32) Shortcut {
	uid := idType2Uid(id, _type)
	// Lock ss.idShortcutMap
	shortcut := ss.idShortcutMap[uid]
	return shortcut
}

// ret0: Conflicting Accel
// ret1: pa parse key error
func (ss *Shortcuts) FindConflictingAccel(pa ParsedAccel) (*Accel, error) {
	key, err := pa.QueryKey(ss.xu)
	if err != nil {
		return nil, err
	}

	logger.Debug("Shortcuts.FindConflictingAccel pa:", pa)
	logger.Debug("key:", key)

	// RLock ss.grabedKeyAccelMap
	accel, ok := ss.grabedKeyAccelMap[key]
	if ok {
		return accel, nil
	}
	return nil, nil
}

func (ss *Shortcuts) AddSystem(gsettings *gio.Settings) {
	logger.Debug("AddSystem")
	idNameMap := getSystemIdNameMap()
	for _, id := range gsettings.ListKeys() {
		name := idNameMap[id]
		if name == "" {
			name = id
		}
		accels := gsettings.GetStrv(id)
		gs := NewGSettingsShortcut(gsettings, id, ShortcutTypeSystem, accels, name)
		sysShortcut := &SystemShortcut{
			GSettingsShortcut: gs,
			arg: &ActionExecCmdArg{
				Cmd: getSystemAction(id),
			},
		}
		ss.AddWithoutLock(sysShortcut)
	}
}

func (ss *Shortcuts) AddWM(gsettings *gio.Settings) {
	logger.Debug("AddWM")
	idNameMap := getWMIdNameMap()
	for _, id := range gsettings.ListKeys() {
		name := idNameMap[id]
		if name == "" {
			name = id
		}
		accels := gsettings.GetStrv(id)
		gs := NewGSettingsShortcut(gsettings, id, ShortcutTypeWM, accels, name)
		ss.AddWithoutLock(gs)
	}
}

func (ss *Shortcuts) AddMedia(gsettings *gio.Settings) {
	logger.Debug("AddMedia")
	idNameMap := getMediaIdNameMap()
	for _, id := range gsettings.ListKeys() {
		name := idNameMap[id]
		if name == "" {
			name = id
		}
		accels := gsettings.GetStrv(id)
		gs := NewGSettingsShortcut(gsettings, id, ShortcutTypeMedia, accels, name)
		mediaShortcut := &MediaShortcut{
			GSettingsShortcut: gs,
		}
		ss.AddWithoutLock(mediaShortcut)
	}
}

func (ss *Shortcuts) AddCustom(csm *CustomShortcutManager) {
	logger.Debug("AddCustom")
	for _, shortcut := range csm.List() {
		ss.AddWithoutLock(shortcut)
	}
}

func (ss *Shortcuts) AddSpecial() {
	idNameMap := getSpecialIdNameMap()

	// add SwitchKbdLayout <Super>Space
	s0 := NewFakeShortcut(&Action{Type: ActionTypeSwitchKbdLayout, Arg: SKLSuperSpace})
	pa, err := ParseStandardAccel("<Super>Space")
	if err != nil {
		panic(err)
	}
	s0.Id = "switch-kbd-layout"
	s0.Name = idNameMap[s0.Id]
	s0.Accels = []ParsedAccel{pa}
	ss.AddWithoutLock(s0)
}
