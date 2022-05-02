// Copyright 2019 The Android Open Source Project
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
	"android/soong/android"
)

func init() {
	// Linux_cross-arm64 uses mostly the same rust toolchain as the Android-arm64 but with different link flags
	registerToolchainFactory(android.LinuxBionic, android.Arm64, arm64LinuxBionicToolchainFactory)
}

func arm64LinuxBionicToolchainFactory(arch android.Arch) Toolchain {
	return arm64ToolchainFactory(arch, "${cc_config.Arm64LinuxBionicLldflags}")
}
