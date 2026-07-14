package compiler

import (
	"testing"

	"github.com/open-beagle/bdpulse-yaml/yaml"
)

func TestCompilePipelineEnvironment(t *testing.T) {
	pipeline := &yaml.Pipeline{
		Name: "test",
		Environment: map[string]*yaml.Variable{
			"GO_VERSION":    {Value: "1.26"},
			"REGISTRY_USER": {Value: "${{ secrets.REGISTRY_USER }}"},
		},
		Steps: []*yaml.Container{{
			Name:  "build",
			Image: "golang:${{ env.GO_VERSION }}",
			Environment: map[string]*yaml.Variable{
				"GO_VERSION": {Value: "1.27"},
			},
			Settings: map[string]*yaml.Parameter{
				"version":  {Value: "${{ env.GO_VERSION }}"},
				"password": {Value: "${{ secrets.REGISTRY_PASSWORD }}"},
			},
			Commands: []string{"echo ${{ secrets.RELEASE_TOKEN }}"},
		}},
	}

	spec := new(Compiler).Compile(pipeline)
	if got, want := len(spec.Steps), 2; got != want {
		t.Fatalf("steps = %d, want %d", got, want)
	}
	clone, step := spec.Steps[0], spec.Steps[1]
	if _, ok := clone.Envs["GO_VERSION"]; ok {
		t.Fatal("clone inherited pipeline environment")
	}
	if got, want := step.Envs["GO_VERSION"], "1.27"; got != want {
		t.Fatalf("GO_VERSION = %q, want %q", got, want)
	}
	if got, want := step.Envs["PLUGIN_VERSION"], "1.27"; got != want {
		t.Fatalf("PLUGIN_VERSION = %q, want %q", got, want)
	}
	if got, want := step.Docker.Image, "docker.io/library/golang:1.27"; got != want {
		t.Fatalf("image = %q, want %q", got, want)
	}
	secrets := map[string]string{}
	for _, secret := range step.Secrets {
		secrets[secret.Env] = secret.Name
	}
	if got, want := secrets["REGISTRY_USER"], "REGISTRY_USER"; got != want {
		t.Fatalf("REGISTRY_USER secret = %q, want %q", got, want)
	}
	if got, want := secrets["PLUGIN_PASSWORD"], "REGISTRY_PASSWORD"; got != want {
		t.Fatalf("PLUGIN_PASSWORD secret = %q, want %q", got, want)
	}
	if got, want := secrets["AWE_SECRET_RELEASE_TOKEN"], "RELEASE_TOKEN"; got != want {
		t.Fatalf("command secret = %q, want %q", got, want)
	}
}
