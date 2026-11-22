#!/usr/bin/env python3
"""Create compressed JPEG copies of the manual screenshots.

The original Markdown keeps the PNG paths. Pandoc uses the generated
build/manual-optimized.md, which points at build/images/*.jpg. This keeps the
PDF under the usual 10 MB upload limit without changing the layout.
"""

from __future__ import annotations

import re
from pathlib import Path

from PIL import Image

BASE = Path(__file__).resolve().parent.parent
MD = BASE / "02_软件说明书.md"
OUT_MD = BASE / "build" / "manual-optimized.md"
OUT_DIR = BASE / "build" / "images"

IMAGE_RE = re.compile(r"screenshots/([A-Za-z0-9_.-]+)\.png")


def main() -> None:
    OUT_DIR.mkdir(parents=True, exist_ok=True)
    text = MD.read_text(encoding="utf-8")
    names = sorted(set(IMAGE_RE.findall(text)))
    if not names:
        raise SystemExit("no screenshot references found")
    total_before = 0
    total_after = 0
    for name in names:
        src = BASE / "screenshots" / f"{name}.png"
        if not src.exists():
            raise SystemExit(f"missing screenshot: {src}")
        dst = OUT_DIR / f"{name}.jpg"
        with Image.open(src) as image:
            total_before += src.stat().st_size
            rgb = image.convert("RGB")
            if rgb.width > 1440:
                height = round(rgb.height * 1440 / rgb.width)
                rgb = rgb.resize((1440, height), Image.LANCZOS)
            rgb.save(dst, "JPEG", quality=82, optimize=True, progressive=True)
            total_after += dst.stat().st_size
            print(f"{name}: {image.size[0]}x{image.size[1]} -> {rgb.width}x{rgb.height} "
                  f"{src.stat().st_size / 1024:.0f}KB -> {dst.stat().st_size / 1024:.0f}KB")
    text = IMAGE_RE.sub(lambda m: f"build/images/{m.group(1)}.jpg", text)
    OUT_MD.write_text(text, encoding="utf-8")
    print(f"images: {len(names)}  total {total_before / 1024 / 1024:.1f}MB -> "
          f"{total_after / 1024 / 1024:.1f}MB")
    print(f"optimized markdown: {OUT_MD}")


if __name__ == "__main__":
    main()
