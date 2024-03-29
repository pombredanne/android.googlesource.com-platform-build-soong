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
	"android/soong/android"
	"android/soong/apex"
	"android/soong/cc"
	"android/soong/etc"
	"android/soong/java"
	"android/soong/sh"

	"fmt"
	"testing"
)

func runApexTestCase(t *testing.T, tc Bp2buildTestCase) {
	t.Helper()
	RunBp2BuildTestCase(t, registerApexModuleTypes, tc)
}

func registerApexModuleTypes(ctx android.RegistrationContext) {
	// CC module types needed as they can be APEX dependencies
	cc.RegisterCCBuildComponents(ctx)

	ctx.RegisterModuleType("sh_binary", sh.ShBinaryFactory)
	ctx.RegisterModuleType("cc_binary", cc.BinaryFactory)
	ctx.RegisterModuleType("cc_library", cc.LibraryFactory)
	ctx.RegisterModuleType("apex_key", apex.ApexKeyFactory)
	ctx.RegisterModuleType("android_app_certificate", java.AndroidAppCertificateFactory)
	ctx.RegisterModuleType("filegroup", android.FileGroupFactory)
	ctx.RegisterModuleType("prebuilt_etc", etc.PrebuiltEtcFactory)
	ctx.RegisterModuleType("cc_test", cc.TestFactory)
}

func runOverrideApexTestCase(t *testing.T, tc Bp2buildTestCase) {
	t.Helper()
	RunBp2BuildTestCase(t, registerOverrideApexModuleTypes, tc)
}

func registerOverrideApexModuleTypes(ctx android.RegistrationContext) {
	// CC module types needed as they can be APEX dependencies
	cc.RegisterCCBuildComponents(ctx)

	ctx.RegisterModuleType("sh_binary", sh.ShBinaryFactory)
	ctx.RegisterModuleType("cc_binary", cc.BinaryFactory)
	ctx.RegisterModuleType("cc_library", cc.LibraryFactory)
	ctx.RegisterModuleType("apex_key", apex.ApexKeyFactory)
	ctx.RegisterModuleType("android_app_certificate", java.AndroidAppCertificateFactory)
	ctx.RegisterModuleType("filegroup", android.FileGroupFactory)
	ctx.RegisterModuleType("apex", apex.BundleFactory)
	ctx.RegisterModuleType("prebuilt_etc", etc.PrebuiltEtcFactory)
}

func TestApexBundleSimple(t *testing.T) {
	runApexTestCase(t, Bp2buildTestCase{
		Description:                "apex - example with all props, file_context is a module in same Android.bp",
		ModuleTypeUnderTest:        "apex",
		ModuleTypeUnderTestFactory: apex.BundleFactory,
		Filesystem:                 map[string]string{},
		Blueprint: `
apex_key {
	name: "com.android.apogee.key",
	public_key: "com.android.apogee.avbpubkey",
	private_key: "com.android.apogee.pem",
	bazel_module: { bp2build_available: false },
}

android_app_certificate {
	name: "com.android.apogee.certificate",
	certificate: "com.android.apogee",
	bazel_module: { bp2build_available: false },
}

cc_library {
	name: "native_shared_lib_1",
	bazel_module: { bp2build_available: false },
}

cc_library {
	name: "native_shared_lib_2",
	bazel_module: { bp2build_available: false },
}

prebuilt_etc {
	name: "prebuilt_1",
	bazel_module: { bp2build_available: false },
}

prebuilt_etc {
	name: "prebuilt_2",
	bazel_module: { bp2build_available: false },
}

filegroup {
	name: "com.android.apogee-file_contexts",
	srcs: [
		"com.android.apogee-file_contexts",
	],
	bazel_module: { bp2build_available: false },
}

cc_binary { name: "cc_binary_1", bazel_module: { bp2build_available: false } }
sh_binary { name: "sh_binary_2", bazel_module: { bp2build_available: false } }

apex {
	name: "com.android.apogee",
	manifest: "apogee_manifest.json",
	androidManifest: "ApogeeAndroidManifest.xml",
	file_contexts: ":com.android.apogee-file_contexts",
	min_sdk_version: "29",
	key: "com.android.apogee.key",
	certificate: ":com.android.apogee.certificate",
	updatable: false,
	installable: false,
	compressible: false,
	native_shared_libs: [
	    "native_shared_lib_1",
	    "native_shared_lib_2",
	],
	binaries: [
		"cc_binary_1",
		"sh_binary_2",
	],
	prebuilts: [
	    "prebuilt_1",
	    "prebuilt_2",
	],
	package_name: "com.android.apogee.test.package",
	logging_parent: "logging.parent",
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.android.apogee", AttrNameToString{
				"android_manifest": `"ApogeeAndroidManifest.xml"`,
				"binaries": `[
        ":cc_binary_1",
        ":sh_binary_2",
    ]`,
				"certificate":     `":com.android.apogee.certificate"`,
				"file_contexts":   `":com.android.apogee-file_contexts"`,
				"installable":     "False",
				"key":             `":com.android.apogee.key"`,
				"manifest":        `"apogee_manifest.json"`,
				"min_sdk_version": `"29"`,
				"native_shared_libs_32": `select({
        "//build/bazel/platforms/arch:arm": [
            ":native_shared_lib_1",
            ":native_shared_lib_2",
        ],
        "//build/bazel/platforms/arch:x86": [
            ":native_shared_lib_1",
            ":native_shared_lib_2",
        ],
        "//conditions:default": [],
    })`,
				"native_shared_libs_64": `select({
        "//build/bazel/platforms/arch:arm64": [
            ":native_shared_lib_1",
            ":native_shared_lib_2",
        ],
        "//build/bazel/platforms/arch:x86_64": [
            ":native_shared_lib_1",
            ":native_shared_lib_2",
        ],
        "//conditions:default": [],
    })`,
				"prebuilts": `[
        ":prebuilt_1",
        ":prebuilt_2",
    ]`,
				"updatable":      "False",
				"compressible":   "False",
				"package_name":   `"com.android.apogee.test.package"`,
				"logging_parent": `"logging.parent"`,
			}),
		}})
}

func TestApexBundleSimple_fileContextsInAnotherAndroidBp(t *testing.T) {
	runApexTestCase(t, Bp2buildTestCase{
		Description:                "apex - file contexts is a module in another Android.bp",
		ModuleTypeUnderTest:        "apex",
		ModuleTypeUnderTestFactory: apex.BundleFactory,
		Filesystem: map[string]string{
			"a/b/Android.bp": `
filegroup {
	name: "com.android.apogee-file_contexts",
	srcs: [
		"com.android.apogee-file_contexts",
	],
	bazel_module: { bp2build_available: false },
}
`,
		},
		Blueprint: `
apex {
	name: "com.android.apogee",
	file_contexts: ":com.android.apogee-file_contexts",
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.android.apogee", AttrNameToString{
				"file_contexts": `"//a/b:com.android.apogee-file_contexts"`,
				"manifest":      `"apex_manifest.json"`,
			}),
		}})
}

