package template

import (
	"os"
	"path/filepath"
)

const CppTemplate = `#include <iostream>

using namespace std;

/* -- libraries --*/


void solve() {

}

int main() {
	cin.tie(0);
	ios::sync_with_stdio(false);
	solve();
	return 0;
}
`

const LibraryTemplate = `#pragma once

/* -- library code --*/


`

func CreateTemplate(rootDir string) error {
	os.MkdirAll(filepath.Join(rootDir, ".cpx", "templates", "source"), 0o755)
	mainPath := filepath.Join(rootDir, ".cpx", "templates", "source", "source_template.cpp")
	_, err := os.Stat(mainPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(mainPath, []byte(CppTemplate), 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}

func CreateLibraryTemplate(rootDir string) error {
	os.MkdirAll(filepath.Join(rootDir, ".cpx", "templates", "library"), 0o755)
	libraryPath := filepath.Join(rootDir, ".cpx", "templates", "library", "library_template.hpp")
	_, err := os.Stat(libraryPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(libraryPath, []byte(LibraryTemplate), 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}

func GetSourceCode(lang string, rootDir string) (string, error) {
	path := filepath.Join(rootDir, ".cpx", "templates", "source", "source_template."+lang)
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", err
		}
		return "", err
	}
	file, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(file), nil
}
