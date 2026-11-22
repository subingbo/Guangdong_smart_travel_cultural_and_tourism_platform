# 软著申请材料

软件全称：南粤文化入境游 AI 智能伴游系统  
版本号：V1.0  
著作权人：subingbo  
编制日期：2026-11-10

## 直接上传的文件

| 文件 | 页数 | 大小 | 用途 |
|---|---:|---:|---|
| `02_软件说明书.pdf` | 47 | 约 2.2 MB | 文档鉴别材料（封面、目录、功能、实现、数据库、接口、部署） |
| `03_源代码文档.pdf` | 60 | 约 0.3 MB | 源程序鉴别材料（前 30 页 + 后 30 页，每页 50 行） |

说明书已做 JPEG 压缩和 PDF 字体子集化，避免超过常见的 10 MB 上传限制；源码 PDF 是纯文本，体积很小。身份证明、申请表在线填写，不在本目录。

## 源码包（备份 / 代理机构查阅）

| 文件 | 大小 | 说明 |
|---|---:|---|
| `05_源代码包.zip` | 0.4 MB | 271 个文件，完整工程快照；含 Go/Web/小程序/云函数/Java/SQL/部署资产；排除依赖、构建产物、真实密钥、第三方 typings |
| `05_源代码包.sha256.txt` | 1 KB | zip 的 SHA-256 校验 |

在线提交仍用 `03_源代码文档.pdf`（60 页）；zip 不是在线提交文件，适合备份、交给代理机构或后续维护。包内 `SOURCE_PACKAGE_README.txt` 写明了包含/排除范围。

## 申请步骤

1. 替换 `meta.json` 和 `01_申请表信息.md` 里的 6 项个人信息（身份证号、地址、邮编、电话、邮箱、日期/名称）。
2. 打开中国版权保护中心官网，注册账号并完成实名认证。
3. 进入“软件著作权登记”，按 `01_申请表信息.md` 填软件信息、著作权人信息、开发信息。
4. 上传 `03_源代码文档.pdf`、`02_软件说明书.pdf` 和身份证扫描件；在线表单里填“主要功能/技术特点”。
5. 选择普通或加急办理，按官网当期标准缴费，保存受理号。
6. 等待审查；如收到补正通知，按通知修改后重新上传；通过后下载电子证书。

详细字段、常见驳回原因和重新生成命令见 `04_提交指南.md`。找代理机构时，把两份 PDF、`05_源代码包.zip`、身份证扫描件和申请表信息一起发过去即可。

## 可编辑副本

| 文件 | 说明 |
|---|---|
| `02_软件说明书.docx` | 说明书 Word 版，含 29 个表格、13 张截图；TOC 是 Word 域，打开后右键更新 |
| `03_源代码文档.docx` | 60 页源码 Word 版；页脚是 PAGE 域，已设置打开时更新；PDF 是提交准据 |
| `02_软件说明书.md` | 说明书源稿；改文字后运行 `scripts/build_manual.sh` |
| `meta.json` | 名称、版本、著作权人、日期的唯一配置源 |

## 工作文档（不提交）

| 文件 | 说明 |
|---|---|
| `01_申请表信息.md` | 按在线表单整理的填写内容；含【身份证号】等占位符 |
| `04_提交指南.md` | 注册、上传、常见驳回原因、修改后重新生成 |
| `screenshots/` | 13 张真实界面截图（12 张正文 + 1 张异常示例） |
| `scripts/` | 构建脚本、源码顺序清单、PDF 校验、风格自检 |
| `build/` | 中间 HTML/PDF，可删除；重新生成时会重建 |
| `rendered/` | 视觉检查用的页面 PNG，可删除 |

## 关键数字

| 项 | 数值 |
|---|---:|
| 应用源码 | 25,706 行 / 207 个文件 |
| Go 生产代码 | 10,792 行 / 63 个文件（测试另计 993 行） |
| Web 前端 | 5,145 行 / 25 个文件 |
| 微信小程序 | 5,630 行 / 46 个文件 |
| Java 参考后端 | 2,080 行 / 62 个文件（含 Mapper XML） |
| 云函数 | 1,522 行 / 6 个文件 |
| 数据库表 | 16 张 |
| 业务接口 | 47 个（另有 1 个 AGENT_DEBUG 调试接口） |
| 说明书截图 | 13 张 |

