/**
 * Copyright (C) 2014 Deepin Technology Co., Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 **/

package subthemes

import (
	"dbus/com/deepin/api/cursorhelper"
	"fmt"
	"gir/gio-2.0"
	"os"
	"path"
	"pkg.deepin.io/dde/api/themes"
	"pkg.deepin.io/dde/api/thumbnails/cursor"
	"pkg.deepin.io/dde/api/thumbnails/gtk"
	"pkg.deepin.io/dde/api/thumbnails/icon"
	dutils "pkg.deepin.io/lib/utils"
	"strings"
)

const (
	thumbWidth  = 320
	thumbHeight = 70

	thumbDir   = "/usr/share/personalization/thumbnail"
	thumbBgDir = "/var/cache/appearance/thumbnail/background"

	appearanceSchema  = "com.deepin.dde.appearance"
	gsKeyExcludedIcon = "excluded-icon-themes"
)

type Theme struct {
	Id   string
	Path string

	Deletable bool
}
type Themes []*Theme

var (
	cacheGtkThemes    Themes
	cacheIconThemes   Themes
	cacheCursorThemes Themes

	home = os.Getenv("HOME")
)

func RefreshGtkThemes() {
	cacheGtkThemes = getThemes(themes.ListGtkTheme())
}

func RefreshIconThemes() {
	infos := getThemes(themes.ListIconTheme())
	s := gio.NewSettings(appearanceSchema)
	defer s.Unref()
	blacklist := s.GetStrv(gsKeyExcludedIcon)

	var ret Themes
	for _, info := range infos {
		if isItemInList(info.Id, blacklist) {
			continue
		}
		ret = append(ret, info)
	}
	cacheIconThemes = ret
}

func RefreshCursorThemes() {
	cacheCursorThemes = getThemes(themes.ListCursorTheme())
}

func ListGtkTheme() Themes {
	if len(cacheGtkThemes) == 0 {
		RefreshGtkThemes()
	}
	return cacheGtkThemes
}

func ListIconTheme() Themes {
	if len(cacheIconThemes) == 0 {
		RefreshIconThemes()
	}
	return cacheIconThemes
}

func ListCursorTheme() Themes {
	if len(cacheCursorThemes) == 0 {
		RefreshCursorThemes()
	}
	return cacheCursorThemes
}

func IsGtkTheme(id string) bool {
	return themes.IsThemeInList(id, themes.ListGtkTheme())
}

func IsIconTheme(id string) bool {
	return themes.IsThemeInList(id, themes.ListIconTheme())
}

func IsCursorTheme(id string) bool {
	return themes.IsThemeInList(id, themes.ListCursorTheme())
}

func SetGtkTheme(id string) error {
	return themes.SetGtkTheme(id)
}

func SetIconTheme(id string) error {
	return themes.SetIconTheme(id)
}

func SetCursorTheme(id string) error {
	helper, err := cursorhelper.NewCursorHelper("com.deepin.api.CursorHelper",
		"/com/deepin/api/CursorHelper")
	if err != nil {
		return err
	}
	helper.Set(id)
	return nil
}

func GetGtkThumbnail(id string) (string, error) {
	info := ListGtkTheme().Get(id)
	if info == nil {
		return "", fmt.Errorf("Not found '%s'", id)
	}

	var thumb = path.Join(thumbDir, "WindowThemes", id+"-thumbnail.png")
	if dutils.IsFileExist(thumb) {
		return thumb, nil
	}

	return gtk.ThumbnailForTheme(path.Join(info.Path, "index.theme"), "",
		thumbWidth, thumbHeight, false)
}

func GetIconThumbnail(id string) (string, error) {
	info := ListIconTheme().Get(id)
	if info == nil {
		return "", fmt.Errorf("Not found '%s'", id)
	}

	var thumb = path.Join(thumbDir, "IconThemes", id+"-thumbnail.png")
	if dutils.IsFileExist(thumb) {
		return thumb, nil
	}
	return icon.ThumbnailForTheme(path.Join(info.Path, "index.theme"), "",
		thumbWidth, thumbHeight, false)
}

func GetCursorThumbnail(id string) (string, error) {
	info := ListCursorTheme().Get(id)
	if info == nil {
		return "", fmt.Errorf("Not found '%s'", id)
	}

	var thumb = path.Join(thumbDir, "CursorThemes", id+"-thumbnail.png")
	if dutils.IsFileExist(thumb) {
		return thumb, nil
	}
	return cursor.ThumbnailForTheme(path.Join(info.Path, "cursor.theme"), "",
		thumbWidth, thumbHeight, false)
}

func (infos Themes) GetIds() []string {
	var ids []string
	for _, info := range infos {
		ids = append(ids, info.Id)
	}
	return ids
}

func (infos Themes) Get(id string) *Theme {
	for _, info := range infos {
		if id == info.Id {
			return info
		}
	}
	return nil
}

func (infos Themes) Delete(id string) error {
	info := infos.Get(id)
	if info == nil {
		return fmt.Errorf("Not found '%s'", id)
	}
	return info.Delete()
}

func (info *Theme) Delete() error {
	if !info.Deletable {
		return fmt.Errorf("Permission Denied")
	}
	return os.RemoveAll(info.Path)
}

func getThemes(files []string) Themes {
	var infos Themes
	for _, v := range files {
		infos = append(infos, &Theme{
			Id:        path.Base(v),
			Path:      v,
			Deletable: isDeletable(v),
		})
	}
	return infos
}

func isDeletable(file string) bool {
	if strings.Contains(file, home) {
		return true
	}
	return false
}

func isItemInList(item string, list []string) bool {
	for _, v := range list {
		if item == v {
			return true
		}
	}
	return false
}
