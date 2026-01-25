# go-yaml

[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

## 项目概述

go-yaml 是一个用于 [Drone CI](https://github.com/drone/drone) 配置文件的 YAML 处理工具库和命令行工具。

### 主要功能

- **解析（Parse）** - 解析 Drone YAML 配置文件
- **校验（Lint）** - 检查配置文件的语法和语义错误
- **格式化（Format）** - 格式化和美化 YAML 配置文件
- **签名（Sign）** - 使用 HMAC 对配置文件进行签名
- **验证（Verify）** - 验证配置文件的签名
- **编译（Compile）** - 将 YAML 配置编译为 Drone 运行时格式
- **转换（Convert）** - 支持从其他 CI 系统（GitLab CI、Bitbucket Pipelines 等）转换配置

### 技术栈

- **语言**: Go 1.24+
- **YAML 处理**: gopkg.in/yaml.v3, gopkg.in/yaml.v2
- **命令行**: kingpin.v2
- **运行时**: drone-runtime

## 快速开始

### 编译构建

#### 环境要求

- Go 1.24 或更高版本
- Git

#### 编译步骤

```bash
# 克隆仓库
git clone https://github.com/open-beagle/go-yaml.git
cd go-yaml

# 下载依赖
go mod download

# 编译
go build -o go-yaml-cli main.go

# 或使用 make（如果有 Makefile）
make build
```

#### 构建产物

编译成功后会生成 `go-yaml-cli` 可执行文件，可以直接运行。

### 使用示例

#### 校验配置文件

```bash
./go-yaml-cli lint samples/simple.yml
```

#### 格式化配置文件

```bash
# 输出到标准输出
./go-yaml-cli fmt samples/simple.yml

# 保存到原文件
./go-yaml-cli fmt samples/simple.yml --save
```

#### 签名配置文件

使用 32 字节的密钥对配置文件进行 HMAC-SHA256 签名：

```bash
# 输出签名
./go-yaml-cli sign 642909eb4c3d47e33999235c0598353c samples/simple.yml

# 将签名写入文件
./go-yaml-cli sign 642909eb4c3d47e33999235c0598353c samples/simple.yml --save
```

#### 验证签名

```bash
./go-yaml-cli verify 642909eb4c3d47e33999235c0598353c samples/simple.yml
```

#### 编译配置

将 YAML 配置编译为 Drone 运行时 JSON 格式：

```bash
./go-yaml-cli compile samples/simple.yml
```

#### 转换配置

从其他 CI 系统转换配置：

```bash
./go-yaml-cli convert .gitlab-ci.yml
```

### 自动测试

#### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行测试并显示详细输出
go test -v ./...

# 运行测试并生成覆盖率报告
go test -cover ./...

# 生成 HTML 覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

#### 测试要求

- 单元测试覆盖率目标：> 70%
- 所有测试必须通过
- 无竞态条件（使用 `go test -race`）

### 发布版本

#### 版本号规范

本项目遵循 [语义化版本](https://semver.org/lang/zh-CN/) 规范：

- **主版本号（MAJOR）**: 不兼容的 API 变更
- **次版本号（MINOR）**: 向下兼容的功能新增
- **修订号（PATCH）**: 向下兼容的问题修正

格式：`vMAJOR.MINOR.PATCH`，例如：`v1.2.3`

#### 发布流程

1. **更新版本号**

   ```bash
   # 更新代码中的版本信息
   # 更新 CHANGELOG.md
   ```

2. **创建 Git 标签**

   ```bash
   git tag -a v1.2.3 -m "Release v1.2.3"
   git push origin v1.2.3
   ```

3. **构建发布版本**

   ```bash
   # 构建多平台版本
   GOOS=linux GOARCH=amd64 go build -o go-yaml-cli-linux-amd64 main.go
   GOOS=darwin GOARCH=amd64 go build -o go-yaml-cli-darwin-amd64 main.go
   GOOS=windows GOARCH=amd64 go build -o go-yaml-cli-windows-amd64.exe main.go
   ```

4. **发布检查清单**
   - [ ] 所有测试通过
   - [ ] 代码已合并到主分支
   - [ ] CHANGELOG.md 已更新
   - [ ] 版本号已更新
   - [ ] Git 标签已创建
   - [ ] 构建产物已生成
   - [ ] 发布说明已准备

## 项目结构

```
.
├── main.go                 # 命令行入口
├── yaml/                   # YAML 处理核心库
│   ├── compiler/          # 编译器
│   ├── converter/         # 格式转换器
│   ├── linter/            # 语法检查器
│   ├── pretty/            # 格式化器
│   └── signer/            # 签名工具
├── samples/               # 示例配置文件
└── .tmp/                  # 临时文档和迁移记录
```

## 开发指南

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 使用 `golint` 检查代码质量
- 添加必要的注释和文档

### 提交规范

提交信息格式：`<type>: <subject>`

类型（type）：

- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建/工具链相关

## 依赖说明

本项目已完成 YAML 库迁移，使用官方维护的版本：

- `gopkg.in/yaml.v3` - 主要 YAML 处理库
- `gopkg.in/yaml.v2` - 用于 MapSlice 支持（legacy 代码）

详细的迁移记录请查看 `.tmp/FINAL_MIGRATION_REPORT.md`

### 为什么同时使用 yaml.v2 和 yaml.v3？

yaml.v2 仅用于 legacy 转换器中的 `MapSlice` 功能，用于保持 YAML 键值对的顺序。这对于确保 CI/CD 流程按正确顺序执行至关重要。

详细说明请参考：[为什么使用 MapSlice？](docs/why-mapslice.md)

## 文档

- [为什么使用 MapSlice？](docs/why-mapslice.md) - 解释 yaml.v2 MapSlice 的作用和业务价值

## 许可证

Apache License 2.0

## 贡献

欢迎提交 Issue 和 Pull Request！

## 相关链接

- [Drone CI](https://github.com/drone/drone)
- [Drone Runtime](https://github.com/drone/drone-runtime)
- [YAML 规范](https://yaml.org/)