func TestApexBundleSimple_fileContextsIsFile(t *testing.T) {
	runApexTestCase(t, Bp2buildTestCase{
		Description:                "apex - file contexts is a file",
		ModuleTypeUnderTest:        "apex",
		ModuleTypeUnderTestFactory: apex.BundleFactory,
		Filesystem:                 map[string]string{},
		Blueprint: `
apex {
	name: "com.android.apogee",
	file_contexts: "file_contexts_file",
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.android.apogee", AttrNameToString{
				"file_contexts": `"file_contexts_file"`,
				"manifest":      `"apex_manifest.json"`,
			}),
		}})
}

func TestApexBundleSimple_fileContextsIsNotSpecified(t *testing.T) {
	runApexTestCase(t, Bp2buildTestCase{
		Description:                "apex - file contexts is not specified",
		ModuleTypeUnderTest:        "apex",
		ModuleTypeUnderTestFactory: apex.BundleFactory,
		Filesystem: map[string]string{
			"system/sepolicy/apex/Android.bp": `
filegroup {
	name: "com.android.apogee-file_contexts",
	srcs: [
		"com.android.apogee-file_contexts",
	],
	bazel_module: { bp2build_available: false },
}
`,
		},
		Blueprint: `
apex {
	name: "com.android.apogee",
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.android.apogee", AttrNameToString{
				"file_contexts": `"//system/sepolicy/apex:com.android.apogee-file_contexts"`,
				"manifest":      `"apex_manifest.json"`,
			}),
		}})
}

func TestApexBundleCompileMultilibBoth(t *testing.T) {
	runApexTestCase(t, Bp2buildTestCase{
		Description:                "apex - example with compile_multilib=both",
		ModuleTypeUnderTest:        "apex",
		ModuleTypeUnderTestFactory: apex.BundleFactory,
		Filesystem: map[string]string{
			"system/sepolicy/apex/Android.bp": `
filegroup {
	name: "com.android.apogee-file_contexts",
	srcs: [ "apogee-file_contexts", ],
	bazel_module: { bp2build_available: false },
}
`,
		},
		Blueprint: createMultilibBlueprint(`compile_multilib: "both",`),
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.android.apogee", AttrNameToString{
				"native_shared_libs_32": `[
        ":unnested_native_shared_lib",
        ":native_shared_lib_for_both",
        ":native_shared_lib_for_lib32",
    ] + select({
        "//build/bazel/platforms/arch:arm": [":native_shared_lib_for_first"],
        "//build/bazel/platforms/arch:x86": [":native_shared_lib_for_first"],
        "//conditions:default": [],
    })`,
				"native_shared_libs_64": `select({
        "//build/bazel/platforms/arch:arm64": [
            ":unnested_native_shared_lib",
            ":native_shared_lib_for_both",
            ":native_shared_lib_for_lib64",
            ":native_shared_lib_for_first",
        ],
        "//build/bazel/platforms/arch:x86_64": [
            ":unnested_native_shared_lib",
            ":native_shared_lib_for_both",
            ":native_shared_lib_for_lib64",
            ":native_shared_lib_for_first",
        ],
        "//conditions:default": [],
    })`,
				"file_contexts": `"//system/sepolicy/apex:com.android.apogee-file_contexts"`,
				"manifest":      `"apex_manifest.json"`,
			}),
		}})
}

