/**
 * Copyright (C) 2014 Deepin Technology Co., Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 **/

package background

import (
	"fmt"
	"os"
	"path"

	"gir/gio-2.0"
	"pkg.deepin.io/dde/api/thumbnails/images"
	"pkg.deepin.io/lib/graphic"
	"pkg.deepin.io/lib/strv"
	dutils "pkg.deepin.io/lib/utils"
	"pkg.deepin.io/lib/xdg/userdir"
	"sync"
)

const (
	thumbWidth  int = 128
	thumbHeight     = 72

	wrapBgSchema    = "com.deepin.wrap.gnome.desktop.background"
	gsKeyBackground = "picture-uri"
)

type Background struct {
	Id string

	Deletable bool
}
type Backgrounds []*Background

var (
	cacheBackgrounds Backgrounds

	locker  sync.Mutex
	setting *gio.Settings
)

func RefreshBackground() {
	locker.Lock()
	defer locker.Unlock()
	var infos Backgrounds
	for _, file := range getBgFiles() {
		infos = append(infos, &Background{
			Id:        dutils.EncodeURI(file, dutils.SCHEME_FILE),
			Deletable: isDeletable(file),
		})
	}
	cacheBackgrounds = infos
}

func ListBackground() Backgrounds {
	if len(cacheBackgrounds) == 0 {
		RefreshBackground()
	}
	return cacheBackgrounds
}

var supportedFormats = strv.Strv([]string{"jpeg", "png", "bmp", "tiff"})

func IsBackgroundFile(file string) bool {
	file = dutils.DecodeURI(file)
	format, err := graphic.SniffImageFormat(file)
	if err != nil {
		return false
	}

	if supportedFormats.Contains(format) {
		return true
	}
	return false
}

func (infos Backgrounds) EnsureExists(uri string) (string, error) {
	uri = dutils.EncodeURI(uri, dutils.SCHEME_FILE)
	if isFileInSpecialDir(uri, ListDirs()) {
		return uri, nil
	}

	file := dutils.DecodeURI(uri)
	dest, err := getBgDest(file)
	if err != nil {
		return "", err
	}

	if !dutils.IsFileExist(dest) {
		err = os.MkdirAll(path.Dir(dest), 0755)
		if err != nil {
			return "", err
		}

		err = dutils.CopyFile(file, dest)
		if err != nil {
			return "", err
		}
		RefreshBackground()
	}
	uri = dutils.EncodeURI(dest, dutils.SCHEME_FILE)

	return uri, nil
}

func (infos Backgrounds) GetIds() []string {
	var ids []string
	for _, info := range infos {
		ids = append(ids, info.Id)
	}
	return ids
}

func (infos Backgrounds) Get(uri string) *Background {
	// Ensure list not changed
	locker.Lock()
	defer locker.Unlock()
	uri = dutils.EncodeURI(uri, dutils.SCHEME_FILE)
	for _, info := range infos {
		if uri == info.Id {
			return info
		}
	}
	return nil
}

func (infos Backgrounds) Delete(uri string) error {
	info := infos.Get(uri)
	if info == nil {
		return fmt.Errorf("Not found '%s'", uri)
	}

	return info.Delete()
}

func (infos Backgrounds) Thumbnail(uri string) (string, error) {
	info := infos.Get(uri)
	if info == nil {
		return "", fmt.Errorf("Not found '%s'", uri)
	}

	return info.Thumbnail()
}

func (info *Background) Delete() error {
	if !info.Deletable {
		return fmt.Errorf("Permission Denied")
	}

	return os.Remove(dutils.DecodeURI(info.Id))
}

func (info *Background) Thumbnail() (string, error) {
	return images.ThumbnailForTheme(info.Id, thumbWidth, thumbHeight, false)
}

func getBgDest(file string) (string, error) {
	id, ok := dutils.SumFileMd5(file)
	if !ok {
		return "", fmt.Errorf("Not found '%s'", file)
	}
	return path.Join(
		getUserPictureDir(),
		"Wallpapers", id+path.Ext(file)), nil
}

func getUserPictureDir() string {
	dir := userdir.Get(userdir.Pictures)
	// Ensure dir exists
	os.MkdirAll(dir, 0755)
	return dir
}
