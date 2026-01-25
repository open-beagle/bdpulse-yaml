# 为什么使用 MapSlice？

## 问题背景

Go 语言的 `map` 类型是**无序的**，这在处理 YAML 配置文件时会导致问题。

### 示例：顺序丢失的问题

```yaml
# YAML 文件中定义的 CI/CD 流程
pipeline:
  build: { ... } # 第 1 步：构建
  test: { ... } # 第 2 步：测试
  deploy: { ... } # 第 3 步：部署
```

如果使用普通的 `map[string]interface{}` 解析：

```go
type Config struct {
    Pipeline map[string]interface{} `yaml:"pipeline"`
}

// 遍历时顺序是随机的！
for name := range config.Pipeline {
    fmt.Println(name)
}
// 可能输出：deploy, build, test ❌
// 每次运行顺序都可能不同
```

**问题**：CI/CD 流程必须按正确顺序执行，如果先部署再构建，就会失败！

## MapSlice 解决方案

`MapSlice` 是 `gopkg.in/yaml.v2` 提供的**有序键值对列表**：

```go
type MapSlice []MapItem

type MapItem struct {
    Key   interface{}  // 键名
    Value interface{}  // 键值
}
```

### 使用 MapSlice 保持顺序

```go
type Config struct {
    Pipeline yaml.MapSlice `yaml:"pipeline"`
}

// 遍历时顺序固定
for _, item := range config.Pipeline {
    fmt.Println(item.Key)
}
// 输出：build, test, deploy ✅
// 与 YAML 文件顺序一致
```

## 本项目中的应用场景

### 场景 1：保持容器执行顺序

**文件**：`yaml/converter/legacy/internal/container.go`

```go
func (c *Containers) UnmarshalYAML(unmarshal func(interface{}) error) error {
    slice := yaml.MapSlice{}
    if err := unmarshal(&slice); err != nil {
        return err
    }

    // 按 YAML 文件中的顺序处理容器
    for _, s := range slice {
        container := Container{}
        yaml.Unmarshal(yaml.Marshal(s.Value), &container)
        container.Name = fmt.Sprintf("%v", s.Key)
        c.Containers = append(c.Containers, &container)
    }
    return nil
}
```

**业务价值**：确保 CI/CD 流程按正确顺序执行。

### 场景 2：展开 YAML Merge Keys

**文件**：`yaml/converter/legacy/internal/yaml.go`

```go
type temporary struct {
    Attributes map[string]interface{} `yaml:",inline"`
    Pipeline   yaml.MapSlice          `yaml:"pipeline"`
}

func expandMergeKeys(b []byte) ([]byte, error) {
    v := new(temporary)
    yaml.Unmarshal(b, v)  // 自动展开 merge keys
    return yaml.Marshal(v)
}
```

**示例**：

```yaml
# 输入：使用 merge keys
defaults: &defaults
  image: golang:1.11

pipeline:
  <<: *defaults
  commands:
    - go build

# 输出：展开后保持字段顺序
pipeline:
  image: golang:1.11
  commands:
    - go build
```

**业务价值**：支持 Drone 0.8 旧版本配置迁移到 1.0+。

## 为什么保留 yaml.v2？

### yaml.v3 的变化

yaml.v3 移除了 `MapSlice`，改用更复杂的 `Node` API：

```go
// yaml.v2（简单）
slice := yaml.MapSlice{}
yaml.Unmarshal(data, &slice)
for _, item := range slice {
    // 3 行代码搞定
}

// yaml.v3（复杂）
var node yaml.Node
yaml.Unmarshal(data, &node)
for i := 0; i < len(node.Content); i += 2 {
    keyNode := node.Content[i]
    valueNode := node.Content[i+1]
    // 需要 20+ 行代码处理
}
```

### 权衡分析

| 方案           | 优点                                      | 缺点                                                  |
| -------------- | ----------------------------------------- | ----------------------------------------------------- |
| 保留 yaml.v2   | ✅ 代码简单<br>✅ 功能稳定<br>✅ 无需重构 | ⚠️ 多一个依赖                                         |
| 迁移到 yaml.v3 | ✅ 统一依赖                               | ❌ 需要大量重构<br>❌ 容易引入 bug<br>❌ 测试工作量大 |

### 当前决策

**保留 yaml.v2 用于 legacy 代码**

理由：

1. **影响范围小**：只有 3 个文件使用
2. **代码性质**：legacy 转换器（未来可能废弃）
3. **功能稳定**：无 bug，无需改动
4. **成本收益**：重构成本高，收益低

## 实际影响

### 依赖情况

```go
// go.mod
require (
    gopkg.in/yaml.v2 v2.4.0  // 用于 MapSlice（3 个文件）
    gopkg.in/yaml.v3 v3.0.1  // 主要使用（其他所有文件）
)
```

### 使用统计

- **yaml.v3**：9 个文件（主要 YAML 处理）
- **yaml.v2**：3 个文件（legacy 转换器）

## 未来计划

如果需要完全移除 yaml.v2：

1. **评估必要性**
   - legacy 转换器是否还需要？
   - 是否还有用户使用 Drone 0.8？

2. **重构方案**
   - 使用 yaml.v3 的 Node API
   - 重写 UnmarshalYAML 方法
   - 添加充分的测试

3. **测试验证**
   - 单元测试
   - 集成测试
   - 回归测试

## 总结

- **MapSlice 作用**：保持 YAML 键值对的顺序
- **业务价值**：确保 CI/CD 流程按正确顺序执行
- **使用场景**：
  - 保持容器执行顺序
  - 展开 merge keys 时保持字段顺序
- **保留原因**：legacy 代码，功能稳定，重构成本高收益低
- **技术债务**：合理的技术债务，在功能稳定和代码简洁之间取得平衡

## 相关文件

- `yaml/converter/legacy/internal/container.go` - 容器顺序处理
- `yaml/converter/legacy/internal/yaml.go` - Merge keys 展开
- `yaml/converter/legacy/internal/config.go` - Legacy 配置转换

## 参考资料

- [YAML v2 MapSlice 文档](https://pkg.go.dev/gopkg.in/yaml.v2#MapSlice)
- [YAML v3 Node API 文档](https://pkg.go.dev/gopkg.in/yaml.v3#Node)
- [Go Map 无序性说明](https://go.dev/blog/maps)
