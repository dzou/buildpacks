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
	"github.com/GoogleCloudPlatform/buildpacks/pkg/env"
	gcp "github.com/GoogleCloudPlatform/buildpacks/pkg/gcpbuildpack"
	"os"
	"path/filepath"
)

const (
	layerName  = "java-graalvm"
	graalvmUrl = "https://github.com/graalvm/graalvm-ce-builds/releases/download/vm-21.0.0.2/graalvm-ce-java11-linux-amd64-21.0.0.2.tar.gz"
)

func main() {
	gcp.Main(detectFn, buildFn)
}

func detectFn(ctx *gcp.Context) (gcp.DetectResult, error) {
	if _, ok := os.LookupEnv(env.FunctionTarget); ok {
		return gcp.OptInEnvSet(env.FunctionTarget), nil
	}
	return gcp.OptOutEnvNotSet(env.FunctionTarget), nil
}

func buildFn(ctx *gcp.Context) error {
	target, ok := os.LookupEnv(env.FunctionTarget)
	if !ok {
		return fmt.Errorf("No function target set.")
	}

	installGraalVm(ctx)

	if ctx.FileExists("pom.xml") {
		invokeMavenGraalvm(ctx)
	}

	ctx.AddWebProcess([]string{"./target/com.google.cloud.functions.invoker.runner.invoker", "--target", target})

	return nil
}

func installGraalVm(ctx *gcp.Context) error {
	graalLayer := ctx.Layer(layerName, gcp.CacheLayer, gcp.BuildLayer, gcp.LaunchLayerIfDevMode)

	// Install graalvm into layer.
	command := fmt.Sprintf(
		"curl --fail --show-error --silent --location %s "+
			"| tar xz --directory %s --strip-components=1", graalvmUrl, graalLayer.Path)
	ctx.Exec([]string{"bash", "-c", command}, gcp.WithUserAttribution)

	// Run gu install native-image
	graalUpdater := filepath.Join(graalLayer.Path, "bin", "gu")
	ctx.Exec([]string{graalUpdater, "install", "native-image"}, gcp.WithUserAttribution)

	// graalLayer.BuildEnvironment.Override("JAVA_HOME", filepath.Join(graalLayer.Path))
	ctx.Setenv("JAVA_HOME", graalLayer.Path)

	return nil
}

func invokeMavenGraalvm(ctx *gcp.Context) error {
	mvn := "mvn"

	// For now, let's just assume they have a Maven profile containing the native-image plugin.
	// This is what will invoke native-image compilation.
	ctx.Exec([]string{mvn, "package", "-P", "native"}, gcp.WithUserAttribution)

	return nil
}
