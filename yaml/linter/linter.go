package linter

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/open-beagle/bdpulse-yaml/yaml"
)

var os = map[string]struct{}{
	"linux":   struct{}{},
	"windows": struct{}{},
}

var arch = map[string]struct{}{
	"arm":     struct{}{},
	"arm64":   struct{}{},
	"amd64":   struct{}{},
	"ppc64le": struct{}{},
	"s390x":   struct{}{},
}

var environmentExpression = regexp.MustCompile(`\$\{\{\s*env\.([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`)

var environmentKey = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

var secretExpression = regexp.MustCompile(`\$\{\{\s*secrets\.([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`)

var secretExpressionExact = regexp.MustCompile(`^\$\{\{\s*secrets\.[A-Za-z_][A-Za-z0-9_]*\s*\}\}$`)

// ErrDuplicateStepName is returned when two Pipeline steps
// have the same name.
var ErrDuplicateStepName = errors.New("linter: duplicate step names")

// ErrMissingDependency is returned when a Pipeline step
// defines dependencies that are invlid or unknown.
var ErrMissingDependency = errors.New("linter: invalid or unknown step dependency")

// ErrCyclicalDependency is returned when a Pipeline step
// defines a cyclical dependency, which would result in an
// infinite execution loop.
var ErrCyclicalDependency = errors.New("linter: cyclical step dependency detected")

// Lint performs lint operations for a resource.
func Lint(resource yaml.Resource, trusted bool) error {
	switch v := resource.(type) {
	case *yaml.Cron:
		return v.Validate()
	case *yaml.Pipeline:
		return checkPipeline(v, true)
	case *yaml.Secret:
		return v.Validate()
	case *yaml.Registry:
		return v.Validate()
	case *yaml.Signature:
		return v.Validate()
	default:
		return nil
	}
}