func TestApexBundleCompileMultilibFirstAndDefaultValue(t *testing.T) {
	expectedBazelTargets := []string{
		MakeBazelTarget("apex", "com.android.apogee", AttrNameToString{
			"native_shared_libs_32": `select({
        "//build/bazel/platforms/arch:arm": [
            ":unnested_native_shared_lib",
            ":native_shared_lib_for_both",
            ":native_shared_lib_for_lib32",
            ":native_shared_lib_for_first",
        ],
        "//build/bazel/platforms/arch:x86": [
            ":unnested_native_shared_lib",
            ":native_shared_lib_for_both",
            ":native_shared_lib_for_lib32",
            ":native_shared_lib_for_first",
        ],
        "//conditions:default": [],
    })`,
			"native_shared_libs_64": `select({
        "//build/bazel/platforms/arch:arm64": [
            ":unnested_native_shared_lib",
            ":native_shared_lib_for_both",
            ":native_shared_lib_for_lib64",
            ":native_shared_lib_for_first",
        ],
        "//build/bazel/platforms/arch:x86_64": [
            ":unnested_native_shared_lib",
            ":native_shared_lib_for_both",
            ":native_shared_lib_for_lib64",
            ":native_shared_lib_for_first",
        ],
        "//conditions:default": [],
    })`,
			"file_contexts": `"//system/sepolicy/apex:com.android.apogee-file_contexts"`,
			"manifest":      `"apex_manifest.json"`,
		}),
	}

	// "first" is the default value of compile_multilib prop so `compile_multilib_: "first"` and unset compile_multilib
	// should result to the same bp2build output
	compileMultiLibPropValues := []string{`compile_multilib: "first",`, ""}
	for _, compileMultiLibProp := range compileMultiLibPropValues {
		descriptionSuffix := compileMultiLibProp
		if descriptionSuffix == "" {
			descriptionSuffix = "compile_multilib unset"
		}
		runApexTestCase(t, Bp2buildTestCase{
			Description:                "apex - example with " + compileMultiLibProp,
			ModuleTypeUnderTest:        "apex",
			ModuleTypeUnderTestFactory: apex.BundleFactory,
			Filesystem: map[string]string{
				"system/sepolicy/apex/Android.bp": `
    filegroup {
        name: "com.android.apogee-file_contexts",
        srcs: [ "apogee-file_contexts", ],
        bazel_module: { bp2build_available: false },
    }
    `,
			},
			Blueprint:            createMultilibBlueprint(compileMultiLibProp),
			ExpectedBazelTargets: expectedBazelTargets,
		})
	}
}

func TestApexBundleCompileMultilib32(t *testing.T) {
	runApexTestCase(t, Bp2buildTestCase{
		Description:                "apex - example with compile_multilib=32",
		ModuleTypeUnderTest:        "apex",
		ModuleTypeUnderTestFactory: apex.BundleFactory,
		Filesystem: map[string]string{
			"system/sepolicy/apex/Android.bp": `
filegroup {
	name: "com.android.apogee-file_contexts",
	srcs: [ "apogee-file_contexts", ],
	bazel_module: { bp2build_available: false },
}
`,
		},
		Blueprint: createMultilibBlueprint(`compile_multilib: "32",`),
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.android.apogee", AttrNameToString{
				"native_shared_libs_32": `[
        ":unnested_native_shared_lib",
        ":native_shared_lib_for_both",
        ":native_shared_lib_for_lib32",
    ] + select({
        "//build/bazel/platforms/arch:arm": [":native_shared_lib_for_first"],
        "//build/bazel/platforms/arch:x86": [":native_shared_lib_for_first"],
        "//conditions:default": [],
    })`,
				"file_contexts": `"//system/sepolicy/apex:com.android.apogee-file_contexts"`,
				"manifest":      `"apex_manifest.json"`,
			}),
		}})
}

func TestApexBundleCompileMultilib64(t *testing.T) {
	runApexTestCase(t, Bp2buildTestCase{
		Description:                "apex - example with compile_multilib=64",
		ModuleTypeUnderTest:        "apex",
		ModuleTypeUnderTestFactory: apex.BundleFactory,
		Filesystem: map[string]string{
			"system/sepolicy/apex/Android.bp": `
filegroup {
	name: "com.android.apogee-file_contexts",
	srcs: [ "apogee-file_contexts", ],
	bazel_module: { bp2build_available: false },
}
`,
		},
		Blueprint: createMultilibBlueprint(`compile_multilib: "64",`),
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.android.apogee", AttrNameToString{
				"native_shared_libs_64": `select({
        "//build/bazel/platforms/arch:arm64": [
            ":unnested_native_shared_lib",
            ":native_shared_lib_for_both",
            ":native_shared_lib_for_lib64",
            ":native_shared_lib_for_first",
        ],
        "//build/bazel/platforms/arch:x86_64": [
            ":unnested_native_shared_lib",
            ":native_shared_lib_for_both",
            ":native_shared_lib_for_lib64",
            ":native_shared_lib_for_first",
        ],
        "//conditions:default": [],
    })`,
				"file_contexts": `"//system/sepolicy/apex:com.android.apogee-file_contexts"`,
				"manifest":      `"apex_manifest.json"`,
			}),
		}})
}

func createMultilibBlueprint(compile_multilib string) string {
	return fmt.Sprintf(`
cc_library {
	name: "native_shared_lib_for_both",
	bazel_module: { bp2build_available: false },
}

cc_library {
	name: "native_shared_lib_for_first",
	bazel_module: { bp2build_available: false },
}

cc_library {
	name: "native_shared_lib_for_lib32",
	bazel_module: { bp2build_available: false },
}

cc_library {
	name: "native_shared_lib_for_lib64",
	bazel_module: { bp2build_available: false },
}

cc_library {
	name: "unnested_native_shared_lib",
	bazel_module: { bp2build_available: false },
}

apex {
	name: "com.android.apogee",
	%s
	native_shared_libs: ["unnested_native_shared_lib"],
	multilib: {
		both: {
			native_shared_libs: [
				"native_shared_lib_for_both",
			],
		},
		first: {
			native_shared_libs: [
				"native_shared_lib_for_first",
			],
		},
		lib32: {
			native_shared_libs: [
				"native_shared_lib_for_lib32",
			],
		},
		lib64: {
			native_shared_libs: [
				"native_shared_lib_for_lib64",
			],
		},
	},
}`, compile_multilib)
}

