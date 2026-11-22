#!/usr/bin/env python3
"""Create a small Pandoc reference DOCX with Chinese fonts and A4 margins."""

from __future__ import annotations

from pathlib import Path

from docx import Document
from docx.enum.text import WD_LINE_SPACING
from docx.oxml import OxmlElement
from docx.oxml.ns import qn
from docx.shared import Cm, Pt, RGBColor

BASE = Path(__file__).resolve().parent.parent
OUT = BASE / "scripts" / "reference.docx"


def set_cjk(style, ascii_name: str = "Microsoft YaHei", east_name: str = "微软雅黑") -> None:
    rpr = style.element.get_or_add_rPr()
    rfonts = rpr.find(qn("w:rFonts"))
    if rfonts is None:
        rfonts = OxmlElement("w:rFonts")
        rpr.append(rfonts)
    rfonts.set(qn("w:ascii"), ascii_name)
    rfonts.set(qn("w:hAnsi"), ascii_name)
    rfonts.set(qn("w:eastAsia"), east_name)


def main() -> None:
    doc = Document()
    for section in doc.sections:
        section.top_margin = Cm(2.0)
        section.bottom_margin = Cm(1.8)
        section.left_margin = Cm(1.8)
        section.right_margin = Cm(1.8)

    for name in ["Normal", "Body Text", "First Paragraph", "Compact"]:
        try:
            style = doc.styles[name]
        except KeyError:
            continue
        style.font.name = "Microsoft YaHei"
        style.font.size = Pt(10.5)
        set_cjk(style)
        pf = style.paragraph_format
        pf.line_spacing = 1.5
        pf.space_after = Pt(6)

    for name, size in {"Heading 1": 18, "Heading 2": 14, "Heading 3": 12, "Heading 4": 11}.items():
        try:
            style = doc.styles[name]
        except KeyError:
            continue
        style.font.name = "Microsoft YaHei"
        style.font.size = Pt(size)
        style.font.bold = True
        style.font.color.rgb = RGBColor(0, 0, 0)
        set_cjk(style)
        pf = style.paragraph_format
        pf.space_before = Pt(12)
        pf.space_after = Pt(6)
        pf.keep_with_next = True

    for name in ["Source Code", "Verbatim Char"]:
        try:
            style = doc.styles[name]
        except KeyError:
            continue
        style.font.name = "Consolas"
        style.font.size = Pt(8.2)
        set_cjk(style, ascii_name="Consolas")
        if hasattr(style, "paragraph_format"):
            style.paragraph_format.line_spacing_rule = WD_LINE_SPACING.SINGLE

    doc.save(str(OUT))
    print(f"reference docx: {OUT}")


if __name__ == "__main__":
    main()
