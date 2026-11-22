#!/usr/bin/env python3
"""Sync the manual YAML metadata from meta.json.

This keeps the software name, owner, version and date in one place.
"""

from __future__ import annotations

import json
import re
from pathlib import Path

BASE = Path(__file__).resolve().parent.parent
MD = BASE / "02_软件说明书.md"


def main() -> None:
    meta = json.loads((BASE / "meta.json").read_text(encoding="utf-8"))
    software = meta["software"]
    owner = meta["owner"]
    dates = meta["dates"]
    yaml_block = (
        "---\n"
        f'title: "{software["full_name"]}"\n'
        'subtitle: "软件说明书"\n'
        f'author: "{owner["name"]}"\n'
        f'date: "{dates["doc_date"]}"\n'
        f'version: "{software["version"]}"\n'
        "lang: zh-CN\n"
        "---\n"
    )
    text = MD.read_text(encoding="utf-8")
    updated, count = re.subn(r"^---\s*\n.*?\n---\s*\n", yaml_block, text, count=1, flags=re.S)
    if count != 1:
        raise SystemExit("manual YAML header not found")
    MD.write_text(updated, encoding="utf-8")
    print(f"synced metadata: {software['full_name']} {software['version']} / {owner['name']}")


if __name__ == "__main__":
    main()
