package e2e

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/owenrumney/go-sarif/v2/sarif"
	storagev1alpha1 "github.com/rancher/sbombastic/api/storage/v1alpha1"
	v1alpha1 "github.com/rancher/sbombastic/api/v1alpha1"
	"github.com/spdx/tools-golang/spdx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestRegistryCreation(t *testing.T) {
	spdxPath := filepath.Join("..", "fixtures", "golang-1.12-alpine.spdx.json")
	reportPath := filepath.Join("..", "fixtures", "golang-1.12-alpine.sarif.json")
	golangAlpineTag := "1.12-alpine"
	registryName := "test-registry"
	pollInterval := 1 * time.Second
	pollTimeout := 60 * time.Second
	var sbom storagev1alpha1.SBOM

	f := features.New("Registry CR Creation test").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client := cfg.Client()
			storagev1alpha1.AddToScheme(client.Resources(cfg.Namespace()).GetScheme())
			v1alpha1.AddToScheme(client.Resources(cfg.Namespace()).GetScheme())
			return ctx
		}).
		Assess("Create Registry CR", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client := cfg.Client()
			registry := &v1alpha1.Registry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      registryName,
					Namespace: cfg.Namespace(),
				},
				Spec: v1alpha1.RegistrySpec{
					URI:          "ghcr.io",
					Repositories: []string{"rancher-sandbox/sbombastic/test-assets/golang"},
				},
			}
			err := client.Resources(cfg.Namespace()).Create(ctx, registry)
			require.NoError(t, err)
			return ctx
		}).
		Assess("SPDX SBOM is created with expected content", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client := cfg.Client()

			assert.Eventually(t, func() bool {
				sboms := &storagev1alpha1.SBOMList{}
				if err := client.Resources(cfg.Namespace()).List(ctx, sboms); err != nil {
					return false
				}
				for _, item := range sboms.Items {
					if item.Spec.ImageMetadata.Tag == golangAlpineTag {
						sbom = item
						return true
					}
				}
				return false
			}, pollTimeout, pollInterval, "SBOM CR was not generated or no matching image was found")

			spdxData, err := os.ReadFile(spdxPath)
			require.NoError(t, err)

			expectedSPDX := &spdx.Document{}
			err = json.Unmarshal(spdxData, expectedSPDX)
			require.NoError(t, err)

			generatedSPDX := &spdx.Document{}
			err = json.Unmarshal(sbom.Spec.SPDX.Raw, generatedSPDX)
			require.NoError(t, err)

			// Filter out "DocumentNamespace" and any field named "AnnotationDate" or "Created" regardless of nesting,
			// since they contain timestamps and are not deterministic.
			filter := cmp.FilterPath(func(path cmp.Path) bool {
				lastField := path.Last().String()
				return lastField == ".DocumentNamespace" || lastField == ".AnnotationDate" || lastField == ".Created"
			}, cmp.Ignore())
			diff := cmp.Diff(expectedSPDX, generatedSPDX, filter, cmpopts.IgnoreUnexported(spdx.Package{}))
			assert.Empty(t, diff)
			return ctx
		}).
		Assess("Vulnerability Report is created with expected content", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client := cfg.Client()
			var vulnReport storagev1alpha1.VulnerabilityReport

			assert.Eventually(t, func() bool {
				vulnReports := &storagev1alpha1.VulnerabilityReportList{}
				if err := client.Resources(cfg.Namespace()).List(ctx, vulnReports); err != nil {
					return false
				}
				for _, item := range vulnReports.Items {
					if item.Spec.ImageMetadata.Tag == golangAlpineTag {
						vulnReport = item
						return true
					}
				}
				return false
			}, pollTimeout, pollInterval, "VulnerabilityReport CR was not generated or no matching image was found")

			generatedReport := &sarif.Report{}
			err := json.Unmarshal(vulnReport.Spec.SARIF.Raw, generatedReport)
			require.NoError(t, err)

			assert.Equal(t, sbom.GetImageMetadata(), vulnReport.GetImageMetadata())
			assert.Equal(t, sbom.UID, vulnReport.GetOwnerReferences()[0].UID)

			reportData, err := os.ReadFile(reportPath)
			require.NoError(t, err)

			expectedReport := &sarif.Report{}
			err = json.Unmarshal(reportData, expectedReport)
			require.NoError(t, err)

			// Filter out fields containing the file path from the comparison
			filter := cmp.FilterPath(func(path cmp.Path) bool {
				lastField := path.Last().String()
				return lastField == ".URI" || lastField == ".Text"
			}, cmp.Comparer(func(a, b *string) bool {
				if strings.Contains(*a, ".json") && strings.Contains(*b, ".json") {
					return true
				}

				return cmp.Equal(a, b)
			}))
			diff := cmp.Diff(expectedReport, generatedReport, filter)

			assert.Empty(t, diff)
			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client := cfg.Client()

			registry := &v1alpha1.Registry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      registryName,
					Namespace: cfg.Namespace(),
				},
			}
			err := client.Resources(cfg.Namespace()).Delete(ctx, registry)
			if err != nil {
				t.Fatal(err)
			}

			// Ensure that the SBOM and VulnerabilityReport CRs are deleted after the Registry CR is deleted
			assert.Eventually(t, func() bool {
				sboms := &storagev1alpha1.SBOMList{}
				if err := client.Resources(cfg.Namespace()).List(ctx, sboms); err != nil {
					return true
				}

				sbomDeleted := true
				for _, item := range sboms.Items {
					if item.Spec.ImageMetadata.Tag == golangAlpineTag {
						sbomDeleted = false
						break
					}
				}
				return sbomDeleted
			}, pollTimeout, pollInterval, "SBOM CR was not deleted after Registry CR was deleted")

			assert.Eventually(t, func() bool {
				vulnReports := &storagev1alpha1.VulnerabilityReportList{}
				if err := client.Resources(cfg.Namespace()).List(ctx, vulnReports); err != nil {
					return true
				}

				vulnReportDeleted := true
				for _, item := range vulnReports.Items {
					if item.Spec.ImageMetadata.Tag == golangAlpineTag {
						vulnReportDeleted = false
						vulnReportDeleted = false
						break
					}
				}
				return vulnReportDeleted
			}, pollTimeout, pollInterval, "VulnerabilityReport CR was not deleted after Registry CR was deleted")
			return ctx
		})

	testenv.Test(t, f.Feature())
}
