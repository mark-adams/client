// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// This is a utility which binds to libkb to get the correct version
// for printing out or generating compiled resources for the windows
// executlable.

// +build windows

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/josephspurrier/goversioninfo"
	"github.com/keybase/client/go/libkb"
)

// Create the syso file
func main() {

	outPtr := flag.String("o", "rsrc_windows.syso", "resource output pathname")
	printverPtr := flag.Bool("v", false, "print version to console (no .syso output)")
	printWinVerPtr := flag.Bool("w", false, "print windows format version to console (no .syso output)")
	iconPtr := flag.String("i", "../../../keybase/public/images/favicon.ico", "icon pathname")

	flag.Parse()

	var fv goversioninfo.FileVersion

	if int, err := fmt.Sscanf(libkb.Version, "%d.%d.%d", &fv.Major, &fv.Minor, &fv.Patch); int != 3 || err != nil {
		log.Printf("Error parsing version %v", err)
		os.Exit(3)
	}
	if int, err := fmt.Sscanf(libkb.Build(), "%d", &fv.Build); int != 1 || err != nil {
		log.Printf("Error parsing build %v", err)
		os.Exit(3)
	}

	if *printverPtr {
		fmt.Printf("%d.%d.%d-%d", fv.Major, fv.Minor, fv.Patch, fv.Build)
		return
	}

	if *printWinVerPtr {
		fmt.Printf("%d.%d.%d.%d", fv.Major, fv.Minor, fv.Patch, fv.Build)
		return
	}

	// Create a new container
	vi := &goversioninfo.VersionInfo{
		FixedFileInfo: goversioninfo.FixedFileInfo{
			FileVersion:    fv,
			ProductVersion: fv,
			FileFlagsMask:  "3f",
			FileFlags:      "00",
			FileOS:         "040004",
			FileType:       "01",
			FileSubType:    "00",
		},
		StringFileInfo: goversioninfo.StringFileInfo{
			CompanyName:      "Keybase, Inc.",
			FileDescription:  "Keybase utility",
			InternalName:     "Keybase",
			LegalCopyright:   "Copyright (c) 2015, Keybase",
			OriginalFilename: "keybase.exe",
			ProductName:      "Keybase",
			ProductVersion:   libkb.VersionString(),
		},
		VarFileInfo: goversioninfo.VarFileInfo{
			Translation: goversioninfo.Translation{
				LangID:    0x409, // english
				CharsetID: 0x4B0, // unicode
			},
		},
	}

	// Fill the structures with config data
	vi.Build()

	// Write the data to a buffer
	vi.Walk()

	// Optionally, embed an icon by path
	// If the icon has multiple sizes, all of the sizes will be embedded
	vi.IconPath = *iconPtr

	// Create the file
	if err := vi.WriteSyso(*outPtr); err != nil {
		log.Printf("Error writing %s: %v", *outPtr, err)
		os.Exit(3)
	}
}
