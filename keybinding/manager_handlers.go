/**
 * Copyright (C) 2016 Deepin Technology Co., Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 **/

package keybinding

import (
	"fmt"
	. "pkg.deepin.io/dde/daemon/keybinding/shortcuts"
	"time"
)

func (m *Manager) initHandlers() {
	m.handlers = make([]KeyEventFunc, ActionTypeCount)
	logger.Debug("initHandlers", len(m.handlers))

	m.handlers[ActionTypeNonOp] = func(ev *KeyEvent) {
		logger.Debug("non-Op do nothing")
	}

	m.handlers[ActionTypeExecCmd] = func(ev *KeyEvent) {
		action := ev.Shortcut.GetAction()
		arg, ok := action.Arg.(*ActionExecCmdArg)
		if !ok {
			logger.Warning(ErrTypeAssertionFail)
			return
		}

		go func() {
			err := execCmd(arg.Cmd)
			if err != nil {
				logger.Warning("execCmd error:", err)
			}
		}()
	}

	m.handlers[ActionTypeShowNumLockOSD] = func(ev *KeyEvent) {
		state, err := queryNumLockState(m.xu)
		if err != nil {
			logger.Warning(err)
			return
		}
		save := m.keyboardSetting.GetBoolean(gsKeySaveNumLockState)
		switch state {
		case NumLockOn:
			if save {
				m.NumLockState.Set(int32(NumLockOn))
			}
			showOSD("NumLockOn")
		case NumLockOff:
			if save {
				m.NumLockState.Set(int32(NumLockOff))
			}
			showOSD("NumLockOff")
		}
	}

	m.handlers[ActionTypeShowCapsLockOSD] = func(ev *KeyEvent) {
		if !canShowCapsOSD() {
			return
		}

		state, err := queryCapsLockState(m.xu)
		if err != nil {
			logger.Warning(err)
			return
		}

		switch state {
		case CapsLockOff:
			showOSD("CapsLockOff")
		case CapsLockOn:
			showOSD("CapsLockOn")
		}
	}

	m.handlers[ActionTypeOpenMimeType] = func(ev *KeyEvent) {
		action := ev.Shortcut.GetAction()
		mimeType, ok := action.Arg.(string)
		if !ok {
			logger.Warning(ErrTypeAssertionFail)
			return
		}

		go func() {
			err := execCmd(queryCommandByMime(mimeType))
			if err != nil {
				logger.Warning("execCmd error:", err)
			}
		}()
	}

	m.handlers[ActionTypeAudioCtrl] = buildHandlerFromController(m.audioController)
	m.handlers[ActionTypeMediaPlayerCtrl] = buildHandlerFromController(m.mediaPlayerController)
	m.handlers[ActionTypeDisplayCtrl] = buildHandlerFromController(m.displayController)
	m.handlers[ActionTypeKbdLightCtrl] = buildHandlerFromController(m.kbdLightController)
	m.handlers[ActionTypeTouchpadCtrl] = buildHandlerFromController(m.touchpadController)
	logger.Debug("-----------ManageWireless:", ManageWireless)
	if ManageWireless == "enabled" {
		m.handlers[ActionTypeToggleWireless] = func(ev *KeyEvent) {
			err := toggleWireless()
			if err != nil {
				logger.Warning("Failed to toggle wireless:", err)
			}
		}
	}

	m.handlers[ActionTypeSystemShutdown] = func(ev *KeyEvent) {
		cmd := getPowerButtonPressedExec()
		if err := execCmd(cmd); err != nil {
			logger.Warning(err)
		}
	}

	m.handlers[ActionTypeSystemSuspend] = func(ev *KeyEvent) {
		systemSuspend()
	}

	// handle Switch Kbd Layout
	m.handlers[ActionTypeSwitchKbdLayout] = func(ev *KeyEvent) {
		logger.Debug("Switch Kbd Layout state", m.switchKbdLayoutState)
		flags := m.ShortcutSwitchLayout.Get()
		action := ev.Shortcut.GetAction()
		arg, ok := action.Arg.(uint32)
		if !ok {
			logger.Warning(ErrTypeAssertionFail)
			return
		}

		if arg&flags == 0 {
			return
		}

		switch m.switchKbdLayoutState {
		case SKLStateNone:
			m.switchKbdLayoutState = SKLStateWait
			go m.sklWait()

		case SKLStateWait:
			m.switchKbdLayoutState = SKLStateOSDShown
			m.terminateSKLWait()
			showOSD("SwitchLayout")

		case SKLStateOSDShown:
			showOSD("SwitchLayout")
		}
	}

	m.shortcuts.SetAllModKeysReleasedCallback(func() {
		switch m.switchKbdLayoutState {
		case SKLStateWait:
			showOSD("DirectSwitchLayout")
			m.terminateSKLWait()
		case SKLStateOSDShown:
			showOSD("SwitchLayoutDone")
		case SKLStateNone:
			return
		}
		m.switchKbdLayoutState = SKLStateNone
	})
}

func (m *Manager) sklWait() {
	defer func() {
		logger.Debug("sklWait end")
		m.sklWaitQuit = nil
	}()

	m.sklWaitQuit = make(chan int)
	timer := time.NewTimer(350 * time.Millisecond)
	select {
	case <-m.sklWaitQuit:
		return
	case <-timer.C:
		logger.Debug("timer fired")
		if m.switchKbdLayoutState == SKLStateWait {
			m.switchKbdLayoutState = SKLStateOSDShown
			showOSD("SwitchLayout")
		}
	}
}

func (m *Manager) terminateSKLWait() {
	if m.sklWaitQuit != nil {
		close(m.sklWaitQuit)
	}
}

type Controller interface {
	ExecCmd(cmd ActionCmd) error
	Name() string
}

func buildHandlerFromController(c Controller) KeyEventFunc {
	return func(ev *KeyEvent) {
		if c == nil {
			logger.Warning("controller is nil")
			return
		}
		name := c.Name()

		action := ev.Shortcut.GetAction()
		cmd, ok := action.Arg.(ActionCmd)
		if !ok {
			logger.Warning(ErrTypeAssertionFail)
			return
		}
		logger.Debugf("%v Controller exec cmd %v", name, cmd)
		if err := c.ExecCmd(cmd); err != nil {
			logger.Warning(name, "Controller exec cmd err:", err)
		}
	}
}

type ErrInvalidActionCmd struct {
	Cmd ActionCmd
}

func (err ErrInvalidActionCmd) Error() string {
	return fmt.Sprintf("invalid action cmd %v", err.Cmd)
}

type ErrIsNil struct {
	Name string
}

func (err ErrIsNil) Error() string {
	return fmt.Sprintf("%s is nil", err.Name)
}
