// Copyright 2016 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"android/soong/android"
)

var (
	darwinCflags = []string{
		"-fdiagnostics-color",

		"-fPIC",
		"-funwind-tables",

		// Workaround differences in inttypes.h between host and target.
		//See bug 12708004.
		"-D__STDC_FORMAT_MACROS",
		"-D__STDC_CONSTANT_MACROS",

		"-isysroot ${macSdkRoot}",
		"-mmacosx-version-min=${macMinVersion}",
		"-DMACOSX_DEPLOYMENT_TARGET=${macMinVersion}",

		"-m64",
	}

	darwinLdflags = []string{
		"-isysroot ${macSdkRoot}",
		"-Wl,-syslibroot,${macSdkRoot}",
		"-mmacosx-version-min=${macMinVersion}",
		"-m64",
	}

	darwinClangCflags = append(ClangFilterUnknownCflags(darwinCflags), []string{
		"-integrated-as",
		"-fstack-protector-strong",
	}...)

	darwinClangLdflags = ClangFilterUnknownCflags(darwinLdflags)

	darwinClangLldflags = ClangFilterUnknownLldflags(darwinClangLdflags)

	darwinSupportedSdkVersions = []string{
		"11",
		"12",
	}

	darwinAvailableLibraries = append(
		addPrefix([]string{
			"c",
			"dl",
			"m",
			"ncurses",
			"objc",
			"pthread",
		}, "-l"),
		"-framework AppKit",
		"-framework CoreFoundation",
		"-framework Foundation",
		"-framework IOKit",
		"-framework Security",
		"-framework SystemConfiguration",
	)
)

const (
	darwinGccVersion = "4.2.1"
)

func init() {
	pctx.VariableFunc("macSdkRoot", func(ctx android.PackageVarContext) string {
		return getMacTools(ctx).sdkRoot
	})
	pctx.StaticVariable("macMinVersion", "10.10")
	pctx.VariableFunc("MacArPath", func(ctx android.PackageVarContext) string {
		return getMacTools(ctx).arPath
	})

	pctx.VariableFunc("MacLipoPath", func(ctx android.PackageVarContext) string {
		return getMacTools(ctx).lipoPath
	})

	pctx.VariableFunc("MacStripPath", func(ctx android.PackageVarContext) string {
		return getMacTools(ctx).stripPath
	})

	pctx.VariableFunc("MacToolPath", func(ctx android.PackageVarContext) string {
		return getMacTools(ctx).toolPath
	})

	pctx.StaticVariable("DarwinGccVersion", darwinGccVersion)
	pctx.SourcePathVariable("DarwinGccRoot",
		"prebuilts/gcc/${HostPrebuiltTag}/host/i686-apple-darwin-${DarwinGccVersion}")

	pctx.StaticVariable("DarwinGccTriple", "i686-apple-darwin11")

	pctx.StaticVariable("DarwinClangCflags", strings.Join(darwinClangCflags, " "))
	pctx.StaticVariable("DarwinClangLdflags", strings.Join(darwinClangLdflags, " "))
	pctx.StaticVariable("DarwinClangLldflags", strings.Join(darwinClangLldflags, " "))

	pctx.StaticVariable("DarwinYasmFlags", "-f macho -m amd64")
}

type macPlatformTools struct {
	once sync.Once
	err  error

	sdkRoot   string
	arPath    string
	lipoPath  string
	stripPath string
	toolPath  string
}

var macTools = &macPlatformTools{}

func getMacTools(ctx android.PackageVarContext) *macPlatformTools {
	macTools.once.Do(func() {
		xcrunTool := "/usr/bin/xcrun"

		xcrun := func(args ...string) string {
			if macTools.err != nil {
				return ""
			}

			bytes, err := exec.Command(xcrunTool, append([]string{"--sdk", "macosx"}, args...)...).Output()
			if err != nil {
				macTools.err = fmt.Errorf("xcrun %q failed with: %q", args, err)
				return ""
			}

			return strings.TrimSpace(string(bytes))
		}

		sdkVersion := xcrun("--show-sdk-version")
		sdkVersionSupported := false
		for _, version := range darwinSupportedSdkVersions {
			if version == sdkVersion || strings.HasPrefix(sdkVersion, version+".") {
				sdkVersionSupported = true
			}
		}
		if !sdkVersionSupported {
			macTools.err = fmt.Errorf("Unsupported macOS SDK version %q not in %v", sdkVersion, darwinSupportedSdkVersions)
			return
		}

		macTools.sdkRoot = xcrun("--show-sdk-path")

		macTools.arPath = xcrun("--find", "ar")
		macTools.lipoPath = xcrun("--find", "lipo")
		macTools.stripPath = xcrun("--find", "strip")
		macTools.toolPath = filepath.Dir(xcrun("--find", "ld"))
	})
	if macTools.err != nil {
		ctx.Errorf("%q", macTools.err)
	}
	return macTools
}

type toolchainDarwin struct {
	cFlags, ldFlags string
	toolchain64Bit
}

type toolchainDarwinX86 struct {
	toolchainDarwin
}

type toolchainDarwinArm struct {
	toolchainDarwin
}

func (t *toolchainDarwinArm) Name() string {
	return "arm64"
}

func (t *toolchainDarwinX86) Name() string {
	return "x86_64"
}

func (t *toolchainDarwinArm) GccRoot() string {
	panic("unimplemented")
}

func (t *toolchainDarwinArm) GccTriple() string {
	panic("unimplemented")
}

func (t *toolchainDarwinArm) GccVersion() string {
	panic("unimplemented")
}

func (t *toolchainDarwinX86) GccRoot() string {
	return "${config.DarwinGccRoot}"
}

func (t *toolchainDarwinX86) GccTriple() string {
	return "${config.DarwinGccTriple}"
}

func (t *toolchainDarwinX86) GccVersion() string {
	return darwinGccVersion
}

func (t *toolchainDarwin) IncludeFlags() string {
	return ""
}

func (t *toolchainDarwinArm) ClangTriple() string {
	return "aarch64-apple-darwin"
}

func (t *toolchainDarwinX86) ClangTriple() string {
	return "x86_64-apple-darwin"
}

func (t *toolchainDarwin) ClangCflags() string {
	return "${config.DarwinClangCflags}"
}

func (t *toolchainDarwin) ClangCppflags() string {
	return ""
}

func (t *toolchainDarwin) ClangLdflags() string {
	return "${config.DarwinClangLdflags}"
}

func (t *toolchainDarwin) ClangLldflags() string {
	return "${config.DarwinClangLldflags}"
}

func (t *toolchainDarwin) YasmFlags() string {
	return "${config.DarwinYasmFlags}"
}

func (t *toolchainDarwin) ShlibSuffix() string {
	return ".dylib"
}

func (t *toolchainDarwin) AvailableLibraries() []string {
	return darwinAvailableLibraries
}

func (t *toolchainDarwin) Bionic() bool {
	return false
}

func (t *toolchainDarwin) ToolPath() string {
	return "${config.MacToolPath}"
}

var toolchainDarwinArmSingleton Toolchain = &toolchainDarwinArm{}
var toolchainDarwinX86Singleton Toolchain = &toolchainDarwinX86{}

func darwinArmToolchainFactory(arch android.Arch) Toolchain {
	return toolchainDarwinArmSingleton
}

func darwinX86ToolchainFactory(arch android.Arch) Toolchain {
	return toolchainDarwinX86Singleton
}

func init() {
	registerToolchainFactory(android.Darwin, android.Arm64, darwinArmToolchainFactory)
	registerToolchainFactory(android.Darwin, android.X86_64, darwinX86ToolchainFactory)
}
