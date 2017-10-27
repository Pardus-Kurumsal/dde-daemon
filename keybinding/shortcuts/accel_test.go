/**
 * Copyright (C) 2016 Deepin Technology Co., Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 **/

package shortcuts

import (
	"github.com/BurntSushi/xgb/xproto"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSplitStandardAccel(t *testing.T) {
	Convey("splitStandardAccel", t, func() {
		var keys []string
		var err error
		keys, err = splitStandardAccel("<Super>L")
		So(err, ShouldBeNil)
		So(keys, ShouldResemble, []string{"Super", "L"})

		// single key
		keys, err = splitStandardAccel("<Super>")
		So(err, ShouldBeNil)
		So(keys, ShouldResemble, []string{"Super"})

		keys, err = splitStandardAccel("Super_L")
		So(err, ShouldBeNil)
		So(keys, ShouldResemble, []string{"Super_L"})

		keys, err = splitStandardAccel("<Shift><Super>T")
		So(err, ShouldBeNil)
		So(keys, ShouldResemble, []string{"Shift", "Super", "T"})

		// abnormal situation:
		keys, err = splitStandardAccel("<Super>>")
		So(err, ShouldNotBeNil)

		keys, err = splitStandardAccel("<Super><")
		So(err, ShouldNotBeNil)

		keys, err = splitStandardAccel("Super<")
		So(err, ShouldNotBeNil)

		keys, err = splitStandardAccel("<Super><shiftT")
		So(err, ShouldNotBeNil)

		keys, err = splitStandardAccel("<Super><Shift><>T")
		So(err, ShouldNotBeNil)
	})
}

func TestParseStandardAccel(t *testing.T) {
	Convey("ParseStandardAccel", t, func() {
		var parsed ParsedAccel
		var err error

		parsed, err = ParseStandardAccel("Super_L")
		So(err, ShouldBeNil)
		So(parsed, ShouldResemble, ParsedAccel{Key: "Super_L"})

		parsed, err = ParseStandardAccel("Num_Lock")
		So(err, ShouldBeNil)
		So(parsed, ShouldResemble, ParsedAccel{Key: "Num_Lock"})

		parsed, err = ParseStandardAccel("<Control><Super>T")
		So(err, ShouldBeNil)
		So(parsed, ShouldResemble, ParsedAccel{
			Key:  "T",
			Mods: xproto.ModMask4 | xproto.ModMaskControl,
		})

		parsed, err = ParseStandardAccel("<Control><Alt><Shift><Super>T")
		So(err, ShouldBeNil)
		So(parsed, ShouldResemble, ParsedAccel{
			Key:  "T",
			Mods: xproto.ModMaskShift | xproto.ModMask4 | xproto.ModMask1 | xproto.ModMaskControl,
		})

		parsed, err = ParseStandardAccel("<Shift>XXXXX")
		So(err, ShouldBeNil)
		So(parsed, ShouldResemble, ParsedAccel{Key: "XXXXX", Mods: xproto.ModMaskShift})

		// abnormal situation:
		parsed, err = ParseStandardAccel("")
		So(err, ShouldNotBeNil)

		parsed, err = ParseStandardAccel("<lock><Shift>A")
		So(err, ShouldNotBeNil)
	})
}

func TestParsedAccelMethodString(t *testing.T) {
	Convey("ParsedAccel.String", t, func() {
		var parsed ParsedAccel
		parsed = ParsedAccel{
			Key:  "percent",
			Mods: xproto.ModMaskControl | xproto.ModMaskShift,
		}
		So(parsed.String(), ShouldEqual, "<Shift><Control>percent")

		parsed = ParsedAccel{
			Key:  "T",
			Mods: xproto.ModMaskShift | xproto.ModMask4 | xproto.ModMask1 | xproto.ModMaskControl,
		}
		So(parsed.String(), ShouldEqual, "<Shift><Control><Alt><Super>T")
	})
}
