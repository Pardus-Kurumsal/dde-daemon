/**
 * Copyright (C) 2014 Deepin Technology Co., Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 **/

package systeminfo

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func TestCPUInfo(t *testing.T) {
	Convey("Test cpu info", t, func() {
		cpu, err := GetCPUInfo("testdata/cpuinfo")
		So(cpu, ShouldEqual,
			"Intel(R) Core(TM) i3 CPU M 330 @ 2.13GHz x 4")
		So(err, ShouldBeNil)

		cpu, err = GetCPUInfo("testdata/sw-cpuinfo")
		So(cpu, ShouldEqual, "sw 1.40GHz x 4")
		So(err, ShouldBeNil)

		cpu, err = GetCPUInfo("testdata/loonson3-cpuinfo")
		So(cpu, ShouldEqual, "ICT Loongson-3B V0.7 FPU V0.1 x 6")
		So(err, ShouldBeNil)

		cpu, err = GetCPUInfo("testdata/arm-cpuinfo")
		So(cpu, ShouldEqual, "ARMv7 Processor rev 0 (v7l) x 4")
		So(err, ShouldBeNil)
	})
}

func TestMemInfo(t *testing.T) {
	Convey("Test memory info", t, func() {
		mem, err := getMemoryFromFile("testdata/meminfo")
		So(mem, ShouldEqual, uint64(4005441536))
		So(err, ShouldBeNil)
	})
}

func TestVersion(t *testing.T) {
	Convey("Test os version", t, func() {
		lang := os.Getenv("LANGUAGE")
		os.Setenv("LANGUAGE", "en_US")
		defer os.Setenv("LANGUAGE", lang)

		deepin, err := getVersionFromDeepin("testdata/deepin-version")
		So(deepin, ShouldEqual, "2015 Desktop Alpha1")
		So(err, ShouldBeNil)

		lsb, err := getVersionFromLSB("testdata/lsb-release")
		So(lsb, ShouldEqual, "2014.3")
		So(err, ShouldBeNil)
	})
}

func TestDistro(t *testing.T) {
	Convey("Test os distro", t, func() {
		lang := os.Getenv("LANGUAGE")
		os.Setenv("LANGUAGE", "en_US")
		defer os.Setenv("LANGUAGE", lang)

		distroId, distroDesc, distroVer, err := getDistroFromLSB("testdata/lsb-release")
		So(distroId, ShouldEqual, "Deepin")
		So(distroDesc, ShouldEqual, "Deepin 2014.3")
		So(distroVer, ShouldEqual, "2014.3")
		So(err, ShouldBeNil)
	})
}

func TestSystemBit(t *testing.T) {
	Convey("Test getconf", t, func() {
		v := systemBit()
		if v != "32" {
			So(v, ShouldEqual, "64")
		}

		if v != "64" {
			So(v, ShouldEqual, "32")
		}
	})
}