func TestApexBundleDefaultPropertyValues(t *testing.T) {
	runApexTestCase(t, Bp2buildTestCase{
		Description:                "apex - default property values",
		ModuleTypeUnderTest:        "apex",
		ModuleTypeUnderTestFactory: apex.BundleFactory,
		Filesystem: map[string]string{
			"system/sepolicy/apex/Android.bp": `
filegroup {
	name: "com.android.apogee-file_contexts",
	srcs: [ "apogee-file_contexts", ],
	bazel_module: { bp2build_available: false },
}
`,
		},
		Blueprint: `
apex {
	name: "com.android.apogee",
	manifest: "apogee_manifest.json",
}
`,
		ExpectedBazelTargets: []string{MakeBazelTarget("apex", "com.android.apogee", AttrNameToString{
			"manifest":      `"apogee_manifest.json"`,
			"file_contexts": `"//system/sepolicy/apex:com.android.apogee-file_contexts"`,
		}),
		}})
}

func TestApexBundleHasBazelModuleProps(t *testing.T) {
	runApexTestCase(t, Bp2buildTestCase{
		Description:                "apex - has bazel module props",
		ModuleTypeUnderTest:        "apex",
		ModuleTypeUnderTestFactory: apex.BundleFactory,
		Filesystem: map[string]string{
			"system/sepolicy/apex/Android.bp": `
filegroup {
	name: "apogee-file_contexts",
	srcs: [ "apogee-file_contexts", ],
	bazel_module: { bp2build_available: false },
}
`,
		},
		Blueprint: `
apex {
	name: "apogee",
	manifest: "manifest.json",
	bazel_module: { bp2build_available: true },
}
`,
		ExpectedBazelTargets: []string{MakeBazelTarget("apex", "apogee", AttrNameToString{
			"manifest":      `"manifest.json"`,
			"file_contexts": `"//system/sepolicy/apex:apogee-file_contexts"`,
		}),
		}})
}

func TestBp2BuildOverrideApex(t *testing.T) {
	runOverrideApexTestCase(t, Bp2buildTestCase{
		Description:                "override_apex",
		ModuleTypeUnderTest:        "override_apex",
		ModuleTypeUnderTestFactory: apex.OverrideApexFactory,
		Filesystem:                 map[string]string{},
		Blueprint: `
apex_key {
	name: "com.android.apogee.key",
	public_key: "com.android.apogee.avbpubkey",
	private_key: "com.android.apogee.pem",
	bazel_module: { bp2build_available: false },
}

android_app_certificate {
	name: "com.android.apogee.certificate",
	certificate: "com.android.apogee",
	bazel_module: { bp2build_available: false },
}

cc_library {
	name: "native_shared_lib_1",
	bazel_module: { bp2build_available: false },
}

cc_library {
	name: "native_shared_lib_2",
	bazel_module: { bp2build_available: false },
}

prebuilt_etc {
	name: "prebuilt_1",
	bazel_module: { bp2build_available: false },
}

prebuilt_etc {
	name: "prebuilt_2",
	bazel_module: { bp2build_available: false },
}

filegroup {
	name: "com.android.apogee-file_contexts",
	srcs: [
		"com.android.apogee-file_contexts",
	],
	bazel_module: { bp2build_available: false },
}

cc_binary { name: "cc_binary_1", bazel_module: { bp2build_available: false } }
sh_binary { name: "sh_binary_2", bazel_module: { bp2build_available: false } }

apex {
	name: "com.android.apogee",
	manifest: "apogee_manifest.json",
	androidManifest: "ApogeeAndroidManifest.xml",
	file_contexts: ":com.android.apogee-file_contexts",
	min_sdk_version: "29",
	key: "com.android.apogee.key",
	certificate: ":com.android.apogee.certificate",
	updatable: false,
	installable: false,
	compressible: false,
	native_shared_libs: [
	    "native_shared_lib_1",
	    "native_shared_lib_2",
	],
	binaries: [
		"cc_binary_1",
		"sh_binary_2",
	],
	prebuilts: [
	    "prebuilt_1",
	    "prebuilt_2",
	],
	bazel_module: { bp2build_available: false },
}

apex_key {
	name: "com.google.android.apogee.key",
	public_key: "com.google.android.apogee.avbpubkey",
	private_key: "com.google.android.apogee.pem",
	bazel_module: { bp2build_available: false },
}

android_app_certificate {
	name: "com.google.android.apogee.certificate",
	certificate: "com.google.android.apogee",
	bazel_module: { bp2build_available: false },
}

override_apex {
	name: "com.google.android.apogee",
	base: ":com.android.apogee",
	key: "com.google.android.apogee.key",
	certificate: ":com.google.android.apogee.certificate",
	prebuilts: [],
	compressible: true,
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.google.android.apogee", AttrNameToString{
				"android_manifest": `"ApogeeAndroidManifest.xml"`,
				"base_apex_name":   `"com.android.apogee"`,
				"binaries": `[
        ":cc_binary_1",
        ":sh_binary_2",
    ]`,
				"certificate":     `":com.google.android.apogee.certificate"`,
				"file_contexts":   `":com.android.apogee-file_contexts"`,
				"installable":     "False",
				"key":             `":com.google.android.apogee.key"`,
				"manifest":        `"apogee_manifest.json"`,
				"min_sdk_version": `"29"`,
				"native_shared_libs_32": `select({
        "//build/bazel/platforms/arch:arm": [
            ":native_shared_lib_1",
            ":native_shared_lib_2",
        ],
        "//build/bazel/platforms/arch:x86": [
            ":native_shared_lib_1",
            ":native_shared_lib_2",
        ],
        "//conditions:default": [],
    })`,
				"native_shared_libs_64": `select({
        "//build/bazel/platforms/arch:arm64": [
            ":native_shared_lib_1",
            ":native_shared_lib_2",
        ],
        "//build/bazel/platforms/arch:x86_64": [
            ":native_shared_lib_1",
            ":native_shared_lib_2",
        ],
        "//conditions:default": [],
    })`,
				"prebuilts":    `[]`,
				"updatable":    "False",
				"compressible": "True",
			}),
		}})
}

func TestApexBundleSimple_manifestIsEmpty_baseApexOverrideApexInDifferentAndroidBp(t *testing.T) {
	runOverrideApexTestCase(t, Bp2buildTestCase{
		Description:                "override_apex - manifest of base apex is empty, base apex and override_apex is in different Android.bp",
		ModuleTypeUnderTest:        "override_apex",
		ModuleTypeUnderTestFactory: apex.OverrideApexFactory,
		Filesystem: map[string]string{
			"system/sepolicy/apex/Android.bp": `
filegroup {
	name: "com.android.apogee-file_contexts",
	srcs: [ "apogee-file_contexts", ],
	bazel_module: { bp2build_available: false },
}`,
			"a/b/Android.bp": `
apex {
	name: "com.android.apogee",
	bazel_module: { bp2build_available: false },
}
`,
		},
		Blueprint: `
override_apex {
	name: "com.google.android.apogee",
	base: ":com.android.apogee",
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.google.android.apogee", AttrNameToString{
				"base_apex_name": `"com.android.apogee"`,
				"file_contexts":  `"//system/sepolicy/apex:com.android.apogee-file_contexts"`,
				"manifest":       `"//a/b:apex_manifest.json"`,
			}),
		}})
}

