/**
 * Copyright (C) 2013 Deepin Technology Co., Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 **/

package accounts

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"pkg.deepin.io/lib/encoding/kv"
	"pkg.deepin.io/lib/graphic"
	"pkg.deepin.io/lib/procfs"
	"pkg.deepin.io/lib/utils"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	polkitManagerUser    = "com.deepin.daemon.accounts.user-administration"
	polkitChangeOwnData  = "com.deepin.daemon.accounts.change-own-user-data"
	polkitSetLoginOption = "com.deepin.daemon.accounts.set-login-option"
)

type ErrCodeType int32

const (
	// 未知错误
	ErrCodeUnkown ErrCodeType = iota
	// 权限认证失败
	ErrCodeAuthFailed
	// 执行命令失败
	ErrCodeExecFailed
	// 传入的参数不合法
	ErrCodeParamInvalid
)

func (code ErrCodeType) String() string {
	switch code {
	case ErrCodeUnkown:
		return "Unkown error"
	case ErrCodeAuthFailed:
		return "Policykit authentication failed"
	case ErrCodeExecFailed:
		return "Exec command failed"
	case ErrCodeParamInvalid:
		return "Invalid parameters"
	}

	return "Unkown error"
}

// return icons uris
func getUserStandardIcons() []string {
	imgs, err := graphic.GetImagesInDir(userIconsDir)
	if err != nil {
		return nil
	}

	var icons []string
	for _, img := range imgs {
		img = utils.EncodeURI(img, utils.SCHEME_FILE)
		if strings.Contains(img, "guest") || img == defaultUserIcon {
			continue
		}

		icons = append(icons, img)
	}

	return icons
}

func getNewUserCustomIconDest(username string) string {
	ns := time.Now().UnixNano()
	base := username + "-" + strconv.FormatInt(ns, 36)
	return filepath.Join(userCustomIconsDir, base)
}

func isUserCustomIconURI(iconURI string, username string) bool {
	iconFile := utils.DecodeURI(iconURI)
	iconFilePrefix := filepath.Join(userCustomIconsDir, username+"-")
	if strings.HasPrefix(iconFile, iconFilePrefix) &&
		len(iconFile) > len(iconFilePrefix) {
		return true
	}
	return false
}

func isStrInArray(str string, array []string) bool {
	for _, v := range array {
		if v == str {
			return true
		}
	}

	return false
}

func isStrvEqual(l1, l2 []string) bool {
	if len(l1) != len(l2) {
		return false
	}

	sort.Strings(l1)
	sort.Strings(l2)
	for i, v := range l1 {
		if v != l2[i] {
			return false
		}
	}
	return true
}

func polkitAuthManagerUser(pid uint32) error {
	return polkitAuthentication(polkitManagerUser, pid)
}

func polkitAuthChangeOwnData(pid uint32) error {
	return polkitAuthentication(polkitChangeOwnData, pid)
}

func polkitAuthentication(action string, pid uint32) error {
	success, err := polkitAuthWithPid(action, pid)
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf(ErrCodeAuthFailed.String())
	}

	return nil
}

func getUidByPid(pid uint32) (string, error) {
	process := procfs.Process(pid)
	status, err := process.Status()
	if err != nil {
		return "", err
	}

	uids, err := status.Uids()
	if err != nil {
		return "", err
	}

	//effective user id
	euid := strconv.FormatUint(uint64(uids[1]), 10)
	return euid, nil
}

func getLocaleFromFile(file string) string {
	f, err := os.Open(file)
	if err != nil {
		return ""
	}
	defer f.Close()

	r := kv.NewReader(f)
	r.Delim = '='
	r.Comment = '#'
	r.TrimSpace = kv.TrimLeadingTailingSpace
	for {
		pair, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return ""
		}

		if pair.Key == "LANG" {
			return pair.Value
		}
	}
	return ""
}

// Get available shells from '/etc/shells'
func getAvailableShells(file string) []string {
	contents, err := ioutil.ReadFile(file)
	if err != nil || len(contents) == 0 {
		return nil
	}
	var shells []string
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		if line == "" || line[0] == '#' {
			continue
		}
		shells = append(shells, line)
	}
	return shells
}
