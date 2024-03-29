// Copyright 2021 Google Inc. All rights reserved.
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

package bp2build

import (
	"testing"

	"android/soong/android"
	"android/soong/cc"
	"android/soong/java"
)

func runJavaBinaryHostTestCase(t *testing.T, tc Bp2buildTestCase) {
	t.Helper()
	(&tc).ModuleTypeUnderTest = "java_binary_host"
	(&tc).ModuleTypeUnderTestFactory = java.BinaryHostFactory
	RunBp2BuildTestCase(t, func(ctx android.RegistrationContext) {
		ctx.RegisterModuleType("cc_library_host_shared", cc.LibraryHostSharedFactory)
		ctx.RegisterModuleType("java_library", java.LibraryFactory)
		ctx.RegisterModuleType("java_import_host", java.ImportFactory)
	}, tc)
}

var testFs = map[string]string{
	"test.mf": "Main-Class: com.android.test.MainClass",
	"other/Android.bp": `cc_library_host_shared {
    name: "jni-lib-1",
    stl: "none",
}`,
}

func TestJavaBinaryHost(t *testing.T) {
	runJavaBinaryHostTestCase(t, Bp2buildTestCase{
		Description: "java_binary_host with srcs, exclude_srcs, jni_libs, javacflags, and manifest.",
		Filesystem:  testFs,
		Blueprint: `java_binary_host {
    name: "java-binary-host-1",
    srcs: ["a.java", "b.java"],
    exclude_srcs: ["b.java"],
    manifest: "test.mf",
    jni_libs: ["jni-lib-1"],
    javacflags: ["-Xdoclint:all/protected"],
    bazel_module: { bp2build_available: true },
    java_version: "8",
}`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("java_binary", "java-binary-host-1", AttrNameToString{
				"srcs":       `["a.java"]`,
				"main_class": `"com.android.test.MainClass"`,
				"deps":       `["//other:jni-lib-1"]`,
				"jvm_flags":  `["-Djava.library.path=$${RUNPATH}other"]`,
				"javacopts": `[
        "-Xdoclint:all/protected",
        "-source 1.8 -target 1.8",
    ]`,
				"target_compatible_with": `select({
        "//build/bazel/platforms/os:android": ["@platforms//:incompatible"],
        "//conditions:default": [],
    })`,
			}),
		},
	})
}

func TestJavaBinaryHostRuntimeDeps(t *testing.T) {
	runJavaBinaryHostTestCase(t, Bp2buildTestCase{
		Description: "java_binary_host with srcs, exclude_srcs, jni_libs, javacflags, and manifest.",
		Filesystem:  testFs,
		Blueprint: `java_binary_host {
    name: "java-binary-host-1",
    static_libs: ["java-dep-1"],
    manifest: "test.mf",
    bazel_module: { bp2build_available: true },
}

java_library {
    name: "java-dep-1",
    srcs: ["a.java"],
    bazel_module: { bp2build_available: false },
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("java_binary", "java-binary-host-1", AttrNameToString{
				"main_class":   `"com.android.test.MainClass"`,
				"runtime_deps": `[":java-dep-1"]`,
				"target_compatible_with": `select({
        "//build/bazel/platforms/os:android": ["@platforms//:incompatible"],
        "//conditions:default": [],
    })`,
			}),
		},
	})
}

func TestJavaBinaryHostLibs(t *testing.T) {
	runJavaBinaryHostTestCase(t, Bp2buildTestCase{
		Description: "java_binary_host with srcs, libs.",
		Filesystem:  testFs,
		Blueprint: `java_binary_host {
    name: "java-binary-host-libs",
    libs: ["java-lib-dep-1"],
    manifest: "test.mf",
    srcs: ["a.java"],
}

java_import_host{
    name: "java-lib-dep-1",
    jars: ["foo.jar"],
    bazel_module: { bp2build_available: false },
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("java_binary", "java-binary-host-libs", AttrNameToString{
				"main_class": `"com.android.test.MainClass"`,
				"srcs":       `["a.java"]`,
				"deps":       `[":java-lib-dep-1-neverlink"]`,
				"target_compatible_with": `select({
        "//build/bazel/platforms/os:android": ["@platforms//:incompatible"],
        "//conditions:default": [],
    })`,
			}),
		},
	})
}

