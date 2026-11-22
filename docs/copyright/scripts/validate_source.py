#!/usr/bin/env python3
"""Validate the printed source-code PDF: page count, header, sample range, no keys."""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

import pymupdf

BASE = Path(__file__).resolve().parent.parent
PDF = BASE / "03_源代码文档.pdf"
SECRET_PATTERNS = [
    re.compile(r"23ZBZ-[A-Z0-9-]+"),
    re.compile(r"sk-[A-Za-z0-9]{20,}"),
    re.compile(r"LTAI[A-Za-z0-9]{12,}"),
    re.compile(r"AKID[A-Za-z0-9]{16,}"),
]


def main() -> int:
    meta = json.loads((BASE / "meta.json").read_text(encoding="utf-8"))
    expected = int(meta["document"]["source_pages"])
    header = meta["software"]["header"]
    if not PDF.exists():
        print(f"missing {PDF}")
        return 1
    doc = pymupdf.open(PDF)
    errors: list[str] = []
    if doc.page_count != expected:
        errors.append(f"page count {doc.page_count} != {expected}")

    texts = [page.get_text() for page in doc]
    for index, text in enumerate(texts):
        if header not in text:
            errors.append(f"page {index + 1}: header missing")
            break
    if "第 1 页 共" not in texts[0]:
        errors.append("page 1: footer missing")
    if f"第 {expected} 页 共 {expected} 页" not in texts[-1]:
        errors.append(f"page {expected}: footer missing")

    if "main.go" not in texts[0]:
        errors.append("page 1: main.go not present")
    first_half = "\n".join(texts[: expected // 2])
    last_half = "\n".join(texts[expected // 2 :])
    if "planner" not in first_half and "agent" not in first_half:
        errors.append("first 30 pages: planner/agent not present")
    if "web/src" not in last_half:
        errors.append("last 30 pages: web/src not present")

    joined = "\n".join(texts)
    for pattern in SECRET_PATTERNS:
        for match in pattern.findall(joined):
            errors.append(f"possible secret in PDF: {match[:12]}...")

    print(f"pages: {doc.page_count}")
    print(f"size: {PDF.stat().st_size / 1024:.0f} KB")
    print(f"first page text: {texts[0].splitlines()[:3]}")
    print(f"last page text tail: {texts[-1].splitlines()[-4:]}")
    if errors:
        for error in errors:
            print("ERROR:", error)
        return 1
    print("source PDF validation: OK")
    return 0


if __name__ == "__main__":
    sys.exit(main())
