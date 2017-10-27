/**
 * Copyright (C) 2013 Deepin Technology Co., Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 **/

package main

import (
	"fmt"
	"os"
	"os/user"
	"strings"
)

func helper() {
	fmt.Println("Initialize the user configuration, if the configuration files exist out directly.\n")
	fmt.Println("Usage: user-config [username]")
	fmt.Println("\tIf the user is not specified, will configure the current user.")
}

func getUsername(args []string) (string, bool, error) {
	if len(args) == 1 {
		u, err := user.Current()
		if err != nil {
			return "", false, err
		}
		return u.Username, false, nil
	}

	var arg = strings.ToLower(args[1])
	if arg == "-h" || arg == "--help" {
		return "", true, nil
	}

	return args[1], false, nil
}

func main() {
	name, help, err := getUsername(os.Args)
	if err != nil {
		fmt.Println("Parse arguments failed:", err)
		return
	}

	if help {
		helper()
		return
	}

	fmt.Printf("Start init '%s' configuration.\n", name)
	CopyUserDatas(name)
	fmt.Printf("Init '%s' configuration over.\n", name)
}
