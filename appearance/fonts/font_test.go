/**
 * Copyright (C) 2014 Deepin Technology Co., Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 **/

package fonts

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCompositeList(t *testing.T) {
	Convey("Test list composition", t, func() {
		So(compositeList([]string{"123", "234"}, []string{"234", "abc"}),
			ShouldResemble, []string{"123", "234", "abc"})
	})
}

func TestFontFamily(t *testing.T) {
	Convey("Test font family", t, func() {
		var families = Families{
			&Family{
				Id:     "Source Code Pro",
				Name:   "Source Code Pro",
				Styles: []string{"Regular", "Bold"},
			},
			&Family{
				Id:     "WenQuanYi Micro Hei",
				Name:   "文泉译微米黑",
				Styles: []string{"Normal", "Bold"},
			},
		}
		So(families.GetIds(), ShouldResemble,
			[]string{"Source Code Pro",
				"WenQuanYi Micro Hei"})
		So(families.Get("WenQuanYi Micro Hei").Id, ShouldEqual,
			"WenQuanYi Micro Hei")
	})
}

func TestGetLangFromLocale(t *testing.T) {
	Convey("Test get lang from locale", t, func() {
		So(getLangFromLocale("zh_CN"), ShouldEqual, "zh-cn")
		So(getLangFromLocale("pap_AW"), ShouldEqual, "pap-aw")
		So(getLangFromLocale("zh_HK"), ShouldEqual, "zh-tw")
		So(getLangFromLocale("en_US"), ShouldEqual, "en")
	})
}
