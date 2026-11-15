# gatework Go 工程规范

版本：1.1.0（2026-11-15）。适用于 gatework 组织的全部 Go 模块。

本规范是长期迭代的唯一基准。各仓库的 `AGENTS.md` 只补充本仓库特有的领域边界与安全语义，不得与本规范冲突；冲突时以本规范为准，并通过修订本规范（而不是在仓库里另立规则）解决。

语义依据为 [Go 语言规范](https://go.dev/ref/spec)、[内存模型](https://go.dev/ref/mem)及最低支持版本的标准库文档；惯用法参考：[Effective Go](https://go.dev/doc/effective_go)、[Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)、[Go Doc Comments](https://go.dev/doc/comment)、[Google Go Style Guide](https://google.github.io/styleguide/go/)。本规范只写这些文档没有定论、或我们有意收紧的部分。

每条规则有稳定编号（如 `NAM-2`），评审意见和提交说明直接引用编号。规则默认为“必须”；“推荐”允许评审说明替代方案，“可选”按需采用。检查标签只表示检查方式，不表示要求强度；未标标签的语义要求由评审验证。标注含义：

- **[lint]** 由共享 `golangci.yml` 强制，CI 失败即违规；
- **[arch]** 由仓库的架构检查（`architecture.yaml` + archcheck）强制；
- **[review]** 无法机器判定，由评审把关。

---

## 0. 总则

- **GEN-1 投产前不承诺兼容。** 各模块在 AGENTS.md 声明成熟度与兼容承诺；v0 不代表未投产。仅明确声明未投产且不承诺兼容的模块适用本条：发现更好的模型、命名或包结构时直接改，不保留兼容层、别名、`Deprecated` 包装或双写路径。破坏性变更在提交说明里写 `!`（如 `refactor!:`）。投产后的模块改为遵循语义化版本，届时在该仓库 `AGENTS.md` 声明。
- **GEN-2 行为等价先于删除。** 重写已有业务逻辑时，先用金样或差分测试证明新旧实现等价，再切换、再删除旧代码。有意的不等价必须在测试里逐条列出并写明理由。
- **GEN-3 一次只做一件事。** 一个提交只包含一个可独立验证的变化；重构与行为变更分开提交。
- **GEN-4 标准库优先。** 新增第三方依赖需在 PR 中说明标准库为何不够。开发工具（golangci-lint、govulncheck、stdsync、archcheck）在共享的 `standard.mk` 中以 `go run module@version` 固定版本，不写入库模块的 `go.mod`，以免工具依赖进入库的依赖图。

## 1. 模块与工具链

- **MOD-1** `go` 指令统一为当前受支持的最旧版本（现为 `go 1.26.0`）；CI 在 1.26.x 与 1.27.x 两个版本上测试。新版本发布后整体上调，不单独落后。
- **MOD-2** 日常构建使用 `GOWORK=off GOFLAGS=-mod=readonly`。不提交 `go.work`，不在 `go.mod` 中持久化 `replace`。跨模块联调用临时 `go.work`（已加入 `.gitignore`）。
- **MOD-3** 依赖内部模块时使用已发布的 tag；需要未发布的改动时先发布被依赖模块。
- **MOD-4** 模块路径为 `github.com/gatework/<repo>`；仓库名即模块名（不再出现 `anytls-go`、`firemon-rb` 这类与模块名不一致的仓库名，迁移时一并更正）。
- **MOD-5** 每个仓库提供统一入口 `make check`，CI 只调用它。仓库的 `Makefile` 以 `include standard.mk` 引入共享目标，依次执行：规范文件一致性（`stdsync check`）、格式（gofmt + goimports）、`go mod tidy -diff`、`go vet`、`golangci-lint run`、架构检查（`archcheck`）、`go test -race -shuffle=on -count=1 ./...`、`govulncheck ./...`；仓库特有的检查通过 `CHECK_EXTRA` 追加。提交前必须在本地跑通 `make check`。
- **MOD-6** `.golangci.yml` 与 `standard.mk` 是由 `stdsync write` 分发的副本，不得在仓库内修改；需要调整时修订本规范并重新分发。仓库确需局部放宽时，只能在代码处使用带理由的 `//nolint:<linter> // 理由`。

## 2. 包与文件组织

### 2.1 布局

以下为复杂业务模块的推荐布局；小型库可用单包，CLI 可只保留命令与内部实现，协议库按协议职责组织。只创建实际需要的目录。

```
<repo>/
  go.mod  doc.go  <root>.go     根包：模块的核心领域 API，包名为简短合法标识符，不要求等于仓库名
  <concept>/                    以领域概念划分的子包（fabric、change、session）
  vendors/<vendor>/             厂商语法适配（解析、渲染），只依赖领域包
  adapter/<system>/             对外部系统/传输的适配（connector、postgres、http）
  internal/                     不对外承诺的实现：工作流、引导、CLI
    workflow/<usecase>/         应用用例（collect、review、apply）
    bootstrap/                  组装依赖（配置 → 对象图）
    cli/                        命令行解析与输出
  cmd/<binary>/main.go          只调用 internal/cli，不含业务逻辑
  testdata/                     测试输入与金样
  docs/  docs/adr/              设计说明与架构决策记录
  scripts/                      开发与发布脚本
  architecture.yaml             包分层声明（ARC-1）
  .golangci.yml  Makefile  AGENTS.md  README.md  CHANGELOG.md  GLOSSARY.md（领域模块）
```

- **PKG-1 [review]** 按领域概念分包，不按技术层分包：禁止 `utils`、`util`、`common`、`helpers`、`models`、`types`、`base`、`misc`、`pkg/` 这类名字 [lint: revive package-naming 覆盖部分禁用名称，其余由评审检查]。一个包回答一个问题："这个包提供什么"。
- **PKG-2** 包名是小写单个单词，不含下划线或大小写混合，通常与目录名一致，根包、版本路径和命令包除外。优先单数（`change` 而非 `changes`）；复数只用于与标准库同类的场合，即包是对某个同名类型或概念集合的操作（`strings`、`slices`、`rules` 作为规则目录）。导入别名只在冲突时使用，且整个模块对同一个包使用同一个别名。
- **PKG-3** 每个包有 `doc.go`，以 `// Package x ...` 开头描述该包的职责和关键不变量 [lint: revive package-comments / stylecheck ST1000]。
- **PKG-4** `internal/` 是默认选择：不确定是否对外暴露时放进 `internal/`。对外 API 面越小越好。
- **PKG-5** `cmd/<binary>/main.go` 不超过 30 行：构造依赖、调用 `cli.Run(ctx, os.Args[1:], deps)`、以其返回码退出。

### 2.2 依赖方向

- **ARC-1 [arch]** 按职责声明无环依赖；复杂业务模块推荐由外向内：`cmd → internal/cli → internal/bootstrap → internal/workflow → 领域包 ← vendors / adapter`。领域包不导入 `adapter`、`vendors`、`internal/*`、`os`/`net`/`database/sql` 等 I/O 包（时间、随机数通过注入获得）。每个仓库在 `architecture.yaml` 中声明包的层级与允许的依赖，新增包必须先登记并审查其规则覆盖；登记本身不限制导入。配置必须为单个 YAML 文档，declared 与 rules 非空，每条规则在排除 except 后至少匹配一个生产包。
- **ARC-2 [arch]** 禁止导入环。同层包可按明确的职责单向依赖；只有共享概念稳定且具有独立职责时才抽取公共包，不为满足层级人为增加转发层。
- **ARC-3 [review]** 跨模块共享的、与具体领域无关的机制放在被依赖的内核模块（如 `go-netops`、`net-connector`、`set`、`parser`），不在多个仓库里各写一份。

### 2.3 文件

- **FIL-1** 文件名为小写蛇形，按"概念"或"概念_方面"命名：`session.go`、`session_read.go`、`plan_verify.go`。不使用 `types.go`、`utils.go`、`helpers.go`、`misc.go`。
- **FIL-2** 一个文件以一个主要类型为中心：类型定义、构造函数、方法依次排列 [lint: funcorder]。文件内顺序：包注释（仅 `doc.go`）→ 导入 → 常量 → 变量 → 主类型 → 构造函数 → 导出方法 → 未导出方法 → 未导出辅助函数。
- **FIL-3** 文件超过约 500 行或含多个互不相关的主类型时拆分。
- **FIL-4** 测试文件与被测文件同名加 `_test.go`；平台差异用构建后缀（`_linux.go`；Unix 家族文件需显式 `//go:build unix`，仅 `_unix.go` 无效）而不是运行时判断。测试专用的导出钩子放在 `export_test.go`。

## 3. 面向对象建模

Go 没有类和继承；"面向对象"在这里指：**先用类型刻画领域名词和它们的不变量，再把行为挂在类型上，用接口表达可替换的角色，用组合构造复杂对象。**

### 3.1 设计顺序

- **OOP-1 [review]** 复杂领域功能推荐按以下顺序设计并在 PR 中体现：① 领域术语进 `GLOSSARY.md`；② 领域类型（值对象、实体、聚合）及其不变量；③ 角色接口（端口）；④ 用例/工作流；⑤ 适配器实现。简单功能直接以具体类型或函数实现；接口从真实调用需求中提取。
- **OOP-2 [review]** 领域模型使用具名结构体与具名类型，禁止以 `map[string]any`、`[]any`、`any` 作为领域数据的载体。`any`/通用记录只允许出现在解码边界，并在边界内立即转换为类型化模型。

### 3.2 类型与不变量

- **OOP-3 解析而非校验。** 不变量在构造时确立：通过 `NewX(...) (X, error)`、`ParseX(string) (X, error)` 或 `DecodeX([]byte) (X, error)` 得到的值必然有效，后续变更也必须维护不变量。零值、显式转换和反序列化可能绕过构造函数，必须在可信边界处理；解码失败不得留下部分有效状态。用专门的类型表达约束（`netip.Addr` 而非 `string` 的地址，`VLAN` 而非 `int`）。
- **OOP-4 封装。** 有不变量的类型使用未导出字段，通过方法读取和变更；没有不变量的纯数据（DTO、配置、事实记录）使用导出字段。不要两者混合：一个结构体要么全是可直接改写的数据，要么全由方法守护。
- **OOP-5 零值。** 类型的零值要么可直接使用（`bytes.Buffer` 式），要么在文档中声明"零值不可用，使用 NewX"，并让方法对零值返回错误而不是 panic。
- **OOP-6 值语义与指针语义。** 小而不可变的值对象（ID、Severity、Subject）用值接收者并按值传递；有身份、有状态或含锁/连接的实体（Session、Report、Device）用指针接收者，构造函数返回 `*T`。同一类型推荐使用一致接收者；值类型的查询/编码方法与指针解码方法可以不同 [lint: recvcheck]。
- **OOP-7 枚举。** 使用具名类型加常量，常量以类型名为前缀：`type Severity int`、`SeverityHigh`。字符串枚举实现 `String()`；需要序列化时实现 `MarshalText/UnmarshalText` 并拒绝未知值；提供 `Valid() bool`。业务穷尽分派的 `switch` 必须显式覆盖全部已知取值，不能用 default 隐藏遗漏；外部未知值在解码边界按协议拒绝或保留 [lint: exhaustive]。零值表示"未设置"时显式命名（`SeverityUnset`）。
- **OOP-8 所有权。** 构造函数和 setter 复制传入的切片/映射，getter 返回副本（`slices.Clone`、`maps.Clone`），这些函数只浅复制；嵌套切片、映射和指针必须隔离到不变量要求的深度，或明确声明借用、共享只读、所有权转移及有效期。

### 3.3 接口

- **IFC-1 接口由使用方定义。** 在需要该能力的包里声明最小接口，实现方返回具体类型（"接受接口，返回结构体"）。不要为只有一个实现、且不是 I/O 边界的类型预先定义接口 [lint: iface]。
- **IFC-2 接口要小。** 一到三个方法；更多方法通常意味着应该拆分角色。通过嵌入组合大接口（`io.ReadWriteCloser` 式）。
- **IFC-3 命名。** 单方法接口用"方法名 + er"（`Reader`、`Renderer`、`Fingerprinter`）；表示角色的多方法接口用名词（`Journal`、`Session`、`Store`）。不使用 `I` 前缀或 `Interface` 后缀。接口方法参数必须具名 [lint: inamedparam]。
- **IFC-4 编译期断言。** 实现某个接口的类型在依赖方向允许的位置（实现包、组装包或测试）写 `var _ change.Change[*Config] = SVIACL{}`。
- **IFC-5 封闭集合。** 不同变体携带不同数据与行为时，可用含未导出方法的接口表达封闭集合；普通状态采用具名枚举和集中式 switch。避免同一分派规则散布多处；封闭接口仍需评审或专项工具检查类型分派完整性。
- **IFC-6 端口。** 领域/工作流需要的外部能力（时钟、存储、会话、日志）显式注入；单一能力可用函数，多个相关行为可用小接口，稳定具体类型无需机械包装；测试使用手写的假实现，而不是 mock 框架。

### 3.4 组合与多态

- **CMP-1** 用组合而不是嵌入来复用实现。嵌入只用于有意把被嵌入类型的方法集提升为外层类型 API 的场合；导出结构体不嵌入导出的具体类型，以免意外扩大 API。嵌入字段放在字段列表最前面 [lint: embeddedstructfieldcheck]。
- **CMP-2** 泛型用于"同一算法作用于不同类型"（容器、`Rule[M]`、`Catalog[M]`、`Observation[T]`），不用来模拟继承层次。类型参数只要求算法实际使用的能力；不额外限制可接受类型。
- **CMP-3** 不使用包级可变状态和 `init()` 注册；注册表、目录、配置都显式构造并作为依赖传入 [lint: gochecknoinits, gochecknoglobals 以 review 为准，预编译正则与哨兵错误除外]。
- **CMP-4** 构造函数的参数：必需依赖作为位置参数；可选项推荐使用 Options 结构体；需要逐项默认值或稳定扩展时可用函数式选项。明确零值、重复选项、非法组合与验证时机，不按选项数量机械选择。

## 4. 命名

- **NAM-1** 使用 MixedCaps；首字母缩略词保持全大写或全小写：`ID`、`URL`、`HTTP`、`IP`、`TCP`、`UDP`、`TLS`、`SSH`、`JSON`、`YAML`、`ACL`、`VLAN`、`VNI`、`VRF`、`BGP`、`EVPN`、`SSID`、`AP` [lint: revive var-naming / stylecheck ST1003]。
- **NAM-2 不重复包名。** 调用方看到的是 `pkg.Name`：`change.Plan` 而非 `change.ChangePlan`，`rules.Fabric` 而非 `rules.FabricRules`，`report.New` 而非 `report.NewReport`。包内唯一的主类型的构造函数叫 `New`。
- **NAM-3** Getter 不加 `Get`：`Name()`；setter 用 `SetName`。布尔函数与方法表达明确谓词，可用 `Is`、`Has`、`Can`、`Should` 前缀或 `Valid`、`Equal`、`Contains`、`Empty` 等惯用名称，布尔字段用形容词或过去分词（`Enabled`、`Complete`）。
- **NAM-4** 接收者名为类型名的一两个字母缩写，同一类型保持一致，不用 `this`/`self` [lint: revive receiver-naming]。
- **NAM-5** 局部变量名长度与作用域成正比：循环变量与短作用域用 `i`、`c`、`r`；跨越多屏的变量用完整单词。不用匈牙利命名、类型后缀（`nameStr`、`userMap`）。
- **NAM-6** 错误：哨兵错误变量 `ErrNotFound`，错误类型 `ParseError`；错误信息小写开头、无结尾标点，以包名前缀限定（`"change: service template not found"`）[lint: errname, revive error-strings / stylecheck ST1005]。
- **NAM-7** 常量不用全大写；导出常量与变量按用途命名（`MaxACLSequence`），不按类型命名。
- **NAM-8** 测试：`TestType_Method`、`TestFunction`，子测试名为描述性短语（`"rejects unknown kind"`）；基准 `BenchmarkX`；模糊测试 `FuzzX`；示例 `ExampleType_Method`。
- **NAM-9** 不遮蔽预声明标识符（`len`、`new`、`error`、`string`、`max`）[lint: predeclared]。
- **NAM-10 语言。** 标识符、代码注释、godoc、错误信息、提交说明使用英文。面向运维人员的输出文案（报告、CLI 提示）可以使用中文，但集中放在规则元数据或文案表中，不在业务逻辑里内联拼接。

## 5. 错误

- **ERR-1** 错误是值：返回 `error`，不 panic。`panic` 只用于不可能发生的程序缺陷，以及 `MustX` 形式的包级初始化（如 `regexp.MustCompile`）。禁止用 `panic`/`recover` 实现控制流。
- **ERR-2** 只在增加有用上下文且愿意公开底层错误契约时用 `%w` 包装：`fmt.Errorf("decode plan %s: %w", path, err)`。实现细节转换为公开的领域错误；允许直接返回已有充分上下文的错误，主动隐藏底层类型时说明理由。不加没有信息量的包装（`"failed: %w"`）[lint: errorlint；wrapcheck 只对 adapter 包启用，且不要求包装 gatework 模块已分类的错误]。
- **ERR-3** 调用方需要分支判断的条件用哨兵错误（`errors.Is`）或错误类型（`errors.As`/`errors.AsType`），不比较错误字符串。
- **ERR-4** 每个错误只处理一次：要么返回，要么记录日志并消化，不能既记录又返回。
- **ERR-5** 错误与诊断中不包含凭据、设备原始输出或用户敏感数据；需要携带时使用专门的、默认脱敏的字段。
- **ERR-6** 生产 API 不返回 `nil, nil`；"不存在"用 `(T, bool)` 或哨兵错误表达 [lint: nilnil，测试中的假实现除外]。

## 6. 上下文、并发与资源

- **CON-1** 会阻塞、做 I/O 或可能耗时的函数第一个参数是 `ctx context.Context`；不把 `Context` 存进结构体（生命周期对象的内部取消除外，需注释说明）[lint: containedctx, noctx]。
- **CON-2** 每个 goroutine 都有明确的所有者和退出条件；启动它的函数负责等待它（`sync.WaitGroup.Go`、`errgroup`）。不留"发射后不管"的 goroutine。
- **CON-3** 通道由发送方关闭；参数和字段使用单向通道类型表达所有权。
- **CON-4** 优先用通道传递所有权、用互斥锁保护状态；锁与它保护的字段写在一起并注释。不复制含锁的值 [lint: govet copylocks]。
- **CON-5** 时间与随机性通过注入的 `func() time.Time`/时钟接口获得；并发时间测试推荐 testing/synctest，允许 bubble 内虚拟 Sleep；不用真实 Sleep 猜测调度完成。安全令牌使用 crypto/rand，不以可预测随机源替代。
- **CON-6** 获得资源的函数紧接着 `defer` 释放；`Close` 的错误在写路径上必须检查。持久化写入使用原子写（临时文件 + fsync + rename），由共享的 store 包提供。

## 7. 日志与可观测性

- **LOG-1** 库包不打印、不调用 `log` 包；需要日志时接受注入的 `*slog.Logger`（默认丢弃）。只有 `cmd` 与 `internal/cli` 写标准输出/错误 [lint: forbidigo 禁止非 CLI 包使用 `fmt.Print*`、`log.*`]。
- **LOG-2** 使用 `slog` 结构化字段，键名小写蛇形；不使用拼接后的消息字符串携带数据 [lint: sloglint]。

## 8. 测试

- **TST-1** 公共行为用外部测试包（`package x_test`）测试；只有验证未导出不变量时才用内部测试。
- **TST-2** 多个用例用表驱动加 `t.Run` 子测试；辅助函数首行 `t.Helper()` [lint: thelper]；可并行的测试调用 `t.Parallel()` [lint: tparallel]；使用 `t.Context()`、`t.TempDir()`、`t.Setenv()` [lint: usetesting]。
- **TST-3** 金样放 `testdata/`，通过 `-update` 标志重新生成，并在提交中审阅差异。真实设备或生产数据不进仓库；需要时通过环境变量指向仓库外的私有语料，只输出聚合结果。
- **TST-4** 解析器、解码器和任何处理外部输入的函数提供模糊测试；性能敏感路径提供基准测试（`b.Loop()`），以 benchstat 对比。
- **TST-5** 默认单元测试不依赖外部网络或真实设备，使用函数、小接口或内存实现。集成测试可使用受控本地服务，明确准备与清理；真实环境验证需显式启用并独立记录。测试层次和运行入口写入 README，额外门禁通过 CHECK_EXTRA 接入。
- **TST-6** 不使用断言库与 mock 框架；用标准库 testing、slices/maps 或 github.com/google/go-cmp/cmp 比较（标准库 cmp 提供排序比较，不提供结构差异比较）。
- **TST-7** 公共 API 至少有一个可运行的 `Example`。

## 9. 文档

- **DOC-1** 所有导出标识符有以其名称开头的完整句子注释 [lint: revive exported]。注释说明"是什么、保证什么、调用方要注意什么"，不复述代码。
- **DOC-2** `README.md`：用途、安装、最小用法、文档索引。`AGENTS.md`：只写本仓库特有的边界、安全语义和检查命令，首行引用本规范。`CHANGELOG.md` 按 Keep a Changelog 维护。领域模块维护 `GLOSSARY.md`；重要决策写 `docs/adr/NNNN-title.md`。
- **DOC-3** 文档与代码同一提交更新；文档中引用的文件、测试名必须真实存在（由仓库的交付检查验证）。

## 10. 格式与静态检查

- **FMT-1 [lint]** gofmt 与 goimports（本地前缀 `github.com/gatework`）。导入分三组：标准库、第三方、`github.com/gatework/...`。
- **FMT-2 [lint]** 函数不超过 80 行、50 条语句，认知复杂度不超过 30 [lint: funlen, gocognit]。超出时审查职责与嵌套；拆分必须形成清晰步骤。集中式分派等拆分后更难理解的情况可使用带具体理由的 nolint。
- **FMT-3 [lint]** `//nolint` 必须指明具体 linter 并写出理由 [lint: nolintlint]。
- **FMT-4 [lint]** 使用现代写法：`for i := range n`、`min`/`max`、`slices`/`maps`、`strings.Cut`、`any` [lint: modernize, intrange]。

## 11. 数据、算法与运行时契约

以下规则均为必须，由评审检查；性能要求限于关键路径或输入规模不受信任的场景。

- **DAT-1 [review]** 按查询、遍历和更新方式及预期规模选择数据结构。明确顺序、重复和唯一性；高频查询避免反复线性扫描，小集合不无依据增加索引。容量已知时推荐预分配，不猜测巨大容量。
- **DAT-2 [review]** API 明确缺失、零值、空集合与未知值的语义；map 查询需要区分不存在时用 comma-ok。内部不无故区分 nil 与空切片，协议边界显式处理 null/空数组。接口返回真正的 nil，避免 typed-nil；类型断言处理失败。
- **DAT-3 [review]** 明确输入是否原地修改、输出是否借用及有效期。切片重切、append、缓存与池复用不得破坏所有权；传递结构体也不代表其引用字段已独立复制。
- **ALG-1 [review]** 关键算法说明前置条件、结果性质及终止条件，以性质测试、差分测试或简单参考实现证明；比较器满足一致的严格弱序，不用可能溢出的减法比较整数。
- **ALG-2 [review]** 关键路径说明预期输入规模及时间、空间复杂度。外部输入限制长度、数量、递归深度与搜索预算；预算耗尽显式报告未完成，不作为完整结果返回。
- **ALG-3 [review]** 报告、执行计划、序列化及哈希需要确定性时，显式排序并定义并列规则；业务结果不得依赖 map 遍历或 goroutine 调度顺序。
- **ALG-4 [review]** 性能优化以真实工作负载测量为依据；基准覆盖典型与边界规模，对比时间、分配和内存。缓存、池、unsafe 或额外索引需说明收益、容量和失效条件。
- **MDL-1 [review]** 领域类型明确实体身份、值相等性、规范化与一致性边界。状态转换用表达业务意图的方法维护约束；例如 Approve 应验证可批准状态，不能由任意 SetStatus 绕过。
- **RUN-1 [review]** 有状态 API 文档说明并发安全性、资源关闭责任及 Close 后行为。资源转移所有权时不用局部 defer 提前关闭；循环内短期资源及时释放。部分初始化失败清理已获得资源；持久化原子替换明确文件与目录同步、平台和文件系统保证。
- **RUN-2 [review]** 外部操作明确取消、超时与阻塞解除方式；调用 cancel 并实际响应取消。重试限定次数或总预算，仅对可重试且幂等的操作执行，退避可取消。工作池、队列和缓存有容量与过载策略；后台任务由生命周期所有者取消并等待。
- **RUN-3 [review]** 多步操作明确部分成功、提交点与恢复方式，不能用成功空值隐藏失败。安全敏感操作验证整数转换边界、路径权限和不可信输入；不得假设所有调用都已由解析器验证。

## 12. 版本控制与发布

- **GIT-1** 提交说明遵循 Conventional Commits：`feat`、`fix`、`refactor`、`perf`、`test`、`docs`、`build`、`ci`、`chore`，破坏性变更加 `!`；标题用英文祈使句，正文说明动机与验证。
- **GIT-2** 分支命名 `<type>/<topic>`（`refactor/typed-model`）。不直接向 `main` 提交；通过 PR 合并，CI 通过后才可合并。
- **GIT-3** 发布 tag、发布仓库、推送到他人仓库属于对外操作，需要明确授权。

---

## 修订

修改本规范时同步更新版本号与 `CHANGELOG.md`，并运行 `stdsync write <repo>...` 把配置分发到各仓库。规则编号不复用：废弃的规则标注"已废弃"保留编号。

规范维护者必须为自动强制的关键规则保留正反样例；工具升级后验证违规仍被拒绝。机器检查只覆盖当前构建平台与标签，跨平台模块须在 CI 声明验证矩阵。
