/**
 * Copyright (C) 2014 Deepin Technology Co., Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 **/

package dock

import (
	"github.com/BurntSushi/xgbutil/ewmh"
	"pkg.deepin.io/lib/appinfo/desktopappinfo"
	"pkg.deepin.io/lib/dbus"
	. "pkg.deepin.io/lib/gettext"
)

func (e *AppEntry) setMenu(menu *Menu) {
	e.coreMenu = menu
	menuJSON := menu.GenerateJSON()
	// set menu JSON
	if e.Menu != menuJSON {
		e.Menu = menuJSON
		dbus.NotifyChange(e, "Menu")
	}
}

func (entry *AppEntry) updateMenu() {
	logger.Debug("Update menu")
	menu := NewMenu()
	menu.AppendItem(entry.getMenuItemLaunch())

	desktopActionMenuItems := entry.getMenuItemDesktopActions()
	menu.AppendItem(desktopActionMenuItems...)

	if entry.hasWindow() {
		menu.AppendItem(entry.getMenuItemCloseAll())
		menu.AppendItem(entry.getMenuItemAllWindows())
	}

	// menu item dock or undock
	logger.Debug(entry.Id, "Item docked?", entry.IsDocked)
	if entry.IsDocked {
		menu.AppendItem(entry.getMenuItemUndock())
	} else {
		menu.AppendItem(entry.getMenuItemDock())
	}

	entry.setMenu(menu)
}

func (entry *AppEntry) getMenuItemDesktopActions() []*MenuItem {
	ai := entry.appInfo
	if ai == nil {
		return nil
	}

	var items []*MenuItem
	launchAction := func(action desktopappinfo.DesktopAction) func(timestamp uint32) {
		return func(timestamp uint32) {
			logger.Debugf("launch action %+v", action)
			ctx := entry.dockManager.launchContext
			ctx.SetTimestamp(timestamp)
			action.Launch(nil, ctx)

			entry.dockManager.markAppLaunched(ai)
		}
	}

	for _, action := range ai.GetActions() {
		item := NewMenuItem(action.Name, launchAction(action), true)
		items = append(items, item)
	}
	return items
}

func (entry *AppEntry) launchApp(timestamp uint32) {
	logger.Debug("launchApp timestamp:", timestamp)
	if entry.appInfo != nil {
		logger.Debug("Has AppInfo")

		ctx := entry.dockManager.launchContext
		ctx.SetTimestamp(timestamp)
		entry.appInfo.Launch(nil, ctx)

		entry.dockManager.markAppLaunched(entry.appInfo)
	} else {
		// TODO
		logger.Debug("not supported")
	}
}

func (entry *AppEntry) getMenuItemLaunch() *MenuItem {
	var itemName string
	if entry.hasWindow() {
		itemName = entry.getDisplayName()
	} else {
		itemName = Tr("_Open")
	}
	logger.Debugf("getMenuItemLaunch, itemName: %q", itemName)
	return NewMenuItem(itemName, entry.launchApp, true)
}

func (entry *AppEntry) getMenuItemCloseAll() *MenuItem {
	return NewMenuItem(Tr("_Close All"), func(timestamp uint32) {
		logger.Debug("Close All")
		for win, _ := range entry.windows {
			ewmh.CloseWindow(XU, win)
		}
	}, true)
}

func (entry *AppEntry) getMenuItemDock() *MenuItem {
	return NewMenuItem(Tr("_Dock"), func(uint32) {
		logger.Debug("menu action dock entry")
		entry.RequestDock()
	}, true)
}

func (entry *AppEntry) getMenuItemUndock() *MenuItem {
	return NewMenuItem(Tr("_Undock"), func(uint32) {
		logger.Debug("menu action undock entry")
		entry.RequestUndock()
	}, true)
}

func (entry *AppEntry) getMenuItemAllWindows() *MenuItem {
	return NewMenuItem(Tr("_All windows"), func(uint32) {
		logger.Debug("menu action all windows")
		entry.PresentWindows()
	}, true)
}
