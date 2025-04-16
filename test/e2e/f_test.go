package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/owenrumney/go-sarif/v2/sarif"
	"github.com/rancher/sbombastic/api/storage/v1alpha1"
	"github.com/spdx/tools-golang/spdx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

var (
	spdxPath   = filepath.Join("..", "fixtures", "golang-1.12-alpine.spdx.json")
	reportPath = filepath.Join("..", "fixtures", "golang-1.12-alpine.sarif.json")
)

func getExpectedSBOM(t *testing.T, sbomPath string) *spdx.Document {
	spdxData, err := os.ReadFile(sbomPath)
	require.NoError(t, err)

	expectedSPDX := &spdx.Document{}
	err = json.Unmarshal(spdxData, expectedSPDX)
	require.NoError(t, err)

	return expectedSPDX
}

func getExpectedVulnerabilityReport(t *testing.T, reportPath string) *sarif.Report {
	reportData, err := os.ReadFile(reportPath)
	require.NoError(t, err)

	expectedReport := &sarif.Report{}
	err = json.Unmarshal(reportData, expectedReport)
	require.NoError(t, err)

	return expectedReport
}

func TestRegistryCreation(t *testing.T) {
	time.Sleep(10 * time.Second)
	log.Println("TestRegistryCreation runningg")
	crName := "84dd0564eb2f958009581262462e413e1a0ea6f632250897c8f4f17d6cae15af"

	f := features.New("Registry CR").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client := cfg.Client()
			v1alpha1.AddToScheme(client.Resources("default").GetScheme())
			return ctx
		}).
		Assess("SPDX SBOM is created with expected content", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			fmt.Println("Registry is created", crName)
			client := cfg.Client()

			var sbom v1alpha1.SBOM
			if err := client.Resources().Get(ctx, crName, "default", &sbom); err != nil {
				t.Fatal(err)
			}

			fmt.Println("sbom", sbom.Spec.ImageMetadata.RegistryURI)

			spdxPath := filepath.Join("..", "fixtures", "golang-1.12-alpine.spdx.json")
			expectedSPDX := getExpectedSBOM(t, spdxPath)

			generatedSPDX := &spdx.Document{}
			err := json.Unmarshal(sbom.Spec.SPDX.Raw, generatedSPDX)
			require.NoError(t, err)

			// Filter out "DocumentNamespace" and any field named "AnnotationDate" or "Created" regardless of nesting,
			// since they contain timestamps and are not deterministic.
			filter := cmp.FilterPath(func(path cmp.Path) bool {
				lastField := path.Last().String()
				return lastField == ".DocumentNamespace" || lastField == ".AnnotationDate" || lastField == ".Created" || lastField == ".DocumentName"
			}, cmp.Ignore())
			diff := cmp.Diff(expectedSPDX, generatedSPDX, filter, cmpopts.IgnoreUnexported(spdx.Package{}))
			assert.Empty(t, diff)
			return ctx
		}).
		Assess("Vulnerability Report is created with expected content", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client := cfg.Client()

			var vulnReport v1alpha1.VulnerabilityReport // Import your SBOM CR type
			if err := client.Resources().Get(ctx, crName, "default", &vulnReport); err != nil {
				t.Fatal(err)
			}
			fmt.Println("sbom", vulnReport.Spec.ImageMetadata.RegistryURI)

			// spdxPath := filepath.Join("..", "fixtures", "golang-1.12-alpine.spdx.json")
			// spdxData, err := os.ReadFile(spdxPath)
			// require.NoError(t, err)

			// expectedSPDX := &spdx.Document{}
			// err = json.Unmarshal(spdxData, expectedSPDX)
			// require.NoError(t, err)

			// generatedSPDX := &spdx.Document{}
			// err = json.Unmarshal(sbom.Spec.SPDX.Raw, generatedSPDX)
			// require.NoError(t, err)

			// // Filter out "DocumentNamespace" and any field named "AnnotationDate" or "Created" regardless of nesting,
			// // since they contain timestamps and are not deterministic.
			// filter := cmp.FilterPath(func(path cmp.Path) bool {
			// 	lastField := path.Last().String()
			// 	return lastField == ".DocumentNamespace" || lastField == ".AnnotationDate" || lastField == ".Created" || lastField == ".DocumentName"
			// }, cmp.Ignore())
			// diff := cmp.Diff(expectedSPDX, generatedSPDX, filter, cmpopts.IgnoreUnexported(spdx.Package{}))
			// assert.Empty(t, diff)
			return ctx
		})

	testenv.Test(t, f.Feature())
}