func TestApexBundleSimple_manifestIsSet_baseApexOverrideApexInDifferentAndroidBp(t *testing.T) {
	runOverrideApexTestCase(t, Bp2buildTestCase{
		Description:                "override_apex - manifest of base apex is set, base apex and override_apex is in different Android.bp",
		ModuleTypeUnderTest:        "override_apex",
		ModuleTypeUnderTestFactory: apex.OverrideApexFactory,
		Filesystem: map[string]string{
			"system/sepolicy/apex/Android.bp": `
filegroup {
	name: "com.android.apogee-file_contexts",
	srcs: [ "apogee-file_contexts", ],
	bazel_module: { bp2build_available: false },
}`,
			"a/b/Android.bp": `
apex {
	name: "com.android.apogee",
  manifest: "apogee_manifest.json",
	bazel_module: { bp2build_available: false },
}
`,
		},
		Blueprint: `
override_apex {
	name: "com.google.android.apogee",
  base: ":com.android.apogee",
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.google.android.apogee", AttrNameToString{
				"base_apex_name": `"com.android.apogee"`,
				"file_contexts":  `"//system/sepolicy/apex:com.android.apogee-file_contexts"`,
				"manifest":       `"//a/b:apogee_manifest.json"`,
			}),
		}})
}

func TestApexBundleSimple_manifestIsEmpty_baseApexOverrideApexInSameAndroidBp(t *testing.T) {
	runOverrideApexTestCase(t, Bp2buildTestCase{
		Description:                "override_apex - manifest of base apex is empty, base apex and override_apex is in same Android.bp",
		ModuleTypeUnderTest:        "override_apex",
		ModuleTypeUnderTestFactory: apex.OverrideApexFactory,
		Filesystem: map[string]string{
			"system/sepolicy/apex/Android.bp": `
filegroup {
	name: "com.android.apogee-file_contexts",
	srcs: [ "apogee-file_contexts", ],
	bazel_module: { bp2build_available: false },
}`,
		},
		Blueprint: `
apex {
	name: "com.android.apogee",
	bazel_module: { bp2build_available: false },
}

override_apex {
	name: "com.google.android.apogee",
  base: ":com.android.apogee",
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.google.android.apogee", AttrNameToString{
				"base_apex_name": `"com.android.apogee"`,
				"file_contexts":  `"//system/sepolicy/apex:com.android.apogee-file_contexts"`,
				"manifest":       `"apex_manifest.json"`,
			}),
		}})
}

func TestApexBundleSimple_manifestIsSet_baseApexOverrideApexInSameAndroidBp(t *testing.T) {
	runOverrideApexTestCase(t, Bp2buildTestCase{
		Description:                "override_apex - manifest of base apex is set, base apex and override_apex is in same Android.bp",
		ModuleTypeUnderTest:        "override_apex",
		ModuleTypeUnderTestFactory: apex.OverrideApexFactory,
		Filesystem: map[string]string{
			"system/sepolicy/apex/Android.bp": `
filegroup {
	name: "com.android.apogee-file_contexts",
	srcs: [ "apogee-file_contexts", ],
	bazel_module: { bp2build_available: false },
}`,
		},
		Blueprint: `
apex {
	name: "com.android.apogee",
  manifest: "apogee_manifest.json",
	bazel_module: { bp2build_available: false },
}

override_apex {
	name: "com.google.android.apogee",
  base: ":com.android.apogee",
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.google.android.apogee", AttrNameToString{
				"base_apex_name": `"com.android.apogee"`,
				"file_contexts":  `"//system/sepolicy/apex:com.android.apogee-file_contexts"`,
				"manifest":       `"apogee_manifest.json"`,
			}),
		}})
}

func TestApexBundleSimple_packageNameOverride(t *testing.T) {
	runOverrideApexTestCase(t, Bp2buildTestCase{
		Description:                "override_apex - override package name",
		ModuleTypeUnderTest:        "override_apex",
		ModuleTypeUnderTestFactory: apex.OverrideApexFactory,
		Filesystem: map[string]string{
			"system/sepolicy/apex/Android.bp": `
filegroup {
	name: "com.android.apogee-file_contexts",
	srcs: [ "apogee-file_contexts", ],
	bazel_module: { bp2build_available: false },
}`,
		},
		Blueprint: `
apex {
	name: "com.android.apogee",
	bazel_module: { bp2build_available: false },
}

override_apex {
	name: "com.google.android.apogee",
	base: ":com.android.apogee",
	package_name: "com.google.android.apogee",
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.google.android.apogee", AttrNameToString{
				"base_apex_name": `"com.android.apogee"`,
				"file_contexts":  `"//system/sepolicy/apex:com.android.apogee-file_contexts"`,
				"manifest":       `"apex_manifest.json"`,
				"package_name":   `"com.google.android.apogee"`,
			}),
		}})
}

