/**
 * Copyright (C) 2016 Deepin Technology Co., Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 **/

package dock

import (
	"github.com/BurntSushi/xgbutil/xrect"
)

type HideModeType int32

const (
	HideModeKeepShowing HideModeType = iota
	HideModeKeepHidden
	HideModeAutoHide // invalid
	HideModeSmartHide
)

func (t HideModeType) String() string {
	switch t {
	case HideModeKeepShowing:
		return "Keep showing mode"
	case HideModeKeepHidden:
		return "Keep hidden mode"
	case HideModeAutoHide:
		return "Auto hide mode"
	case HideModeSmartHide:
		return "Smart hide mode"
	default:
		return "Unknown mode"
	}
}

type HideStateType int32

const (
	HideStateUnknown HideStateType = iota
	HideStateShow
	HideStateHide
)

func (s HideStateType) String() string {
	switch s {
	case HideStateShow:
		return "Show"
	case HideStateHide:
		return "Hide"
	default:
		return "Unknown"
	}
}

type DisplayModeType int32

const (
	DisplayModeFashionMode DisplayModeType = iota
	DisplayModeEfficientMode
	DisplayModeClassicMode
)

func (t DisplayModeType) String() string {
	switch t {
	case DisplayModeFashionMode:
		return "Fashion mode"
	case DisplayModeEfficientMode:
		return "Efficient mode"
	case DisplayModeClassicMode:
		return "Classic mode"
	default:
		return "Unknown mode"
	}
}

type positionType int32

const (
	positionTop positionType = iota
	positionRight
	positionBottom
	positionLeft
)

func (p positionType) String() string {
	switch p {
	case positionTop:
		return "Top"
	case positionRight:
		return "Right"
	case positionBottom:
		return "Bottom"
	case positionLeft:
		return "Left"
	default:
		return "Unknown"
	}
}

type Rect struct {
	X, Y          int32
	Width, Height uint32
}

func NewRect() *Rect {
	return &Rect{}
}

func (r *Rect) ToXRect() xrect.Rect {
	return xrect.New(int(r.X), int(r.Y), int(r.Width), int(r.Height))
}
