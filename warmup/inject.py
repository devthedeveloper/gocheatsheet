#!/usr/bin/env python3
"""Inject verified checkpoint code into the warm-up book's <code data-src=...> blocks.

data-src format:  "checkpoints:<relpath>"   e.g. "checkpoints:ch01/main.go"
Whole files only (the book always shows complete files — no fragments).
"""
import html
import re
from pathlib import Path

WARMUP = Path('/workspace/gocheatsheet/warmup')
CODE = WARMUP / 'code'

CODE_RE = re.compile(r'(<code class="[^"]*" data-src="([^"]+)">)(.*?)(</code>)', re.S)

def extract(src: str) -> str:
    root, rel = src.split(':', 1)
    assert root == 'checkpoints', root
    return (CODE / 'checkpoints' / rel).read_text().rstrip()

total = 0
for page in sorted(WARMUP.glob('ch*.html')):
    text = page.read_text()
    out, n, pos = [], 0, 0
    for m in CODE_RE.finditer(text):
        code = extract(m.group(2))
        out.append(text[pos:m.start()])
        out.append(m.group(1) + html.escape(code, quote=False) + m.group(4))
        pos = m.end()
        n += 1
    out.append(text[pos:])
    page.write_text(''.join(out))
    print(f'{page.name}: {n} blocks injected')
    total += n
print(f'TOTAL: {total}')