统计口径：不含 `_test.go`、`node_modules`、`dist`、`target`、`miniprogram_npm`、`typings/`、`prototype/`、`docs/`、`rules/`、构建产物和二进制。

## 重新生成

```bash
cd D:/code/Travel_Guangdong

# 说明书：同步 meta.json → HTML → PDF → DOCX
bash docs/copyright/scripts/build_manual.sh

# 源代码：前 30 + 后 30 页 → HTML → PDF → DOCX → 校验
bash docs/copyright/scripts/build_source.sh

# 只做风格自检
python docs/copyright/scripts/style_check.py
```

依赖：Pandoc 3.x、Edge 或 Chrome、Python 3.12、python-docx、PyMuPDF。脚本不启动子进程执行用户输入；Pandoc/Edge 的调用固定在 `build_manual.sh` / `build_source.sh` 里。

## 提交前必须替换的 6 项

1. `meta.json` 里的 `owner.id_number`（身份证号）。
2. `owner.address`、`owner.postcode`（通讯地址、邮编）。
3. `owner.phone`、`owner.email`（电话、邮箱）。
4. `dates.dev_completed`（开发完成日期，当前 2026-11-01）。
5. `dates.first_published` 与 `dates.published`（首次发表日期/是否发表，当前 2026-11-01/已发表）。
6. 如改软件名称或版本：`software.full_name`、`short_name`、`version`，然后重跑两条构建命令。

改完 `meta.json` 后，说明书封面/页眉和源码页眉都会跟着变；`01_申请表信息.md` 里的文字需要手动同步一次。

## 提交前检查

- [ ] `03_源代码文档.pdf` 页脚从“第 1 页 共 60 页”到“第 60 页 共 60 页”。
- [ ] `02_软件说明书.pdf` 封面名称/版本/著作权人与申请表一致，共 47 页。
- [ ] 两份 PDF 里搜不到 `23ZBZ`、`sk-`、`LTAI`、`AKID`。
- [ ] 申请表、说明书封面、源码页眉三处名称和版本一致。
- [ ] 开发完成日期不晚于首次发表日期。
- [ ] 截图没有真实用户隐私；当前用种子账号 demo/alice。
- [ ] 身份证扫描件清晰、四角完整。
- [ ] 如果按“未发表”申请，申请表选“未发表”，并改 `meta.json` 的 `dates.published=false`。

## 已知情况

- 编制期间生产环境不可达（`20.194.43.11:8091` 首页 502、SSH 超时），截图来自本地隔离环境：MySQL 8.0 容器（3307）+ Go（18081）+ Vite（5174），使用 `sql/seed.sql` 的种子数据。代码与生产一致，不影响材料真实性。
- 源码样本前 30 页是 `server-go/main.go`、`internal/planner`、`internal/agent`；后 30 页是 Web 组件与页面，最后落在 `GuidePage.tsx`。小程序、云函数、SQL、部署脚本在完整工程的中间段，不进入 60 页样本。
- Java 后端是同一接口合同的参考实现，生产只运行 Go；说明书已注明，不把 Java 描述为功能完全对等。
- `docs/agent/README.md` 写明 Agent 部分从 `gdufs-campus` 迁移。**如果这部分不是本人或团队有权登记的项目**，请把 `scripts/source_manifest.json` 里的 `internal/agent`、`internal/planner` 两段删掉，重跑 `build_source.sh`，并同步修改说明书 3.3/5.2–5.4 的表述。
- 说明书 4.9 节保留了 Agent 步数超限的失败截图，用于说明错误边界；这不是正常流程。
- `03_源代码文档.docx` 的页脚页码是 Word 域，已在文档设置里开启 `updateFields`；如果打开后仍显示空白，按 `Ctrl+A` 再 `F9` 更新。提交用 PDF，不受影响。
- 本目录的材料不替代法律意见；以学校/公司名义申请前，先确认代码权属和职务作品问题。

## 本地预览（可选）

截图时启动的本地环境如果还在运行：

- 页面：`http://127.0.0.1:5174/`
- 后端：`http://127.0.0.1:18081/api/scenics`
- 演示账号：`demo / demo123`、`alice / demo123`

停止：

```bash
docker stop travel-gd-mysql
# 结束 18081/5174 上的 Go/Vite 进程即可
```