func TestApexBundleSimple_NoPrebuiltsOverride(t *testing.T) {
	runOverrideApexTestCase(t, Bp2buildTestCase{
		Description:                "override_apex - no override",
		ModuleTypeUnderTest:        "override_apex",
		ModuleTypeUnderTestFactory: apex.OverrideApexFactory,
		Filesystem: map[string]string{
			"system/sepolicy/apex/Android.bp": `
filegroup {
	name: "com.android.apogee-file_contexts",
	srcs: [ "apogee-file_contexts", ],
	bazel_module: { bp2build_available: false },
}`,
		},
		Blueprint: `
prebuilt_etc {
	name: "prebuilt_file",
	bazel_module: { bp2build_available: false },
}

apex {
	name: "com.android.apogee",
	bazel_module: { bp2build_available: false },
    prebuilts: ["prebuilt_file"]
}

override_apex {
	name: "com.google.android.apogee",
	base: ":com.android.apogee",
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.google.android.apogee", AttrNameToString{
				"base_apex_name": `"com.android.apogee"`,
				"file_contexts":  `"//system/sepolicy/apex:com.android.apogee-file_contexts"`,
				"manifest":       `"apex_manifest.json"`,
				"prebuilts":      `[":prebuilt_file"]`,
			}),
		}})
}

func TestApexBundleSimple_PrebuiltsOverride(t *testing.T) {
	runOverrideApexTestCase(t, Bp2buildTestCase{
		Description:                "override_apex - ooverride",
		ModuleTypeUnderTest:        "override_apex",
		ModuleTypeUnderTestFactory: apex.OverrideApexFactory,
		Filesystem: map[string]string{
			"system/sepolicy/apex/Android.bp": `
filegroup {
	name: "com.android.apogee-file_contexts",
	srcs: [ "apogee-file_contexts", ],
	bazel_module: { bp2build_available: false },
}`,
		},
		Blueprint: `
prebuilt_etc {
	name: "prebuilt_file",
	bazel_module: { bp2build_available: false },
}

prebuilt_etc {
	name: "prebuilt_file2",
	bazel_module: { bp2build_available: false },
}

apex {
	name: "com.android.apogee",
	bazel_module: { bp2build_available: false },
    prebuilts: ["prebuilt_file"]
}

override_apex {
	name: "com.google.android.apogee",
	base: ":com.android.apogee",
    prebuilts: ["prebuilt_file2"]
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.google.android.apogee", AttrNameToString{
				"base_apex_name": `"com.android.apogee"`,
				"file_contexts":  `"//system/sepolicy/apex:com.android.apogee-file_contexts"`,
				"manifest":       `"apex_manifest.json"`,
				"prebuilts":      `[":prebuilt_file2"]`,
			}),
		}})
}

func TestApexBundleSimple_PrebuiltsOverrideEmptyList(t *testing.T) {
	runOverrideApexTestCase(t, Bp2buildTestCase{
		Description:                "override_apex - override with empty list",
		ModuleTypeUnderTest:        "override_apex",
		ModuleTypeUnderTestFactory: apex.OverrideApexFactory,
		Filesystem: map[string]string{
			"system/sepolicy/apex/Android.bp": `
filegroup {
	name: "com.android.apogee-file_contexts",
	srcs: [ "apogee-file_contexts", ],
	bazel_module: { bp2build_available: false },
}`,
		},
		Blueprint: `
prebuilt_etc {
	name: "prebuilt_file",
	bazel_module: { bp2build_available: false },
}

apex {
	name: "com.android.apogee",
	bazel_module: { bp2build_available: false },
    prebuilts: ["prebuilt_file"]
}

override_apex {
	name: "com.google.android.apogee",
	base: ":com.android.apogee",
    prebuilts: [],
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.google.android.apogee", AttrNameToString{
				"base_apex_name": `"com.android.apogee"`,
				"file_contexts":  `"//system/sepolicy/apex:com.android.apogee-file_contexts"`,
				"manifest":       `"apex_manifest.json"`,
				"prebuilts":      `[]`,
			}),
		}})
}

func TestApexBundleSimple_NoLoggingParentOverride(t *testing.T) {
	runOverrideApexTestCase(t, Bp2buildTestCase{
		Description:                "override_apex - logging_parent - no override",
		ModuleTypeUnderTest:        "override_apex",
		ModuleTypeUnderTestFactory: apex.OverrideApexFactory,
		Filesystem: map[string]string{
			"system/sepolicy/apex/Android.bp": `
filegroup {
	name: "com.android.apogee-file_contexts",
	srcs: [ "apogee-file_contexts", ],
	bazel_module: { bp2build_available: false },
}`,
		},
		Blueprint: `
apex {
	name: "com.android.apogee",
	bazel_module: { bp2build_available: false },
	logging_parent: "foo.bar.baz",
}

override_apex {
	name: "com.google.android.apogee",
	base: ":com.android.apogee",
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.google.android.apogee", AttrNameToString{
				"base_apex_name": `"com.android.apogee"`,
				"file_contexts":  `"//system/sepolicy/apex:com.android.apogee-file_contexts"`,
				"manifest":       `"apex_manifest.json"`,
				"logging_parent": `"foo.bar.baz"`,
			}),
		}})
}

