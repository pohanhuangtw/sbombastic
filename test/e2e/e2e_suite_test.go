package e2e

import (
	"context"
	"fmt"
	"testing"

	storagev1alpha1 "github.com/rancher/sbombastic/api/storage/v1alpha1"
	v1alpha1 "github.com/rancher/sbombastic/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/third_party/helm"
)

func EqualReference(img storagev1alpha1.ImageMetadata, registryURI, registryRepository, tag string) bool {
	return img.RegistryURI == registryURI &&
		img.Repository == registryRepository &&
		img.Tag == tag
}

func TestRegistryCreation(t *testing.T) {
	releaseName := "sbombastic"

	// spdxPath := filepath.Join("..", "fixtures", "golang-1.12-alpine.spdx.json")
	// reportPath := filepath.Join("..", "fixtures", "golang-1.12-alpine.sarif.json")

	// registryName := "test-registry"
	// registryURI := "ghcr.io"
	// registryRepository := "rancher-sandbox/sbombastic/test-assets/golang"
	// golangAlpineTag := "1.12-alpine"

	// pollInterval := 1 * time.Second
	// pollTimeout := 60 * time.Second
	// var sbom storagev1alpha1.SBOM

	f := features.New("Registry CR Creation test").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			manager := helm.New(cfg.KubeconfigFile())
			fmt.Println("installing sbombastic helm chart", t.Name())
			err := manager.RunInstall(helm.WithName(releaseName),
				helm.WithNamespace(cfg.Namespace()),
				helm.WithChart("../../helm"),
				helm.WithWait(),
				helm.WithTimeout("3m"))

			assert.NoError(t, err, "sbombastic helm chart is not installed correctly")

			client := cfg.Client()
			storagev1alpha1.AddToScheme(client.Resources(cfg.Namespace()).GetScheme())
			v1alpha1.AddToScheme(client.Resources(cfg.Namespace()).GetScheme())
			return ctx
		}).
		// Assess("Create Registry CR", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	client := cfg.Client()
		// 	registry := &v1alpha1.Registry{
		// 		ObjectMeta: metav1.ObjectMeta{
		// 			Name:      registryName,
		// 			Namespace: cfg.Namespace(),
		// 		},
		// 		Spec: v1alpha1.RegistrySpec{
		// 			URI:          registryURI,
		// 			Repositories: []string{registryRepository},
		// 		},
		// 	}
		// 	err := client.Resources(cfg.Namespace()).Create(ctx, registry)
		// 	require.NoError(t, err)
		// 	return ctx
		// }).
		// Assess("SPDX SBOM is created with expected content", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	client := cfg.Client()

		// 	assert.Eventually(t, func() bool {
		// 		sboms := &storagev1alpha1.SBOMList{}
		// 		if err := client.Resources(cfg.Namespace()).List(ctx, sboms); err != nil {
		// 			return false
		// 		}
		// 		for _, item := range sboms.Items {
		// 			if EqualReference(item.Spec.ImageMetadata, registryURI, registryRepository, golangAlpineTag) {
		// 				sbom = item
		// 				return true
		// 			}
		// 		}
		// 		return false
		// 	}, pollTimeout, pollInterval, "SBOM CR was not generated or no matching image was found")

		// 	spdxData, err := os.ReadFile(spdxPath)
		// 	require.NoError(t, err)

		// 	expectedSPDX := &spdx.Document{}
		// 	err = json.Unmarshal(spdxData, expectedSPDX)
		// 	require.NoError(t, err)

		// 	generatedSPDX := &spdx.Document{}
		// 	err = json.Unmarshal(sbom.Spec.SPDX.Raw, generatedSPDX)
		// 	require.NoError(t, err)

		// 	// Filter out "DocumentNamespace" and any field named "AnnotationDate" or "Created" regardless of nesting,
		// 	// since they contain timestamps and are not deterministic.
		// 	filter := cmp.FilterPath(func(path cmp.Path) bool {
		// 		lastField := path.Last().String()
		// 		return lastField == ".DocumentNamespace" || lastField == ".AnnotationDate" || lastField == ".Created"
		// 	}, cmp.Ignore())
		// 	diff := cmp.Diff(expectedSPDX, generatedSPDX, filter, cmpopts.IgnoreUnexported(spdx.Package{}))
		// 	assert.Empty(t, diff)
		// 	return ctx
		// }).
		// Assess("Vulnerability Report is created with expected content", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	client := cfg.Client()
		// 	var vulnReport storagev1alpha1.VulnerabilityReport

		// 	assert.Eventually(t, func() bool {
		// 		vulnReports := &storagev1alpha1.VulnerabilityReportList{}
		// 		if err := client.Resources(cfg.Namespace()).List(ctx, vulnReports); err != nil {
		// 			return false
		// 		}
		// 		for _, item := range vulnReports.Items {
		// 			if EqualReference(item.Spec.ImageMetadata, registryURI, registryRepository, golangAlpineTag) {
		// 				vulnReport = item
		// 				return true
		// 			}
		// 		}
		// 		return false
		// 	}, pollTimeout, pollInterval, "VulnerabilityReport CR was not generated or no matching image was found")

		// 	generatedReport := &sarif.Report{}
		// 	err := json.Unmarshal(vulnReport.Spec.SARIF.Raw, generatedReport)
		// 	require.NoError(t, err)

		// 	assert.Equal(t, sbom.GetImageMetadata(), vulnReport.GetImageMetadata())
		// 	assert.Equal(t, sbom.UID, vulnReport.GetOwnerReferences()[0].UID)

		// 	reportData, err := os.ReadFile(reportPath)
		// 	require.NoError(t, err)

		// 	expectedReport := &sarif.Report{}
		// 	err = json.Unmarshal(reportData, expectedReport)
		// 	require.NoError(t, err)

		// 	// Filter out fields containing the file path from the comparison
		// 	filter := cmp.FilterPath(func(path cmp.Path) bool {
		// 		lastField := path.Last().String()
		// 		return lastField == ".URI" || lastField == ".Text"
		// 	}, cmp.Comparer(func(a, b *string) bool {
		// 		if strings.Contains(*a, ".json") && strings.Contains(*b, ".json") {
		// 			return true
		// 		}

		// 		return cmp.Equal(a, b)
		// 	}))
		// 	diff := cmp.Diff(expectedReport, generatedReport, filter)

		// 	assert.Empty(t, diff)
		// 	return ctx
		// }).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// client := cfg.Client()

			// registry := &v1alpha1.Registry{
			// 	ObjectMeta: metav1.ObjectMeta{
			// 		Name:      registryName,
			// 		Namespace: cfg.Namespace(),
			// 	},
			// }
			// err := client.Resources(cfg.Namespace()).Delete(ctx, registry)
			// if err != nil {
			// 	t.Fatal(err)
			// }

			// // Ensure that the SBOM and VulnerabilityReport CRs are deleted after the Registry CR is deleted
			// assert.Eventually(t, func() bool {
			// 	images := &storagev1alpha1.ImageList{}
			// 	if err := client.Resources(cfg.Namespace()).List(ctx, images); err != nil {
			// 		return true
			// 	}

			// 	imageDeleted := true
			// 	for _, item := range images.Items {
			// 		if EqualReference(item.Spec.ImageMetadata, registryURI, registryRepository, golangAlpineTag) {
			// 			imageDeleted = false
			// 			break
			// 		}
			// 	}
			// 	return imageDeleted
			// }, pollTimeout, pollInterval, "Image CR was not deleted after Registry CR was deleted")

			// assert.Eventually(t, func() bool {
			// 	sboms := &storagev1alpha1.SBOMList{}
			// 	if err := client.Resources(cfg.Namespace()).List(ctx, sboms); err != nil {
			// 		return true
			// 	}

			// 	sbomDeleted := true
			// 	for _, item := range sboms.Items {
			// 		if EqualReference(item.Spec.ImageMetadata, registryURI, registryRepository, golangAlpineTag) {
			// 			sbomDeleted = false
			// 			break
			// 		}
			// 	}
			// 	return sbomDeleted
			// }, pollTimeout, pollInterval, "SBOM CR was not deleted after Registry CR was deleted")

			// assert.Eventually(t, func() bool {
			// 	vulnReports := &storagev1alpha1.VulnerabilityReportList{}
			// 	if err := client.Resources(cfg.Namespace()).List(ctx, vulnReports); err != nil {
			// 		return true
			// 	}

			// 	vulnReportDeleted := true
			// 	for _, item := range vulnReports.Items {
			// 		if EqualReference(item.Spec.ImageMetadata, registryURI, registryRepository, golangAlpineTag) {
			// 			vulnReportDeleted = false
			// 			break
			// 		}
			// 	}
			// 	return vulnReportDeleted
			// }, pollTimeout, pollInterval, "VulnerabilityReport CR was not deleted after Registry CR was deleted")

			manager := helm.New(cfg.KubeconfigFile())
			err := manager.RunUninstall(
				helm.WithName(releaseName),
				helm.WithNamespace(cfg.Namespace()),
			)
			assert.NoError(t, err, "sbombastic helm chart is not deleted correctly")
			return ctx
		})

	testenv.Test(t, f.Feature())
}
