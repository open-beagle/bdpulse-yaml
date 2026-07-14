package compiler

import (
	"regexp"
	"strings"

	"github.com/open-beagle/bdpulse-runtime/engine"
	"github.com/open-beagle/bdpulse-yaml/yaml"
)

var envExpression = regexp.MustCompile(`\$\{\{\s*env\.([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`)

var secretExpression = regexp.MustCompile(`^\$\{\{\s*secrets\.([A-Za-z_][A-Za-z0-9_]*)\s*\}\}$`)

var secretExpressionAny = regexp.MustCompile(`\$\{\{\s*secrets\.([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`)

// resolveContainer 将根环境应用到业务容器。
// clone 有意绕过此函数，避免业务变量和 Secret 暴露给内部克隆容器。
func resolveContainer(src *yaml.Container, root map[string]*yaml.Variable) *yaml.Container {
	dst := *src
	dst.Environment = mergeEnvironment(root, src.Environment)
	dst.Image = resolveEnvironmentExpressions(src.Image, dst.Environment)
	if len(src.Commands) > 0 && dst.Environment == nil {
		dst.Environment = map[string]*yaml.Variable{}
	}
	dst.Commands = resolveSecretCommands(src.Commands, dst.Environment)

	if len(src.Settings) == 0 {
		return &dst
	}

	dst.Settings = make(map[string]*yaml.Parameter, len(src.Settings))
	for key, parameter := range src.Settings {
		if parameter == nil {
			continue
		}
		copy := *parameter
		if value, ok := copy.Value.(string); ok {
			copy.Value = resolveEnvironmentExpressions(value, dst.Environment)
		}
		dst.Settings[key] = &copy
	}
	return &dst
}

func resolveSecretCommands(commands []string, environment map[string]*yaml.Variable) []string {
	if len(commands) == 0 {
		return nil
	}
	resolved := make([]string, len(commands))
	for i, command := range commands {
		resolved[i] = secretExpressionAny.ReplaceAllStringFunc(command, func(expression string) string {
			parts := secretExpressionAny.FindStringSubmatch(expression)
			key := "AWE_SECRET_" + parts[1]
			environment[key] = &yaml.Variable{Secret: parts[1]}
			return "${" + key + "}"
		})
	}
	return resolved
}

func mergeEnvironment(root, local map[string]*yaml.Variable) map[string]*yaml.Variable {
	if len(root) == 0 && len(local) == 0 {
		return nil
	}
	merged := make(map[string]*yaml.Variable, len(root)+len(local))
	for key, value := range root {
		merged[key] = value
	}
	for key, value := range local {
		merged[key] = value
	}
	return merged
}

func markEnvironmentOverrides(step *engine.Step, environment map[string]*yaml.Variable) {
	if len(environment) == 0 {
		return
	}
	step.EnvOverrides = make(map[string]bool, len(environment)+2)
	for key := range environment {
		step.EnvOverrides[key] = true
	}
	if _, ok := step.Envs[engine.EnvironmentFileVariable]; ok {
		step.EnvOverrides[engine.EnvironmentFileVariable] = true
		step.EnvOverrides[engine.DroneEnvironmentFileVariable] = true
	}
}

func secretName(value string) (string, bool) {
	parts := secretExpression.FindStringSubmatch(strings.TrimSpace(value))
	if len(parts) == 0 {
		return "", false
	}
	return parts[1], true
}

// resolveEnvironmentExpressions 仅展开编译期已知的变量。
// 未知变量保留给 Runtime 在收集依赖输出后解析。
func resolveEnvironmentExpressions(value string, environment map[string]*yaml.Variable) string {
	return envExpression.ReplaceAllStringFunc(value, func(expression string) string {
		parts := envExpression.FindStringSubmatch(expression)
		variable := environment[parts[1]]
		if variable == nil || variable.Secret != "" {
			return expression
		}
		return variable.Value
	})
}