func TestApexBundleSimple_LoggingParentOverride(t *testing.T) {
	runOverrideApexTestCase(t, Bp2buildTestCase{
		Description:                "override_apex - logging_parent - override",
		ModuleTypeUnderTest:        "override_apex",
		ModuleTypeUnderTestFactory: apex.OverrideApexFactory,
		Filesystem: map[string]string{
			"system/sepolicy/apex/Android.bp": `
filegroup {
	name: "com.android.apogee-file_contexts",
	srcs: [ "apogee-file_contexts", ],
	bazel_module: { bp2build_available: false },
}`,
		},
		Blueprint: `
apex {
	name: "com.android.apogee",
	bazel_module: { bp2build_available: false },
	logging_parent: "foo.bar.baz",
}

override_apex {
	name: "com.google.android.apogee",
	base: ":com.android.apogee",
	logging_parent: "foo.bar.baz.override",
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.google.android.apogee", AttrNameToString{
				"base_apex_name": `"com.android.apogee"`,
				"file_contexts":  `"//system/sepolicy/apex:com.android.apogee-file_contexts"`,
				"manifest":       `"apex_manifest.json"`,
				"logging_parent": `"foo.bar.baz.override"`,
			}),
		}})
}

func TestBp2BuildOverrideApex_CertificateNil(t *testing.T) {
	runOverrideApexTestCase(t, Bp2buildTestCase{
		Description:                "override_apex - don't set default certificate",
		ModuleTypeUnderTest:        "override_apex",
		ModuleTypeUnderTestFactory: apex.OverrideApexFactory,
		Filesystem:                 map[string]string{},
		Blueprint: `
android_app_certificate {
	name: "com.android.apogee.certificate",
	certificate: "com.android.apogee",
	bazel_module: { bp2build_available: false },
}

filegroup {
	name: "com.android.apogee-file_contexts",
	srcs: [
		"com.android.apogee-file_contexts",
	],
	bazel_module: { bp2build_available: false },
}

apex {
	name: "com.android.apogee",
	manifest: "apogee_manifest.json",
	file_contexts: ":com.android.apogee-file_contexts",
	certificate: ":com.android.apogee.certificate",
	bazel_module: { bp2build_available: false },
}

override_apex {
	name: "com.google.android.apogee",
	base: ":com.android.apogee",
	// certificate is deliberately omitted, and not converted to bazel,
	// because the overridden apex shouldn't be using the base apex's cert.
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.google.android.apogee", AttrNameToString{
				"base_apex_name": `"com.android.apogee"`,
				"file_contexts":  `":com.android.apogee-file_contexts"`,
				"manifest":       `"apogee_manifest.json"`,
			}),
		}})
}

func TestApexCertificateIsModule(t *testing.T) {
	runApexTestCase(t, Bp2buildTestCase{
		Description:                "apex - certificate is module",
		ModuleTypeUnderTest:        "apex",
		ModuleTypeUnderTestFactory: apex.BundleFactory,
		Filesystem:                 map[string]string{},
		Blueprint: `
android_app_certificate {
	name: "com.android.apogee.certificate",
	certificate: "com.android.apogee",
	bazel_module: { bp2build_available: false },
}

apex {
	name: "com.android.apogee",
	manifest: "apogee_manifest.json",
	file_contexts: ":com.android.apogee-file_contexts",
	certificate: ":com.android.apogee.certificate",
}
` + simpleModuleDoNotConvertBp2build("filegroup", "com.android.apogee-file_contexts"),
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.android.apogee", AttrNameToString{
				"certificate":   `":com.android.apogee.certificate"`,
				"file_contexts": `":com.android.apogee-file_contexts"`,
				"manifest":      `"apogee_manifest.json"`,
			}),
		}})
}

func TestApexWithStubLib(t *testing.T) {
	runApexTestCase(t, Bp2buildTestCase{
		Description:                "apex - static variant of stub lib should not have apex_available tag",
		ModuleTypeUnderTest:        "apex",
		ModuleTypeUnderTestFactory: apex.BundleFactory,
		Filesystem:                 map[string]string{},
		Blueprint: `
cc_library{
	name: "foo",
	stubs: { symbol_file: "foo.map.txt", versions: ["28", "29", "current"] },
	apex_available: ["myapex"],
}

cc_binary{
	name: "bar",
	static_libs: ["foo"],
	apex_available: ["myapex"],
}

apex {
	name: "myapex",
	manifest: "myapex_manifest.json",
	file_contexts: ":myapex-file_contexts",
	binaries: ["bar"],
	native_shared_libs: ["foo"],
}
` + simpleModuleDoNotConvertBp2build("filegroup", "myapex-file_contexts"),
		ExpectedBazelTargets: []string{
			MakeBazelTarget("cc_binary", "bar", AttrNameToString{
				"local_includes": `["."]`,
				"deps":           `[":foo_bp2build_cc_library_static"]`,
				"tags":           `["apex_available=myapex"]`,
			}),
			MakeBazelTarget("cc_library_static", "foo_bp2build_cc_library_static", AttrNameToString{
				"local_includes": `["."]`,
			}),
			MakeBazelTarget("cc_library_shared", "foo", AttrNameToString{
				"local_includes":    `["."]`,
				"stubs_symbol_file": `"foo.map.txt"`,
				"tags":              `["apex_available=myapex"]`,
			}),
			MakeBazelTarget("cc_stub_suite", "foo_stub_libs", AttrNameToString{
				"soname":         `"foo.so"`,
				"source_library": `":foo"`,
				"symbol_file":    `"foo.map.txt"`,
				"versions": `[
        "28",
        "29",
        "current",
    ]`,
			}),
			MakeBazelTarget("apex", "myapex", AttrNameToString{
				"file_contexts": `":myapex-file_contexts"`,
				"manifest":      `"myapex_manifest.json"`,
				"binaries":      `[":bar"]`,
				"native_shared_libs_32": `select({
        "//build/bazel/platforms/arch:arm": [":foo"],
        "//build/bazel/platforms/arch:x86": [":foo"],
        "//conditions:default": [],
    })`,
				"native_shared_libs_64": `select({
        "//build/bazel/platforms/arch:arm64": [":foo"],
        "//build/bazel/platforms/arch:x86_64": [":foo"],
        "//conditions:default": [],
    })`,
			}),
		},
	})
}

