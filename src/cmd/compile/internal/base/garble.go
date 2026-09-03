// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import (
	"cmd/internal/objabi"
	"strings"
)

// TranslateGarblePkgPath returns the original package path for an obfuscated path.
func TranslateGarblePkgPath(pkgPath string) string {
	return objabi.OriginalPackagePath(pkgPath)
}

// IsGarblePkgPath reports whether pkgPath was explicitly mapped from an
// obfuscated package path.
func IsGarblePkgPath(pkgPath string) bool {
	return objabi.IsPackagePathMapped(pkgPath)
}

// GarbleRuntimeSymbol returns the package path and declaration name used by
// garbled object files for an original toolchain-generated runtime reference.
func GarbleRuntimeSymbol(pkgPath, name string) (string, string) {
	if obfuscated := objabi.ObfuscatedSymbol(pkgPath + "." + name); obfuscated != "" {
		if i := strings.LastIndexByte(obfuscated, '.'); i >= 0 {
			return obfuscated[:i], obfuscated[i+1:]
		}
	}
	return objabi.ObfuscatedPackagePath(pkgPath), name
}
