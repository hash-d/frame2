package skupperexecute

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	frame2 "github.com/hash-d/frame2/pkg"
	"github.com/hash-d/frame2/pkg/execute"
	"github.com/hash-d/frame2/pkg/frames/f2k8s"
)

type SkupperManifestContentImage struct {
	Name       string
	Repository string
}

type SkupperManifestContent struct {
	Images    []SkupperManifestContentImage
	Variables map[string]string
}

// SkupperManifest returns the content of the requested manifest.json
// file as a SkupperManifestContent.
//
// If data is provided in Expected, it also checks that all of its items
// match the actual file's contents.
//
// The check is only on the actual images, as strings, including the tags.
// It has no intelligence to add or remove :latest from a tag, for example.
//
// The Repository field is not used in this verification.
//
// Note this frame never accesses a cluster; the verification is on the
// manifest.json (provided or generated from skupper binary); the comparison
// is against a SkupperManifestContent generated elsewhere and provided to it
type SkupperManifest struct {

	// Path to the manifest.json file; if not provided, it will be
	// searched first on the test root, then on the source root.
	//
	// Starting with 1.5, if the Path is empty, the command
	// `skupper version manifest` will be executed to generate a
	// manifest file, instead of searching for it.
	Path string

	SkipComparison bool

	Expected SkupperManifestContent
	Result   *SkupperManifestContent

	frame2.DefaultRunDealer
	frame2.Log
	execute.SkupperVersionerDefault
}

// TODO: replace this by f2k8s.Namespace
func (m SkupperManifest) GetNamespace() string {
	return ""
}

// Executes skupper version manifest in the given directory,
// and returns the path to it.
//
// If it already exists on the given directory, it will not be
// re-generated, and its path is returned with no further action
func (m SkupperManifest) getManifestPath(dir string) string {
	manifPath := filepath.Join(dir, "manifest.json")
	if _, err := os.Stat(manifPath); err == nil {
		return manifPath
	} else {
		// Execute skupper version manifest on the tmpdir, to
		// generate manifest.json
		phase := frame2.Phase{
			Runner: m.Runner,
			MainSteps: []frame2.Step{
				{
					Modify: &CliSkupper{
						Args: []string{"version", "manifest"},
						Cmd: execute.Cmd{
							Cmd: exec.Cmd{
								Dir: dir,
							},
						},
					},
				},
			},
		}
		err := phase.Run()
		if err == nil {
			return manifPath
		} else {
			return ""
		}
	}
}

func (m SkupperManifest) Validate() error {

	if m.SkipComparison {
		m.Log.Printf("SkupperManifest: Skipping comparison per configuration")
		return nil
	}

	manifestPath := m.Path
	if manifestPath == "" {
		versions := []string{"1.4", "1.5"}
		target := m.WhichSkupperVersion(versions)
		switch target {
		case "1.5", "":
			// starting with 1.5, we get the path from skupper version manifest
			// (if not explicited on the struct)
			tmpdir := m.Runner.T.TempDir()
			manifestPath = m.getManifestPath(tmpdir)
		case "1.4":
			alternates := []string{
				"./manifest.json",
				path.Join(frame2.SourceRoot(), "manifest.json"),
			}
			for _, alternate := range alternates {
				if _, err := os.Stat(alternate); err == nil {
					manifestPath = alternate
					break
				}
			}
		default:
			panic("This should not have happened")
		}
	}
	if manifestPath == "" {
		return fmt.Errorf("SkupperManifest: no path to manifest.json found, and none found on default locations")
	}

	var manifestBytes []byte
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("SkupperManifest: could not read file %q: %w", manifestPath, err)
	}

	m.Result = &SkupperManifestContent{}
	err = json.Unmarshal(manifestBytes, m.Result)
	if err != nil {
		return fmt.Errorf("SkupperManifest: could not unmarshal %q: %w", manifestPath, err)
	}

	// Images verification
	for _, expected := range m.Expected.Images {
		// TODO This is not good; if a variable is set, but it was ignored, the first for will (correctly)
		// consider it as not found, but the second one may (incorrectly) match it with the hardcoded image
		//
		// To fix this, due to the way the manifest is structured, we'll need to extract frmo the image the
		// information of what it refers to, and get intelligence to check them individually
		var found bool
		for varName, mapped := range m.Result.Variables {
			if expected.Name == mapped {
				m.Log.Printf("Deployment image %q matched on %q via variable %q", mapped, manifestPath, varName)
				found = true
				break
			}
		}
		for _, actual := range m.Result.Images {
			if expected.Name == actual.Name {
				m.Log.Printf("Deployment image %q matched on %q (repo %q)", actual.Name, manifestPath, actual.Repository)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Deployment image %q did not match any items from %q", expected.Name, manifestPath)
		}

	}

	return nil
}

// This will get the deployment information on Skupper as currently
// deployed on the namespace (its images), and then compare it against
// the manifest, to ensure it matches.
//
// After an install or an upgrade, you may need to retry this action
// a few times, to ensure the deployment stabilized
type ManifestMatchesDeployment struct {
	Path      string
	Namespace *f2k8s.Namespace

	Ctx context.Context

	frame2.DefaultRunDealer
	frame2.Log
	execute.SkupperVersionerDefault
}

// TODO: replace this by f2k8s.Namespace
func (m ManifestMatchesDeployment) GetNamespace() string {
	return m.Namespace.GetNamespaceName()
}

func (m ManifestMatchesDeployment) Validate() error {
	var err error

	skupperInfo := SkupperInfo{
		Namespace: m.Namespace,
		Ctx:       m.Ctx,
	}
	getInfoPhase := frame2.Phase{
		Runner: m.Runner,
		Doc:    "Get the Skupper deployment info",
		MainSteps: []frame2.Step{
			{
				Validator: &skupperInfo,
			},
		},
	}
	err = getInfoPhase.Run()
	if err != nil {
		return fmt.Errorf("failed to acquire skupper info: %w", err)
	}

	checkManifestPhase := frame2.Phase{
		Runner: m.Runner,
		Doc:    "Compare deployed Skupper images to the manifest",
		MainSteps: []frame2.Step{
			{
				Doc: "Compare Deployments",
				Validator: &SkupperManifest{
					Expected: skupperInfo.Result.Images,
				},
			},
			{
				Doc: "Compare Pods",
				Validator: &SkupperManifest{
					Expected: skupperInfo.Result.PodImages,
				},
			},
		},
	}
	return checkManifestPhase.Run()
}
