#!/usr/bin/env python3
"""Post-process the manual HTML and PDF produced by Pandoc/Edge.

No external process is started here:
  --html-only  inject CSS and remove Pandoc's duplicate TOC heading
  --pdf-only   add a running header and page numbers with PyMuPDF
"""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

BASE = Path(__file__).resolve().parent.parent
SCRIPTS = BASE / "scripts"
HTML = BASE / "build" / "manual.html"
PDF = BASE / "02_软件说明书.pdf"
CJK_FONT = Path(r"C:\Windows\Fonts\simhei.ttf")


def load_meta() -> dict:
    return json.loads((BASE / "meta.json").read_text(encoding="utf-8"))


def postprocess_html() -> None:
    if not HTML.exists():
        raise SystemExit(f"missing {HTML}; run pandoc first")
    html = HTML.read_text(encoding="utf-8")
    html = re.sub(r'<h2 id="toc-title">.*?</h2>\s*', "", html, flags=re.S)
    css = (SCRIPTS / "manual.css").read_text(encoding="utf-8")
    marker = "</head>"
    if marker not in html:
        raise SystemExit("manual HTML has no </head>")
    html = html.replace(marker, f"<style>\n{css}\n</style>\n{marker}", 1)
    HTML.write_text(html, encoding="utf-8")
    print(f"html ready: {HTML} ({HTML.stat().st_size / 1024:.0f} KB)")


def postprocess_pdf() -> int:
    if not PDF.exists():
        raise SystemExit(f"missing {PDF}; run the Edge print step first")
    try:
        import pymupdf
    except ImportError:
        print("pymupdf not installed; skip header/page numbers")
        return 0
    meta = load_meta()
    header = f"{meta['software']['header']}  软件说明书"
    use_cjk = CJK_FONT.exists()
    doc = pymupdf.open(PDF)
    for index, page in enumerate(doc):
        if index > 0:
            if use_cjk:
                page.insert_text(
                    (56, 30),
                    header,
                    fontsize=8,
                    fontname="simhei",
                    fontfile=str(CJK_FONT),
                    color=(0.45, 0.45, 0.45),
                )
            page.insert_text(
                (page.rect.width / 2 - 5, page.rect.height - 18),
                str(index + 1),
                fontsize=9,
                fontname="helv",
                color=(0.4, 0.4, 0.4),
            )
    tmp = PDF.with_suffix(".tmp.pdf")
    try:
        # Edge embeds full CJK fonts; keep only the glyphs actually used.
        doc.subset_fonts()
    except Exception as exc:  # pragma: no cover - depends on the PDF font format
        print(f"font subsetting skipped: {exc}")
    doc.save(tmp, garbage=4, deflate=True, clean=True)
    doc.close()
    tmp.replace(PDF)
    count = pymupdf.open(PDF).page_count
    print(f"pdf ready: {PDF} pages={count} ({PDF.stat().st_size / 1024:.0f} KB)")
    return count


def main(argv: list[str]) -> int:
    html_only = "--html-only" in argv
    pdf_only = "--pdf-only" in argv
    if html_only and pdf_only:
        raise SystemExit("choose --html-only or --pdf-only")
    if html_only:
        postprocess_html()
        return 0
    if pdf_only:
        postprocess_pdf()
        return 0
    postprocess_html()
    postprocess_pdf()
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
