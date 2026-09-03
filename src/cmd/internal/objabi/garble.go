// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garble name mappings are shared by the compiler, assembler, and linker.
package objabi

import (
	"os"
	"strings"
	"sync"
)

const (
	pkgPathMapEnv = "GARBLE_PKGPATH_MAP"
	symbolMapEnv  = "GARBLE_SYMBOL_MAP"
)

type nameMap struct {
	forward map[string]string
	reverse map[string]string
}

func parseNameMap(value string) nameMap {
	m := nameMap{
		forward: make(map[string]string),
		reverse: make(map[string]string),
	}
	for _, entry := range strings.Split(value, ",") {
		obfuscated, original, ok := strings.Cut(entry, "=")
		if !ok || obfuscated == "" || original == "" {
			continue
		}
		m.forward[obfuscated] = original
		m.reverse[original] = obfuscated
	}
	return m
}

var pkgPaths = sync.OnceValue(func() nameMap {
	return parseNameMap(os.Getenv(pkgPathMapEnv))
})

var symbols = sync.OnceValue(func() nameMap {
	return parseNameMap(os.Getenv(symbolMapEnv))
})

// OriginalPackagePath returns the original package path for an obfuscated path.
func OriginalPackagePath(path string) string {
	if original := pkgPaths().forward[path]; original != "" {
		return original
	}
	return path
}

// IsPackagePathMapped reports whether path is an obfuscated package path.
func IsPackagePathMapped(path string) bool {
	_, ok := pkgPaths().forward[path]
	return ok
}

// ObfuscatedPackagePath returns the obfuscated package path for an original
// path, or path when no mapping exists.
func ObfuscatedPackagePath(path string) string {
	if obfuscated := pkgPaths().reverse[path]; obfuscated != "" {
		return obfuscated
	}
	return path
}

// OriginalSymbol returns the original symbol for an obfuscated symbol.
func OriginalSymbol(name string) string {
	if original := symbols().forward[name]; original != "" {
		return original
	}
	return name
}

// OriginalSymbolParts translates an explicitly mapped package path and
// declaration name. It does not apply the package-path fallback.
func OriginalSymbolParts(path, name string) (string, string) {
	original := OriginalSymbol(path + "." + name)
	if original == path+"."+name {
		return path, name
	}
	if dot := strings.LastIndexByte(original, '.'); dot >= 0 {
		return original[:dot], original[dot+1:]
	}
	return path, name
}

// OriginalPackageSymbol translates an obfuscated package path and declaration
// name to their original values.
func OriginalPackageSymbol(path, name string) (string, string) {
	originalPath, originalName := OriginalSymbolParts(path, name)
	if originalPath != path || originalName != name {
		return originalPath, originalName
	}
	return OriginalPackagePath(path), name
}

// ObfuscatedSymbol returns the obfuscated symbol for an original symbol, or an
// empty string when no mapping exists.
func ObfuscatedSymbol(name string) string {
	symbolMap := symbols().reverse
	pathMap := pkgPaths().reverse
	if len(symbolMap) == 0 && len(pathMap) == 0 {
		return ""
	}
	if obfuscated := symbolMap[name]; obfuscated != "" {
		return obfuscated
	}
	// Symbol names separate the package path from the declaration with a dot.
	// Walk those separators from right to left and use map lookups rather than
	// scanning every mapped package path on each linker lookup miss.
	for dot := strings.LastIndexByte(name, '.'); dot >= 0; dot = strings.LastIndexByte(name[:dot], '.') {
		if obfuscated := pathMap[name[:dot]]; obfuscated != "" {
			return obfuscated + name[dot:]
		}
	}
	return ""
}

// OriginalFuncName returns the original fully qualified function name. The
// symbol map takes precedence because it can rename both the package and the
// declaration. The package map handles names absent from the symbol map.
func OriginalFuncName(name string) string {
	if original := symbols().forward[name]; original != "" {
		return original
	}
	for obfuscated, original := range pkgPaths().forward {
		if strings.HasPrefix(name, obfuscated+".") {
			return original + name[len(obfuscated):]
		}
	}
	return name
}
