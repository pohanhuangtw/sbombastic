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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rancher/sbombastic/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/third_party/helm"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

var testenv env.Environment

const (
	releaseName     = "sbombastic"
	repoURL         = "https://helm.cilium.io/"
	kindClusterName = "sbombastic-dev"
)

var namespace = "default"

// envconf.RandomName("sbombastic-e2e-ns", 32)

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

func InstallSbomBastic(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
	curDir, _ := os.Getwd()
	chartPath := filepath.Join(curDir, "../../helm")
	manager := helm.New(cfg.KubeconfigFile())
	err := manager.RunRepo(helm.WithArgs("add", releaseName, repoURL))
	if err != nil {
		return nil, err
	}

	fmt.Println("namespace", namespace)

	err = manager.RunInstall(
		helm.WithName(releaseName),
		helm.WithChart(chartPath),
		helm.WithNamespace(namespace),
	)
	if err != nil {
		return nil, err
	}

	// Wait for a worker node to be ready
	client, err := cfg.NewClient()
	if err != nil {
		return nil, err
	}

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: kindClusterName + "-worker"},
	}

	wait.For(conditions.New(client.Resources()).ResourceMatch(node, func(object k8s.Object) bool {
		d := object.(*corev1.Node)
		status := false
		for _, v := range d.Status.Conditions {
			if v.Type == "Ready" && v.Status == "True" {
				status = true
			}
		}
		return status
	}), wait.WithTimeout(time.Minute*2))
	return ctx, nil
}

func UninstallSbomBastic(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
	manager := helm.New(cfg.KubeconfigFile())
	err := manager.RunRepo(helm.WithArgs("remove", releaseName))
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

func newScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()

	if err := apiextensionsv1.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		panic(err)
	}

	return scheme, nil
}

func newKubeClient(restConfig *rest.Config) (client.Client, error) {
	scheme, err := newScheme()
	if err != nil {
		return nil, err
	}

	return client.New(restConfig, client.Options{
		Scheme: scheme,
	})
}

func TestMain(m *testing.M) {
	// Step 1: Explicitly set the KUBECONFIG if not set
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = os.Getenv("HOME") + "/.kube/config" // or hardcode your kind config
	}

	cfg := envconf.NewWithKubeConfig(kubeconfig).WithNamespace(namespace)
	testenv = env.NewWithConfig(cfg)
	fmt.Println("namespace", namespace)

	// testenv.Setup(
	// 	envfuncs.CreateNamespace(namespace), // create namespace
	// 	InstallSbomBastic,
	// )

	// testenv.Finish(
	// 	envfuncs.DeleteNamespace(namespace),
	// 	UninstallSbomBastic,
	// )

	os.Exit(testenv.Run(m))
}