func TestJavaBinaryHostKotlinSrcs(t *testing.T) {
	runJavaBinaryHostTestCase(t, Bp2buildTestCase{
		Description: "java_binary_host with srcs, libs.",
		Filesystem:  testFs,
		Blueprint: `java_binary_host {
    name: "java-binary-host",
    manifest: "test.mf",
    srcs: ["a.java", "b.kt"],
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("kt_jvm_library", "java-binary-host_kt", AttrNameToString{
				"srcs": `[
        "a.java",
        "b.kt",
    ]`,
				"target_compatible_with": `select({
        "//build/bazel/platforms/os:android": ["@platforms//:incompatible"],
        "//conditions:default": [],
    })`,
			}),
			MakeBazelTarget("java_binary", "java-binary-host", AttrNameToString{
				"main_class":   `"com.android.test.MainClass"`,
				"runtime_deps": `[":java-binary-host_kt"]`,
				"target_compatible_with": `select({
        "//build/bazel/platforms/os:android": ["@platforms//:incompatible"],
        "//conditions:default": [],
    })`,
			}),
		},
	})
}

func TestJavaBinaryHostKotlinCommonSrcs(t *testing.T) {
	runJavaBinaryHostTestCase(t, Bp2buildTestCase{
		Description: "java_binary_host with common_srcs",
		Filesystem:  testFs,
		Blueprint: `java_binary_host {
    name: "java-binary-host",
    manifest: "test.mf",
    srcs: ["a.java"],
    common_srcs: ["b.kt"],
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("kt_jvm_library", "java-binary-host_kt", AttrNameToString{
				"srcs":        `["a.java"]`,
				"common_srcs": `["b.kt"]`,
				"target_compatible_with": `select({
        "//build/bazel/platforms/os:android": ["@platforms//:incompatible"],
        "//conditions:default": [],
    })`,
			}),
			MakeBazelTarget("java_binary", "java-binary-host", AttrNameToString{
				"main_class":   `"com.android.test.MainClass"`,
				"runtime_deps": `[":java-binary-host_kt"]`,
				"target_compatible_with": `select({
        "//build/bazel/platforms/os:android": ["@platforms//:incompatible"],
        "//conditions:default": [],
    })`,
			}),
		},
	})
}

func TestJavaBinaryHostKotlinWithResourceDir(t *testing.T) {
	runJavaBinaryHostTestCase(t, Bp2buildTestCase{
		Description: "java_binary_host with srcs, libs, resource dir  .",
		Filesystem: map[string]string{
			"test.mf":        "Main-Class: com.android.test.MainClass",
			"res/a.res":      "",
			"res/dir1/b.res": "",
		},
		Blueprint: `java_binary_host {
    name: "java-binary-host",
    manifest: "test.mf",
    srcs: ["a.java", "b.kt"],
    java_resource_dirs: ["res"],
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("kt_jvm_library", "java-binary-host_kt", AttrNameToString{
				"srcs": `[
        "a.java",
        "b.kt",
    ]`,
				"target_compatible_with": `select({
        "//build/bazel/platforms/os:android": ["@platforms//:incompatible"],
        "//conditions:default": [],
    })`,
			}),
			MakeBazelTarget("java_binary", "java-binary-host", AttrNameToString{
				"main_class":   `"com.android.test.MainClass"`,
				"runtime_deps": `[":java-binary-host_kt"]`,
				"resources": `[
        "res/a.res",
        "res/dir1/b.res",
    ]`,
				"resource_strip_prefix": `"res"`,
				"target_compatible_with": `select({
        "//build/bazel/platforms/os:android": ["@platforms//:incompatible"],
        "//conditions:default": [],
    })`,
			}),
		},
	})
}

func TestJavaBinaryHostKotlinWithResources(t *testing.T) {
	runJavaBinaryHostTestCase(t, Bp2buildTestCase{
		Description: "java_binary_host with srcs, libs, resources.",
		Filesystem: map[string]string{
			"test.mf":   "Main-Class: com.android.test.MainClass",
			"res/a.res": "",
			"res/b.res": "",
		},
		Blueprint: `java_binary_host {
    name: "java-binary-host",
    manifest: "test.mf",
    srcs: ["a.java", "b.kt"],
    java_resources: ["res/a.res", "res/b.res"],
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("kt_jvm_library", "java-binary-host_kt", AttrNameToString{
				"srcs": `[
        "a.java",
        "b.kt",
    ]`,
				"resources": `[
        "res/a.res",
        "res/b.res",
    ]`,
				"target_compatible_with": `select({
        "//build/bazel/platforms/os:android": ["@platforms//:incompatible"],
        "//conditions:default": [],
    })`,
			}),
			MakeBazelTarget("java_binary", "java-binary-host", AttrNameToString{
				"main_class":   `"com.android.test.MainClass"`,
				"runtime_deps": `[":java-binary-host_kt"]`,
				"target_compatible_with": `select({
        "//build/bazel/platforms/os:android": ["@platforms//:incompatible"],
        "//conditions:default": [],
    })`,
			}),
		},
	})
}
