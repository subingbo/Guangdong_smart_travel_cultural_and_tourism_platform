#!/usr/bin/env python3
"""Build the 60-page source-code document (HTML + DOCX).

The PDF is printed from the HTML by Edge in build_source.sh.
No external process is started from this script.
"""

from __future__ import annotations

import glob
import html
import json
import math
from pathlib import Path

BASE = Path(__file__).resolve().parent.parent
ROOT = BASE.parent.parent
SCRIPTS = BASE / "scripts"
BUILD = BASE / "build"
HTML = BUILD / "source.html"
DOCX = BASE / "03_源代码文档.docx"
MANIFEST = SCRIPTS / "source_manifest.json"

EXCLUDE_PARTS = {
    "node_modules",
    "dist",
    "build",
    "target",
    "vendor",
    ".git",
    "miniprogram_npm",
    "typings",
    "prototype",
    "docs",
    "rules",
    ".codebuddy",
    ".mimosa",
    ".zcode",
}
EXCLUDE_SUFFIXES = (
    "_test.go",
    ".exe",
    ".zip",
    ".rar",
    ".7z",
    ".png",
    ".jpg",
    ".jpeg",
    ".gif",
    ".ico",
    ".woff",
    ".woff2",
    ".ttf",
    ".dll",
    ".so",
    ".dylib",
    ".pdf",
)


def load_json(path: Path) -> dict:
    return json.loads(path.read_text(encoding="utf-8"))


def resolve_files(manifest: dict) -> list[Path]:
    files: list[Path] = []
    seen: set[Path] = set()

    def add(path: Path) -> None:
        path = path.resolve()
        if path in seen:
            return
        rel = path.relative_to(ROOT)
        if any(part in EXCLUDE_PARTS for part in rel.parts):
            return
        if path.name.endswith(EXCLUDE_SUFFIXES):
            return
        if not path.is_file():
            print(f"missing: {rel}")
            return
        seen.add(path)
        files.append(path)

    for section in manifest["sections"]:
        for rel in section.get("files", []):
            add(ROOT / rel)
        for pattern in section.get("globs", []):
            matches = sorted(glob.glob(str(ROOT / pattern), recursive=True))
            for match in matches:
                add(Path(match))
    return files


def clean_line(line: str) -> str:
    line = line.expandtabs(4).rstrip()
    return "".join(ch if ch >= " " or ch == "\t" else " " for ch in line)


def display_width(text: str) -> int:
    return sum(2 if ord(ch) > 127 else 1 for ch in text)


def wrap_line(line: str, limit: int = 100) -> list[str]:
    line = clean_line(line)
    if not line:
        return [""]
    chunks: list[str] = []
    current: list[str] = []
    width = 0
    for ch in line:
        char_width = 2 if ord(ch) > 127 else 1
        if current and width + char_width > limit:
            chunks.append("".join(current))
            current = ["    "]
            width = 4
        current.append(ch)
        width += char_width
    if current:
        chunks.append("".join(current))
    return chunks


def build_rendered_lines(files: list[Path]) -> list[str]:
    rendered: list[str] = []
    for path in files:
        rel = path.relative_to(ROOT).as_posix()
        banner = f"// ===== {rel} ====="
        rendered.extend(wrap_line(banner))
        text = path.read_text(encoding="utf-8", errors="replace")
        if text.startswith("\ufeff"):
            text = text[1:]
        for line in text.splitlines():
            rendered.extend(wrap_line(line))
    return rendered


