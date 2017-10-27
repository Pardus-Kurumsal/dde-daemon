/**
 * Copyright (C) 2013 Deepin Technology Co., Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 **/

package users

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var (
	errInvalidParam = fmt.Errorf("Invalid or empty parameter")
)

var (
	groupFileTimestamp int64 = 0
	groupFileInfo            = make(map[string][]string)
	groupFileLocker    sync.Mutex
)

const CommentFieldsLen = 5

// CommentInfo is passwd file user comment info
type CommentInfo [CommentFieldsLen]string

func newCommentInfo(comment string) *CommentInfo {
	var ci CommentInfo
	parts := strings.Split(comment, ",")

	// length is min(CommentFieldsLen, len(parts))
	length := len(parts)
	if length > CommentFieldsLen {
		length = CommentFieldsLen
	}

	copy(ci[:], parts[:length])
	return &ci
}

func (ci *CommentInfo) String() string {
	return strings.Join(ci[:], ",")
}

func (ci *CommentInfo) FullName() string {
	return ci[0]
}

func (ci *CommentInfo) SetFullName(value string) {
	ci[0] = value
}

func isCommentFieldValid(name string) bool {
	if strings.ContainsAny(name, ",=:") {
		return false
	}
	return true
}

func ModifyFullName(newname, username string) error {
	if !isCommentFieldValid(newname) {
		return errors.New("invalid nickname")
	}

	user, err := GetUserInfoByName(username)
	if err != nil {
		return err
	}
	comment := user.Comment()
	comment.SetFullName(newname)
	return modifyComment(comment.String(), username)
}

func modifyComment(comment, username string) error {
	cmd := exec.Command(userCmdModify, "-c", comment, username)
	return cmd.Run()
}

func ModifyHome(dir, username string) error {
	if len(dir) == 0 {
		return errInvalidParam
	}

	return doAction(userCmdModify, []string{"-m", "-d", dir, username})
}

func ModifyShell(shell, username string) error {
	if len(shell) == 0 {
		return errInvalidParam
	}

	return doAction(userCmdModify, []string{"-s", shell, username})
}

func ModifyPasswd(words, username string) error {
	if len(words) == 0 {
		return errInvalidParam
	}

	return updatePasswd(EncodePasswd(words), username)
}

// passwd -S username
func IsUserLocked(username string) bool {
	output, err := exec.Command("passwd", []string{"-S", username}...).Output()
	if err != nil {
		return true
	}

	items := strings.Split(string(output), " ")
	if items[1] == "L" {
		return true
	}

	return false
}

func IsAutoLoginUser(username string) bool {
	name, _ := GetAutoLoginUser()
	if name == username {
		return true
	}

	return false
}

func IsAdminUser(username string) bool {
	admins, err := getAdminUserList(userFileGroup, userFileSudoers)
	if err != nil {
		return false
	}

	return isStrInArray(username, admins)
}

func getAdminUserList(fileGroup, fileSudoers string) ([]string, error) {
	groups, users, err := getAdmGroupAndUser(fileSudoers)
	if err != nil {
		return nil, err
	}

	groupFileLocker.Lock()
	defer groupFileLocker.Unlock()
	infos, err := parseGroupFile(fileGroup)
	if err != nil {
		return nil, err
	}

	for _, group := range groups {
		v, ok := infos[group]
		if !ok {
			continue
		}
		users = append(users, v...)
	}
	return users, nil
}

var (
	_admGroups       []string
	_admUsers        []string
	_admTimestampMap = make(map[string]int64)
)

// get adm group and user from '/etc/sudoers'
func getAdmGroupAndUser(file string) ([]string, []string, error) {
	finfo, err := os.Stat(file)
	if err != nil {
		return nil, nil, err
	}
	timestamp := finfo.ModTime().Unix()
	if t, ok := _admTimestampMap[file]; ok && t == timestamp {
		return _admGroups, _admUsers, nil
	}

	fr, err := os.Open(file)
	if err != nil {
		return nil, nil, err
	}
	defer fr.Close()

	var (
		groups  []string
		users   []string
		scanner = bufio.NewScanner(fr)
	)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		line = strings.TrimSpace(line)
		if line[0] == '#' || !strings.Contains(line, `ALL=(ALL`) {
			continue
		}

		array := strings.Split(line, "ALL")
		// admin group
		if line[0] == '%' {
			// deepin: %sudo\tALL=(ALL:ALL) ALL
			// archlinux: %wheel ALL=(ALL) ALL
			array = strings.Split(array[0], "%")
			tmp := strings.TrimRight(array[1], "\t")
			groups = append(groups, strings.TrimSpace(tmp))
		} else {
			// admin user
			// deepin: root\tALL=(ALL:ALL) ALL
			// archlinux: root ALL=(ALL) ALL
			tmp := strings.TrimRight(array[0], "\t")
			users = append(users, strings.TrimSpace(tmp))
		}
	}
	_admGroups, _admUsers = groups, users
	_admTimestampMap[file] = timestamp
	return groups, users, nil
}

func isGroupExists(group string) bool {
	groupFileLocker.Lock()
	defer groupFileLocker.Unlock()
	infos, err := parseGroupFile(userFileGroup)
	if err != nil {
		return false
	}
	_, ok := infos[group]
	return ok
}

func isUserInGroup(user, group string) bool {
	groupFileLocker.Lock()
	defer groupFileLocker.Unlock()
	infos, err := parseGroupFile(userFileGroup)
	if err != nil {
		return false
	}
	v, ok := infos[group]
	if !ok {
		return false
	}
	return isStrInArray(user, v)
}

func parseGroupFile(file string) (map[string][]string, error) {
	info, err := os.Stat(file)
	if err != nil {
		return nil, err
	}
	if groupFileTimestamp == info.ModTime().UnixNano() &&
		len(groupFileInfo) != 0 {
		return groupFileInfo, nil
	}

	groupFileTimestamp = info.ModTime().UnixNano()
	groupFileInfo = make(map[string][]string)
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		items := strings.Split(line, ":")
		if len(items) != itemLenGroup {
			continue
		}

		groupFileInfo[items[0]] = strings.Split(items[3], ",")
	}

	return groupFileInfo, nil
}
