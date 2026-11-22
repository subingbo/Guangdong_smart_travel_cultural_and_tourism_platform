#!/usr/bin/env bash
# Build 03_源代码文档.pdf and .docx (60 pages: first 30 + last 30).
set -euo pipefail

cd "$(dirname "$0")/.."
mkdir -p build

python scripts/build_source.py

EDGE="/c/Program Files (x86)/Microsoft/Edge/Application/msedge.exe"
if [ ! -x "$EDGE" ]; then
  EDGE="/c/Program Files/Microsoft/Edge/Application/msedge.exe"
fi
if [ ! -x "$EDGE" ]; then
  echo "Edge not found; open build/source.html and print to PDF manually" >&2
  exit 1
fi

"$EDGE" \
  --headless=new \
  --disable-gpu \
  --no-pdf-header-footer \
  --print-to-pdf="D:/code/Travel_Guangdong/docs/copyright/build/source.pdf" \
  "file:///D:/code/Travel_Guangdong/docs/copyright/build/source.html"

mv -f build/source.pdf 03_源代码文档.pdf
python scripts/validate_source.py
echo "done"
