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
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func Test_getAppIdByFilePath(t *testing.T) {
	Convey("getAppIdByFilePath", t, func() {
		appDirs := []string{"/usr/share/applications", "/usr/local/share/applications", "/home/test_user/.local/share/applications"}

		id := getAppIdByFilePath("/usr/share/applications/d-feet.desktop", appDirs)
		So(id, ShouldEqual, "d-feet")

		id = getAppIdByFilePath("/usr/share/applications/kde4/krita.desktop", appDirs)
		So(id, ShouldEqual, "kde4/krita")

		id = getAppIdByFilePath("/usr/local/share/applications/deepin-screenshot.desktop", appDirs)
		So(id, ShouldEqual, "deepin-screenshot")

		id = getAppIdByFilePath("/home/test_user/.local/share/applications/space test.desktop", appDirs)
		So(id, ShouldEqual, "space test")

		id = getAppIdByFilePath("/other/dir/a.desktop", appDirs)
		So(id, ShouldEqual, "")
	})
}

func Test_getUserAppDir(t *testing.T) {
	Convey("getUserAppDir", t, func() {
		oldHome := os.Getenv("HOME")
		os.Setenv("HOME", "/home/test")
		So(getUserAppDir(), ShouldEqual, "/home/test/.local/share/applications")
		os.Setenv("HOME", oldHome)
	})
}

func Test_runeSliceDiff(t *testing.T) {
	Convey("runeSliceDiff", t, func() {
		// pop
		popCount, runesPush := runeSliceDiff([]rune("abc"), []rune("abc"))
		So(popCount, ShouldEqual, 0)
		So(len(runesPush), ShouldEqual, 0)

		popCount, runesPush = runeSliceDiff([]rune("abc"), []rune("abcd"))
		So(popCount, ShouldEqual, 1)
		So(len(runesPush), ShouldEqual, 0)

		popCount, runesPush = runeSliceDiff([]rune("abc"), []rune("abcde"))
		So(popCount, ShouldEqual, 2)
		So(len(runesPush), ShouldEqual, 0)

		// push
		popCount, runesPush = runeSliceDiff([]rune("abcd"), []rune("abc"))
		So(popCount, ShouldEqual, 0)
		So(len(runesPush), ShouldEqual, 1)
		So(runesPush[0], ShouldEqual, 'd')

		popCount, runesPush = runeSliceDiff([]rune("abcde"), []rune("abc"))
		So(popCount, ShouldEqual, 0)
		So(len(runesPush), ShouldEqual, 2)
		So(runesPush[0], ShouldEqual, 'd')
		So(runesPush[1], ShouldEqual, 'e')

		// pop and push
		popCount, runesPush = runeSliceDiff([]rune("abcd"), []rune("abce"))
		So(popCount, ShouldEqual, 1)
		So(len(runesPush), ShouldEqual, 1)
		So(runesPush[0], ShouldEqual, 'd')

		popCount, runesPush = runeSliceDiff([]rune("deepin"), []rune("deeinp"))
		So(popCount, ShouldEqual, 3)
		So(len(runesPush), ShouldEqual, 3)
		So(runesPush[0], ShouldEqual, 'p')
		So(runesPush[1], ShouldEqual, 'i')
		So(runesPush[2], ShouldEqual, 'n')
	})
}
