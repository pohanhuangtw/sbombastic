/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/kind"
)

var testenv env.Environment

const (
	releaseName = "sbombastic"
	repoURL     = "https://helm.cilium.io/"
)

var namespace = envconf.RandomName("sbombastic-e2e-ns", 32)
var kindClusterName = envconf.RandomName("sbombastic-e2e-cluster", 32)

// func TestMain(m *testing.M) {
// 	testenv = env.New()
// 	kindClusterName := envconf.RandomName("sbombastic-e2e-cluster", 32)
// 	namespace := envconf.RandomName("sbombastic-e2e-ns", 32)
// 	testenv.Setup(
// 		envfuncs.CreateCluster(kind.NewProvider(), kindClusterName),
// 		envfuncs.CreateNamespace(namespace),
// 	)
// 	testenv.Finish(
// 		envfuncs.DeleteNamespace(namespace),
// 		envfuncs.DestroyCluster(kindClusterName),
// 	)
// 	os.Exit(testenv.Run(m))
// }

var (
	workerImage     = "ghcr.io/rancher-sandbox/sbombastic/worker:v0.1.0-alpha1"
	controllerImage = "ghcr.io/rancher-sandbox/sbombastic/controller:v0.1.0-alpha1"
	storageImage    = "ghcr.io/rancher-sandbox/sbombastic/storage:v0.1.0-alpha1"
)

func TestMain(m *testing.M) {
	// Step 1: Explicitly set the KUBECONFIG if not set
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = os.Getenv("HOME") + "/.kube/config" // or hardcode your kind config
	}

	cfg, _ := envconf.NewFromFlags()
	testenv = env.NewWithConfig(cfg)

	testenv.Setup(
		envfuncs.CreateCluster(kind.NewProvider(), kindClusterName),
		envfuncs.CreateNamespace(namespace), // create namespace
		envfuncs.LoadImageToCluster(kindClusterName, workerImage, "--verbose", "--mode", "direct"),
		envfuncs.LoadImageToCluster(kindClusterName, controllerImage, "--verbose", "--mode", "direct"),
		envfuncs.LoadImageToCluster(kindClusterName, storageImage, "--verbose", "--mode", "direct"),
	)

	testenv.Finish(
		envfuncs.DeleteNamespace(namespace),
		envfuncs.DestroyCluster(kindClusterName),
	)

	os.Exit(testenv.Run(m))
}
