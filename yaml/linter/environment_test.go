package linter

import (
	"testing"

	"github.com/open-beagle/bdpulse-yaml/yaml"
)

func TestCheckPipelineAcceptsDirectSecretSetting(t *testing.T) {
	pipeline := &yaml.Pipeline{
		Name: "test",
		Steps: []*yaml.Container{{
			Name:  "publish",
			Image: "alpine",
			Settings: map[string]*yaml.Parameter{
				"password": {Value: "${{ secrets.REGISTRY_PASSWORD }}"},
			},
		}},
	}
	if err := checkPipeline(pipeline, true); err != nil {
		t.Fatal(err)
	}
}

func TestCheckPipelineRejectsSecretInImage(t *testing.T) {
	pipeline := &yaml.Pipeline{
		Name: "test",
		Steps: []*yaml.Container{{
			Name:  "publish",
			Image: "registry.example/${{ secrets.REGISTRY_PASSWORD }}",
		}},
	}
	if err := checkPipeline(pipeline, true); err == nil {
		t.Fatal("expected secret expression in image to fail")
	}
}

func TestCheckPipelineRejectsSecretEnvironmentInImage(t *testing.T) {
	pipeline := &yaml.Pipeline{
		Name: "test",
		Environment: map[string]*yaml.Variable{
			"PRIVATE_TAG": {Value: "${{ secrets.PRIVATE_TAG }}"},
		},
		Steps: []*yaml.Container{{
			Name:  "publish",
			Image: "registry.example:${{ env.PRIVATE_TAG }}",
		}},
	}
	if err := checkPipeline(pipeline, true); err == nil {
		t.Fatal("expected secret environment in image to fail")
	}
}