func checkPipeline(pipeline *yaml.Pipeline, trusted bool) error {
	if err := checkEnvironment(pipeline.Environment); err != nil {
		return err
	}
	err := checkVolumes(pipeline, trusted)
	if err != nil {
		return err
	}
	err = checkPlatform(pipeline.Platform)
	if err != nil {
		return err
	}
	names := map[string]struct{}{}
	if pipeline.Clone.Disable == false {
		names["clone"] = struct{}{}
	}
	for _, container := range pipeline.Steps {
		_, ok := names[container.Name]
		if ok {
			return ErrDuplicateStepName
		}
		names[container.Name] = struct{}{}

		err := checkContainer(container, pipeline.Environment, trusted)
		if err != nil {
			return err
		}

		err = checkDeps(container, names)
		if err != nil {
			return err
		}
	}
	for _, container := range pipeline.Services {
		_, ok := names[container.Name]
		if ok {
			return ErrDuplicateStepName
		}
		names[container.Name] = struct{}{}

		err := checkContainer(container, pipeline.Environment, trusted)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkPlatform(platform yaml.Platform) error {
	if v := platform.OS; v != "" {
		_, ok := os[v]
		if !ok {
			return fmt.Errorf("linter: unsupported os: %s", v)
		}
	}
	if v := platform.Arch; v != "" {
		_, ok := arch[v]
		if !ok {
			return fmt.Errorf("linter: unsupported architecture: %s", v)
		}
	}
	return nil
}

func checkContainer(container *yaml.Container, inherited map[string]*yaml.Variable, trusted bool) error {
	if err := checkEnvironment(container.Environment); err != nil {
		return err
	}
	if err := checkEnvironmentExpressions(container, mergeEnvironment(inherited, container.Environment)); err != nil {
		return err
	}
	err := checkPorts(container.Ports, trusted)
	if err != nil {
		return err
	}
	if container.Build == nil && container.Image == "" {
		return errors.New("linter: invalid or missing image")
	}
	if container.Build != nil && container.Build.Image == "" {
		return errors.New("linter: invalid or missing build image")
	}
	if container.Name == "" {
		return errors.New("linter: invalid or missing name")
	}
	if trusted == false && container.Privileged {
		return errors.New("linter: untrusted repositories cannot enable privileged mode")
	}
	if trusted == false && len(container.Devices) > 0 {
		return errors.New("linter: untrusted repositories cannot mount devices")
	}
	if trusted == false && len(container.DNS) > 0 {
		return errors.New("linter: untrusted repositories cannot configure dns")
	}
	if trusted == false && len(container.DNSSearch) > 0 {
		return errors.New("linter: untrusted repositories cannot configure dns_search")
	}
	if trusted == false && len(container.ExtraHosts) > 0 {
		return errors.New("linter: untrusted repositories cannot configure extra_hosts")
	}
	if trusted == false && len(container.Network) > 0 {
		return errors.New("linter: untrusted repositories cannot configure network_mode")
	}
	for _, mount := range container.Volumes {
		switch mount.Name {
		case "workspace", "_workspace", "_docker_socket":
			return fmt.Errorf("linter: invalid volume name: %s", mount.Name)
		}
	}
	return nil
}

func mergeEnvironment(inherited, local map[string]*yaml.Variable) map[string]*yaml.Variable {
	if len(inherited) == 0 && len(local) == 0 {
		return nil
	}
	merged := make(map[string]*yaml.Variable, len(inherited)+len(local))
	for key, value := range inherited {
		merged[key] = value
	}
	for key, value := range local {
		merged[key] = value
	}
	return merged
}

func checkEnvironment(environment map[string]*yaml.Variable) error {
	for key, value := range environment {
		if !environmentKey.MatchString(key) {
			return fmt.Errorf("linter: invalid environment variable name: %s", key)
		}
		if isProtectedEnvironment(key) {
			return fmt.Errorf("linter: protected environment variable: %s", key)
		}
		if value == nil {
			return fmt.Errorf("linter: invalid environment variable: %s", key)
		}
		if secretExpression.MatchString(value.Value) && !isSecretExpression(value.Value) {
			return fmt.Errorf("linter: secret expression must be the complete environment value: %s", key)
		}
	}
	return nil
}

func checkEnvironmentExpressions(container *yaml.Container, environment map[string]*yaml.Variable) error {
	if secretExpression.MatchString(container.Image) {
		return errors.New("linter: secret expression is not allowed in image")
	}
	values := []string{container.Image}
	for _, parameter := range container.Settings {
		if parameter == nil {
			continue
		}
		if value, ok := parameter.Value.(string); ok {
			if secretExpression.MatchString(value) && !isSecretExpression(value) {
				return errors.New("linter: secret expression must be the complete setting value")
			}
			values = append(values, value)
		}
	}
	for _, value := range values {
		for _, match := range environmentExpression.FindAllStringSubmatch(value, -1) {
			variable, ok := environment[match[1]]
			if ok && variable != nil && (variable.Secret != "" || isSecretExpression(variable.Value)) {
				return fmt.Errorf("linter: environment expression cannot reference secret: %s", match[1])
			}
		}
	}
	return nil
}

func isSecretExpression(value string) bool {
	return secretExpressionExact.MatchString(strings.TrimSpace(value))
}

func isProtectedEnvironment(key string) bool {
	for _, prefix := range []string{"DRONE_", "CI_", "AWE_", "PLUGIN_"} {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

func checkPorts(ports []*yaml.Port, trusted bool) error {
	for _, port := range ports {
		err := checkPort(port, trusted)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkPort(port *yaml.Port, trusted bool) error {
	if trusted == false && port.Host != 0 {
		return errors.New("linter: untrusted repositories cannot map to a host port")
	}
	return nil
}

func checkVolumes(pipeline *yaml.Pipeline, trusted bool) error {
	for _, volume := range pipeline.Volumes {
		if volume.EmptyDir != nil {
			err := checkEmptyDirVolume(volume.EmptyDir, trusted)
			if err != nil {
				return err
			}
		}
		if volume.HostPath != nil {
			err := checkHostPathVolume(volume.HostPath, trusted)
			if err != nil {
				return err
			}
		}
		switch volume.Name {
		case "workspace", "_workspace", "_docker_socket":
			return fmt.Errorf("linter: invalid volume name: %s", volume.Name)
		}
	}
	return nil
}

func checkHostPathVolume(volume *yaml.VolumeHostPath, trusted bool) error {
	if trusted == false {
		return errors.New("linter: untrusted repositories cannot mount host volumes")
	}
	return nil
}

func checkEmptyDirVolume(volume *yaml.VolumeEmptyDir, trusted bool) error {
	if trusted == false && volume.Medium == "memory" {
		return errors.New("linter: untrusted repositories cannot mount in-memory volumes")
	}
	return nil
}

func checkDeps(container *yaml.Container, deps map[string]struct{}) error {
	for _, dep := range container.DependsOn {
		_, ok := deps[dep]
		if !ok {
			return ErrMissingDependency
		}
		if container.Name == dep {
			return ErrCyclicalDependency
		}
	}
	return nil
}
