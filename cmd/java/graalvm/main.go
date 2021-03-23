// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Implements java/functions_framework buildpack.
// The functions_framework buildpack copies the function framework into a layer, and adds it to a compiled function to make an executable app.
package main

import (
	"fmt"
	gcp "github.com/GoogleCloudPlatform/buildpacks/pkg/gcpbuildpack"
	"os"
	"path/filepath"
)

const (
	layerName = "java-graalvm"
	graalvmUrl = "https://github.com/graalvm/graalvm-ce-builds/releases/download/vm-21.0.0.2/graalvm-ce-java11-linux-amd64-21.0.0.2.tar.gz"
)

func main() {
	gcp.Main(detectFn, buildFn)
}

func detectFn(ctx *gcp.Context) (gcp.DetectResult, error) {
	if _, ok := os.LookupEnv("ENABLE_GRAALVM"); ok {
		return gcp.OptInEnvSet("ENABLE_GRAALVM"), nil
	}
	return gcp.OptOutEnvNotSet("ENABLE_GRAALVM"), nil
}

func buildFn(ctx *gcp.Context) error {
	graalLayer := ctx.Layer(layerName, gcp.CacheLayer, gcp.BuildLayer, gcp.LaunchLayerIfDevMode)
	graalLayer.BuildEnvironment.Override("JAVA_HOME", filepath.Join(graalLayer.Path))

	// Install graalvm into layer.
	command := fmt.Sprintf(
		"curl --fail --show-error --silent --location %s " +
			"| tar xz --directory %s --strip-components=1", graalvmUrl, graalLayer.Path)
	ctx.Exec([]string{"bash", "-c", command}, gcp.WithUserAttribution)

	// Run gu install native-image
	graalUpdater := filepath.Join(graalLayer.Path, "bin", "gu")
	ctx.Exec([]string{graalUpdater, "install", "native-image"}, gcp.WithUserAttribution)
	return nil
}
