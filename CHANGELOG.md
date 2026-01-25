# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- 全新的 README.md 文档，包含完整的项目说明和使用指南
- CHANGELOG.md 文件用于记录版本变更
- 依赖升级计划文档（.tmp/DEPENDENCY_UPGRADE_PLAN.md）

### Changed

- **重大变更**: 迁移 YAML 库依赖
  - 移除 `github.com/buildkite/yaml` (已废弃)
  - 移除 `github.com/ghodss/yaml` (功能已被标准库替代)
  - 移除 `github.com/vinzenz/yaml` (过时的 fork)
  - 使用 `gopkg.in/yaml.v3` 作为主要 YAML 处理库
  - 使用 `gopkg.in/yaml.v2` 用于 MapSlice 支持（legacy 代码）
  - 使用标准库 `encoding/json` 进行 YAML→JSON 转换

- **依赖升级**: 升级多个依赖到最新稳定版本
  - `github.com/bmatcuk/doublestar`: v1.1.1 → v1.3.4
  - `github.com/docker/go-units`: v0.3.3 → v0.5.0
  - `github.com/opencontainers/go-digest`: v1.0.0-rc1 → v1.0.0
  - `gopkg.in/yaml.v2`: v2.2.2 → v2.4.0

### Fixed

- 修复 `convert.go` 和 `convert_oss.go` 的构建标签冲突问题

### Security

- 移除 3 个过时/非官方的依赖库，提升安全性
- 升级依赖修复已知安全问题

- 移除 3 个过时/非官方的依赖库，提升安全性

## [Previous Versions]

历史版本信息待补充。

---

## 迁移说明

如果您从旧版本升级，请注意：

1. **依赖变更**: 项目不再依赖 `buildkite/yaml`、`ghodss/yaml` 和 `vinzenz/yaml`
2. **API 兼容**: 所有公开 API 保持兼容，无需修改使用代码
3. **功能验证**: 所有功能已通过测试验证

详细的迁移记录请查看 `.tmp/FINAL_MIGRATION_REPORT.md`
