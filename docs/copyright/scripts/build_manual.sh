#!/usr/bin/env bash
# Build 02_软件说明书.pdf and .docx.
# Run from anywhere: bash docs/copyright/scripts/build_manual.sh
set -euo pipefail

cd "$(dirname "$0")/.."
mkdir -p build

python scripts/sync_meta.py
python scripts/prepare_manual_images.py

# HTML: Pandoc + embedded screenshots.
pandoc build/manual-optimized.md \
  -o build/manual.html \
  --standalone \
  --toc \
  --toc-depth=2 \
  --template=scripts/manual-template.html \
  --embed-resources \
  --resource-path=.

python scripts/postprocess_manual.py --html-only

# PDF: Edge headless prints the self-contained HTML.
EDGE="/c/Program Files (x86)/Microsoft/Edge/Application/msedge.exe"
if [ ! -x "$EDGE" ]; then
  EDGE="/c/Program Files/Microsoft/Edge/Application/msedge.exe"
fi
if [ -x "$EDGE" ]; then
  "$EDGE" \
    --headless=new \
    --disable-gpu \
    --no-pdf-header-footer \
    --print-to-pdf="D:/code/Travel_Guangdong/docs/copyright/build/manual.pdf" \
    "file:///D:/code/Travel_Guangdong/docs/copyright/build/manual.html"
  mv -f build/manual.pdf 02_软件说明书.pdf
  python scripts/postprocess_manual.py --pdf-only
else
  echo "Edge not found; open build/manual.html and print to PDF manually" >&2
fi

# DOCX: Pandoc with a Chinese-font reference document.
python scripts/make_reference_docx.py
pandoc build/manual-optimized.md \
  -o 02_软件说明书.docx \
  --reference-doc=scripts/reference.docx \
  --toc \
  --toc-depth=2 \
  --resource-path=.

echo "done"
