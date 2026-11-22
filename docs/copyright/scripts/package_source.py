#!/usr/bin/env python3
"""Package the project source tree for backup or agency review.

The official copyright source material is 03_源代码文档.pdf (60 pages).
This zip is a clean source snapshot, not the online upload file.
"""

from __future__ import annotations

import hashlib
import json
import zipfile
from pathlib import Path

BASE = Path(__file__).resolve().parent.parent
ROOT = BASE.parent.parent
OUT = BASE / "05_源代码包.zip"
SHA = BASE / "05_源代码包.sha256.txt"
PREFIX = "Travel_Guangdong_source/"

SKIP_DIRS = {
    ".git",
    "node_modules",
    "dist",
    "build",
    "target",
    "miniprogram_npm",
    "typings",
    "prototype",
    "rules",
    ".codebuddy",
    ".mimosa",
    ".zcode",
    "vendor",
}
SKIP_REL_PREFIXES = (
    "docs/copyright",
)
SKIP_REL_FILES = {
    ".env",
    "web/.env.local",
    "deploy/app.env",
}
SKIP_SUFFIXES = (
    ".exe",
    ".exe~",
    ".zip",
    ".pdf",
    ".docx",
    ".dll",
    ".so",
    ".dylib",
    ".pem",
    ".key",
)
ROOT_FILES = {
    "README.md",
    "README.website.md",
    "docker-compose.yml",
    ".env.example",
    ".gitignore",
    "package.json",
    "package-lock.json",
    "tsconfig.json",
    "project.config.json",
    "scripts/smoke_test.ps1",
    "docs/api/openapi.yaml",
    "docs/agent/README.md",
    "docs/agent/benchmark.md",
}
INCLUDE_DIRS = {
    "server-go",
    "server-java",
    "web",
    "miniprogram",
    "cloudfunctions",
    "sql",
    "deploy",
    "nginx",
}


def should_skip(rel: Path) -> bool:
    text = rel.as_posix()
    if any(text == prefix or text.startswith(prefix + "/") for prefix in SKIP_REL_PREFIXES):
        return True
    if text in SKIP_REL_FILES:
        return True
    if any(part in SKIP_DIRS for part in rel.parts):
        return True
    if text.endswith(SKIP_SUFFIXES):
        return True
    return False


def collect_files() -> list[Path]:
    files: list[Path] = []
    for path in ROOT.rglob("*"):
        if not path.is_file():
            continue
        rel = path.relative_to(ROOT)
        if should_skip(rel):
            continue
        text = rel.as_posix()
        if text in ROOT_FILES:
            files.append(path)
            continue
        if rel.parts and rel.parts[0] in INCLUDE_DIRS:
            files.append(path)
    return sorted(files)


def main() -> None:
    meta = json.loads((BASE / "meta.json").read_text(encoding="utf-8"))
    files = collect_files()
    total_bytes = sum(path.stat().st_size for path in files)
    manifest = f"""Travel_Guangdong 源程序包

软件全称：{meta['software']['full_name']}
版本号：{meta['software']['version']}
打包日期：{meta['dates']['doc_date']}

包含：
- server-go：Go 生产后端（入口、planner、agent、service、repo、handler、ai）
- web：React 游客网站（源码、构建配置、公共资源）
- miniprogram：微信小程序源码
- cloudfunctions：CloudBase 云函数
- server-java：Spring Boot 参考后端与 Mapper XML
- sql：schema、seed、migrations
- deploy / nginx：systemd、nginx、bootstrap 部署资产
- docs/api/openapi.yaml、docs/agent：接口合同与 Agent 设计
- 根目录 README、docker-compose、package.json 等构建文件

排除：
- node_modules、dist、target、miniprogram_npm、typings、prototype、rules
- .git、.codebuddy、.mimosa、.zcode 等工具元数据
- .env、web/.env.local、deploy/app.env 等真实密钥文件
- 二进制、PDF、DOCX、压缩包、私钥

说明：
软著在线提交的源程序鉴别材料是 03_源代码文档.pdf（前 30 页 + 后 30 页，共 60 页）。
本 zip 是完整工程快照，供备份、代理机构查阅或后续维护使用，不是在线提交文件。
"""

    with zipfile.ZipFile(OUT, "w", compression=zipfile.ZIP_DEFLATED, compresslevel=9) as archive:
        archive.writestr(PREFIX + "SOURCE_PACKAGE_README.txt", manifest)
        for path in files:
            rel = path.relative_to(ROOT).as_posix()
            archive.write(path, PREFIX + rel)
    digest = hashlib.sha256(OUT.read_bytes()).hexdigest()
    SHA.write_text(f"{digest}  {OUT.name}\n", encoding="utf-8")
    print(f"files: {len(files)}  source bytes: {total_bytes / 1024 / 1024:.1f} MB")
    print(f"zip: {OUT}  {OUT.stat().st_size / 1024 / 1024:.1f} MB")
    print(f"sha256: {digest}")
    roots: dict[str, int] = {}
    for path in files:
        top = path.relative_to(ROOT).parts[0]
        roots[top] = roots.get(top, 0) + 1
    print("top-level:", ", ".join(f"{k}={v}" for k, v in sorted(roots.items())))


if __name__ == "__main__":
    main()