def select_sample(rendered: list[str], per_page: int, pages: int) -> tuple[list[str], list[str], int]:
    half = per_page * (pages // 2)
    if len(rendered) <= per_page * pages:
        # The whole listing fits; keep it and pad the last page.
        total_pages = max(1, math.ceil(len(rendered) / per_page))
        first = rendered[: per_page * total_pages]
        last: list[str] = []
        return first, last, total_pages
    return rendered[:half], rendered[-half:], pages


def build_html(meta: dict, first: list[str], last: list[str], pages: int, per_page: int) -> None:
    header = meta["software"]["header"]
    head = f"""<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<title>{html.escape(header)} 源程序</title>
<style>
@page {{ size: A4; margin: 0; }}
* {{ box-sizing: border-box; -webkit-print-color-adjust: exact; print-color-adjust: exact; }}
body {{ margin: 0; }}
.page {{ width: 210mm; height: 296.6mm; padding: 10mm 12mm; page-break-after: always; position: relative; overflow: hidden; }}
.page:last-child {{ page-break-after: auto; }}
.header {{ height: 8mm; font-family: "Microsoft YaHei", sans-serif; font-size: 8.5pt; color: #333; border-bottom: 0.4pt solid #999; text-align: center; }}
pre {{ margin: 3mm 0 0; font-family: "Consolas", "Courier New", "Microsoft YaHei", monospace; font-size: 8.1pt; line-height: 13.1pt; white-space: pre; }}
.footer {{ position: absolute; bottom: 6mm; left: 0; right: 0; text-align: center; font-family: "Microsoft YaHei", sans-serif; font-size: 8pt; color: #555; }}
</style>
</head>
<body>
"""
    parts = [head]
    for page in range(pages):
        if page < pages // 2:
            chunk = first[page * per_page : (page + 1) * per_page]
        else:
            index = page - pages // 2
            chunk = last[index * per_page : (index + 1) * per_page]
        if len(chunk) < per_page:
            chunk = chunk + [""] * (per_page - len(chunk))
        code = "\n".join(html.escape(line, quote=False) for line in chunk)
        parts.append('<div class="page">')
        parts.append(f'<div class="header">{html.escape(header)} 源程序</div>')
        parts.append(f"<pre>{code}</pre>")
        parts.append(f'<div class="footer">第 {page + 1} 页 共 {pages} 页</div>')
        parts.append("</div>")
    parts.append("</body>\n</html>\n")
    HTML.parent.mkdir(parents=True, exist_ok=True)
    HTML.write_text("".join(parts), encoding="utf-8")
    print(f"html: {HTML} ({HTML.stat().st_size / 1024:.0f} KB)")


def set_cjk(run, ascii_name: str = "Consolas", east_name: str = "微软雅黑") -> None:
    from docx.oxml.ns import qn

    run.font.name = ascii_name
    rpr = run._element.get_or_add_rPr()
    rfonts = rpr.find(qn("w:rFonts"))
    if rfonts is None:
        from docx.oxml import OxmlElement

        rfonts = OxmlElement("w:rFonts")
        rpr.append(rfonts)
    rfonts.set(qn("w:ascii"), ascii_name)
    rfonts.set(qn("w:hAnsi"), ascii_name)
    rfonts.set(qn("w:eastAsia"), east_name)


def add_page_field(paragraph) -> None:
    from docx.oxml import OxmlElement
    from docx.oxml.ns import qn

    run = paragraph.add_run()
    begin = OxmlElement("w:fldChar")
    begin.set(qn("w:fldCharType"), "begin")
    instr = OxmlElement("w:instrText")
    instr.set(qn("xml:space"), "preserve")
    instr.text = "PAGE"
    end = OxmlElement("w:fldChar")
    end.set(qn("w:fldCharType"), "end")
    run._r.append(begin)
    run._r.append(instr)
    run._r.append(end)


def build_docx(meta: dict, first: list[str], last: list[str], pages: int, per_page: int) -> None:
    from docx import Document
    from docx.enum.text import WD_ALIGN_PARAGRAPH, WD_BREAK
    from docx.shared import Cm, Pt

    header = meta["software"]["header"]
    doc = Document()
    section = doc.sections[0]
    section.page_width = Cm(21.0)
    section.page_height = Cm(29.7)
    section.top_margin = Cm(1.0)
    section.bottom_margin = Cm(1.0)
    section.left_margin = Cm(1.2)
    section.right_margin = Cm(1.2)

    hp = section.header.paragraphs[0]
    hp.alignment = WD_ALIGN_PARAGRAPH.CENTER
    hr = hp.add_run(f"{header} 源程序")
    hr.font.size = Pt(8.5)
    set_cjk(hr, ascii_name="Microsoft YaHei", east_name="微软雅黑")

    fp = section.footer.paragraphs[0]
    fp.alignment = WD_ALIGN_PARAGRAPH.CENTER
    prefix = fp.add_run("第 ")
    prefix.font.size = Pt(8)
    add_page_field(fp)
    suffix = fp.add_run(f" 页 共 {pages} 页")
    suffix.font.size = Pt(8)
    for run in (prefix, suffix):
        set_cjk(run, ascii_name="Microsoft YaHei", east_name="微软雅黑")

    for page in range(pages):
        if page < pages // 2:
            chunk = first[page * per_page : (page + 1) * per_page]
        else:
            index = page - pages // 2
            chunk = last[index * per_page : (index + 1) * per_page]
        if len(chunk) < per_page:
            chunk = chunk + [""] * (per_page - len(chunk))
        for line in chunk:
            p = doc.add_paragraph()
            pf = p.paragraph_format
            pf.space_before = Pt(0)
            pf.space_after = Pt(0)
            pf.line_spacing = Pt(13.1)
            run = p.add_run(line)
            run.font.size = Pt(8.1)
            set_cjk(run)
        if page < pages - 1:
            last_p = doc.paragraphs[-1]
            last_p.add_run().add_break(WD_BREAK.PAGE)

    # Ask Word to refresh the PAGE field when the document is opened.
    from docx.oxml import OxmlElement
    from docx.oxml.ns import qn

    settings = doc.settings.element
    update = OxmlElement("w:updateFields")
    update.set(qn("w:val"), "true")
    settings.append(update)

    doc.save(str(DOCX))
    print(f"docx: {DOCX} ({DOCX.stat().st_size / 1024:.0f} KB)")


def main() -> None:
    meta = load_json(BASE / "meta.json")
    manifest = load_json(MANIFEST)
    files = resolve_files(manifest)
    rendered = build_rendered_lines(files)
    per_page = int(meta["document"]["lines_per_page"])
    pages = int(meta["document"]["source_pages"])
    first, last, actual_pages = select_sample(rendered, per_page, pages)
    print(f"files: {len(files)}  rendered lines: {len(rendered)}  pages: {actual_pages}")
    print(f"first file: {files[0].relative_to(ROOT).as_posix() if files else '-'}")
    print(f"last file: {files[-1].relative_to(ROOT).as_posix() if files else '-'}")
    build_html(meta, first, last, actual_pages, per_page)
    build_docx(meta, first, last, actual_pages, per_page)


if __name__ == "__main__":
    main()
