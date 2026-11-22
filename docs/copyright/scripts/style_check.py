#!/usr/bin/env python3
"""Check the copyright documents for AI-ish style markers.

Targets: forbidden marketing words, emoji, bold density, placeholders and
unfinished TODOs. The application form is allowed to contain 【...】 fields.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

BASE = Path(__file__).resolve().parent.parent
FILES = [
    BASE / "02_软件说明书.md",
    BASE / "01_申请表信息.md",
    BASE / "04_提交指南.md",
]

FORBIDDEN = [
    "赋能",
    "旨在",
    "打造",
    "一站式",
    "沉浸式",
    "全方位",
    "深度融合",
    "无限可能",
    "核心价值主张",
    "设计原则",
    "总结与展望",
]
FORBIDDEN_RE = [re.compile(r"开启.{0,8}之旅")]

EMOJI_RE = re.compile(
    "[\U0001F300-\U0001FAFF\U00002600-\U000027BF\U0001F1E6-\U0001F1FF]"
)
TODO_RE = re.compile(r"\bTODO\b|\bFIXME\b|待补充")
PLACEHOLDER_RE = re.compile(r"【[^】]+】")


def check(path: Path) -> tuple[list[str], dict]:
    text = path.read_text(encoding="utf-8")
    # Ignore fenced code blocks for prose metrics.
    prose = re.sub(r"```.*?```", "", text, flags=re.S)
    chars = len(re.sub(r"\s", "", prose))
    issues: list[str] = []
    for word in FORBIDDEN:
        count = prose.count(word)
        if count:
            issues.append(f"forbidden word '{word}' x{count}")
    for pattern in FORBIDDEN_RE:
        for match in pattern.findall(prose):
            issues.append(f"forbidden phrase '{match}'")
    emoji = EMOJI_RE.findall(prose)
    if emoji:
        issues.append(f"emoji x{len(emoji)}: {' '.join(emoji[:5])}")
    bold = prose.count("**") // 2
    density = bold * 1000 / chars if chars else 0
    if density > 8:
        issues.append(f"bold density {density:.1f}/1000 chars (target <8)")
    todos = TODO_RE.findall(prose)
    if todos:
        issues.append(f"unfinished markers x{len(todos)}: {' '.join(todos[:5])}")
    placeholders = PLACEHOLDER_RE.findall(prose)
    stats = {
        "chars": chars,
        "bold": bold,
        "bold_density": density,
        "placeholders": placeholders,
        "headings": len(re.findall(r"^#{1,3} ", prose, flags=re.M)),
        "images": len(re.findall(r"!\[", text)),
        "tables": len(re.findall(r"^\|.*\|$", prose, flags=re.M)),
    }
    return issues, stats


def main() -> int:
    failed = False
    for path in FILES:
        if not path.exists():
            print(f"MISSING: {path}")
            failed = True
            continue
        issues, stats = check(path)
        # Placeholders are expected in the application form.
        if path.name.startswith("01_"):
            issues = [i for i in issues if not i.startswith("unfinished markers")]
        print(f"\n{path.name}")
        print(
            f"  chars={stats['chars']} bold={stats['bold']} "
            f"density={stats['bold_density']:.1f}/1000 "
            f"headings={stats['headings']} images={stats['images']} tables={stats['tables']}"
        )
        if stats["placeholders"]:
            print(f"  placeholders: {', '.join(stats['placeholders'][:8])}")
        if issues:
            failed = True
            for issue in issues:
                print(f"  FAIL: {issue}")
        else:
            print("  OK")
    print("\nstyle check:", "FAIL" if failed else "PASS")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
