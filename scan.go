package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"sourcegraph.com/sourcegraph/srclib/unit"
)

func init() {
	_, err := flagParser.AddCommand("scan",
		"scan for man pages",
		"Scan the directory tree rooted at the current directory for man pages.",
		&scanCmd,
	)
	if err != nil {
		log.Fatal(err)
	}
}

type ScanCmd struct{}

var scanCmd ScanCmd

func (c *ScanCmd) Execute(args []string) error {
	scanDir, err := filepath.EvalSymlinks(getCWD())
	if err != nil {
		return fmt.Errorf("resolving the path to scan failed with: %s", err)
	}

	units, err := scan(scanDir)
	if err != nil {
		return fmt.Errorf("scanning the path failed with: %s", err)
	}

	bytes, err := json.MarshalIndent(units, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling source units failed with: %s, units: %s", err, units)
	}

	if _, err := os.Stdout.Write(bytes); err != nil {
		return fmt.Errorf("writing output failed with: %s", err)
	}

	return nil
}

func scan(scanDir string) ([]*unit.SourceUnit, error) {
	var units []*unit.SourceUnit
	var files []string

	err := filepath.Walk(scanDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walking directory %s failed with: %s", scanDir, err)
		}
		if info.Mode().IsRegular() && strings.HasSuffix(path, ".1p.txt") {
			relpath, err := filepath.Rel(scanDir, path)
			if err != nil {
				return fmt.Errorf("making path %s relative to %s failed with: %s", path, scanDir, err)
			}
			files = append(files, relpath)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scanning for man pages failed with: %s", err)
	}

	units = append(units, &unit.SourceUnit{
		Key: unit.Key{
			Name: "man",
			Type: "ManPages",
		},
		Info: unit.Info{
			Files: files,
		},
	})

	return units, nil
}
