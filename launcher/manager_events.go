/**
 * Copyright (C) 2016 Deepin Technology Co., Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 **/

package launcher

import (
	"os"
	"path/filepath"
	"pkg.deepin.io/lib/appinfo/desktopappinfo"
	"pkg.deepin.io/lib/dbus"
	"time"
)

const (
	desktopFilePattern = `[^.]*.desktop`
)

func isDesktopFile(path string) bool {
	basename := filepath.Base(path)
	matched, _ := filepath.Match(desktopFilePattern, basename)
	return matched
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}

func (m *Manager) handleFsWatcherEvents() {
	watcher := m.fsWatcher
	for {
		select {
		case ev := <-watcher.Event:
			logger.Debugf("fsWatcher event: %v", ev)
			m.delayHandleFileEvent(ev.Name)
		case err := <-watcher.Error:
			logger.Warning("fsWatcher error", err)
		}
	}
}

func (m *Manager) delayHandleFileEvent(name string) {
	m.fsEventTimersMutex.Lock()
	defer m.fsEventTimersMutex.Unlock()

	delay := 2000 * time.Millisecond
	timer, ok := m.fsEventTimers[name]
	if ok {
		timer.Stop()
		timer.Reset(delay)
	} else {
		m.fsEventTimers[name] = time.AfterFunc(delay, func() {
			switch {
			case name == desktopPkgMapFile:
				if err := m.loadDesktopPkgMap(); err != nil {
					logger.Warning(err)
					return
				}
				// retry queryPkgName for m.noPkgItemIDs
				logger.Debugf("m.noPkgItemIDs: %v", m.noPkgItemIDs)
				for id := range m.noPkgItemIDs {
					pkg := m.queryPkgName(id)
					logger.Debugf("item id %q pkg %q", id, pkg)

					if pkg != "" {
						item := m.getItemById(id)
						cid := m._queryCategoryID(item, pkg)

						if cid != item.CategoryID {
							// item.CategoryID changed
							item.CategoryID = cid
							m.emitItemChanged(item, AppStatusModified)
						}
						delete(m.noPkgItemIDs, id)
					}
				}

			case name == applicationsFile:
				if err := m.loadPkgCategoryMap(); err != nil {
					logger.Warning(err)
				}

			case isDesktopFile(name):
				m.checkDesktopFile(name)
			}
		})
	}
}

func (m *Manager) checkDesktopFile(file string) {
	logger.Debug("checkDesktopFile", file)
	appId := m.getAppIdByFilePath(file)
	logger.Debugf("app id %q", appId)
	if appId == "" {
		logger.Warningf("appId is empty, ignore file %q", file)
		return
	}

	item := m.getItemById(appId)

	appInfo := desktopappinfo.NewDesktopAppInfo(appId)
	if appInfo == nil {
		logger.Warningf("appId %q appInfo is nil", appId)
		if item != nil {
			m.removeItem(appId)
			m.emitItemChanged(item, AppStatusDeleted)

			// remove desktop file in user's desktop direcotry
			os.Remove(appInDesktop(appId))
		}
	} else {
		// appInfo is not nil
		shouldShow := appInfo.ShouldShow() &&
			!isDeepinCustomDesktopFile(appInfo.GetFileName())
		newItem := NewItemWithDesktopAppInfo(appInfo)

		// add or update item
		if item != nil {

			if shouldShow {
				// update item
				m.addItemWithLock(newItem)
				m.emitItemChanged(newItem, AppStatusModified)
			} else {
				m.removeItem(appId)
				m.emitItemChanged(newItem, AppStatusDeleted)
			}
		} else {
			if shouldShow {
				if appInfo.IsExecutableOk() {
					m.addItemWithLock(newItem)
					m.emitItemChanged(newItem, AppStatusCreated)
				} else {
					go m.retryAddItem(appInfo, newItem)
				}
			}
		}
	}
}

func (m *Manager) retryAddItem(appInfo *desktopappinfo.DesktopAppInfo, item *Item) {
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)
		logger.Debug("retry add item", item.Path)
		if appInfo.IsExecutableOk() {
			m.addItemWithLock(item)
			m.emitItemChanged(item, AppStatusCreated)
			return
		}
	}
}

func (m *Manager) emitItemChanged(item *Item, status string) {
	m.itemChanged = true
	itemInfo := item.newItemInfo()
	logger.Debugf("emit signal ItemChanged status: %v, itemInfo: %v", status, itemInfo)
	dbus.Emit(m, "ItemChanged", status, itemInfo, itemInfo.CategoryID)
}