func TestApexCertificateIsSrc(t *testing.T) {
	runApexTestCase(t, Bp2buildTestCase{
		Description:                "apex - certificate is src",
		ModuleTypeUnderTest:        "apex",
		ModuleTypeUnderTestFactory: apex.BundleFactory,
		Filesystem:                 map[string]string{},
		Blueprint: `
apex {
	name: "com.android.apogee",
	manifest: "apogee_manifest.json",
	file_contexts: ":com.android.apogee-file_contexts",
	certificate: "com.android.apogee.certificate",
}
` + simpleModuleDoNotConvertBp2build("filegroup", "com.android.apogee-file_contexts"),
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.android.apogee", AttrNameToString{
				"certificate_name": `"com.android.apogee.certificate"`,
				"file_contexts":    `":com.android.apogee-file_contexts"`,
				"manifest":         `"apogee_manifest.json"`,
			}),
		}})
}

func TestBp2BuildOverrideApex_CertificateIsModule(t *testing.T) {
	runOverrideApexTestCase(t, Bp2buildTestCase{
		Description:                "override_apex - certificate is module",
		ModuleTypeUnderTest:        "override_apex",
		ModuleTypeUnderTestFactory: apex.OverrideApexFactory,
		Filesystem:                 map[string]string{},
		Blueprint: `
android_app_certificate {
	name: "com.android.apogee.certificate",
	certificate: "com.android.apogee",
	bazel_module: { bp2build_available: false },
}

filegroup {
	name: "com.android.apogee-file_contexts",
	srcs: [
		"com.android.apogee-file_contexts",
	],
	bazel_module: { bp2build_available: false },
}

apex {
	name: "com.android.apogee",
	manifest: "apogee_manifest.json",
	file_contexts: ":com.android.apogee-file_contexts",
	certificate: ":com.android.apogee.certificate",
	bazel_module: { bp2build_available: false },
}

android_app_certificate {
	name: "com.google.android.apogee.certificate",
	certificate: "com.google.android.apogee",
	bazel_module: { bp2build_available: false },
}

override_apex {
	name: "com.google.android.apogee",
	base: ":com.android.apogee",
	certificate: ":com.google.android.apogee.certificate",
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.google.android.apogee", AttrNameToString{
				"base_apex_name": `"com.android.apogee"`,
				"file_contexts":  `":com.android.apogee-file_contexts"`,
				"certificate":    `":com.google.android.apogee.certificate"`,
				"manifest":       `"apogee_manifest.json"`,
			}),
		}})
}

func TestBp2BuildOverrideApex_CertificateIsSrc(t *testing.T) {
	runOverrideApexTestCase(t, Bp2buildTestCase{
		Description:                "override_apex - certificate is src",
		ModuleTypeUnderTest:        "override_apex",
		ModuleTypeUnderTestFactory: apex.OverrideApexFactory,
		Filesystem:                 map[string]string{},
		Blueprint: `
android_app_certificate {
	name: "com.android.apogee.certificate",
	certificate: "com.android.apogee",
	bazel_module: { bp2build_available: false },
}

filegroup {
	name: "com.android.apogee-file_contexts",
	srcs: [
		"com.android.apogee-file_contexts",
	],
	bazel_module: { bp2build_available: false },
}

apex {
	name: "com.android.apogee",
	manifest: "apogee_manifest.json",
	file_contexts: ":com.android.apogee-file_contexts",
	certificate: ":com.android.apogee.certificate",
	bazel_module: { bp2build_available: false },
}

override_apex {
	name: "com.google.android.apogee",
	base: ":com.android.apogee",
	certificate: "com.google.android.apogee.certificate",
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "com.google.android.apogee", AttrNameToString{
				"base_apex_name":   `"com.android.apogee"`,
				"file_contexts":    `":com.android.apogee-file_contexts"`,
				"certificate_name": `"com.google.android.apogee.certificate"`,
				"manifest":         `"apogee_manifest.json"`,
			}),
		}})
}

func TestApexTestBundleSimple(t *testing.T) {
	runApexTestCase(t, Bp2buildTestCase{
		Description:                "apex_test - simple",
		ModuleTypeUnderTest:        "apex_test",
		ModuleTypeUnderTestFactory: apex.TestApexBundleFactory,
		Filesystem:                 map[string]string{},
		Blueprint: `
cc_test { name: "cc_test_1", bazel_module: { bp2build_available: false } }

apex_test {
	name: "test_com.android.apogee",
	file_contexts: "file_contexts_file",
	tests: ["cc_test_1"],
}
`,
		ExpectedBazelTargets: []string{
			MakeBazelTarget("apex", "test_com.android.apogee", AttrNameToString{
				"file_contexts": `"file_contexts_file"`,
				"manifest":      `"apex_manifest.json"`,
				"testonly":      `True`,
				"tests":         `[":cc_test_1"]`,
			}),
		}})
}
