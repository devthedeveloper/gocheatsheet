#!/usr/bin/env python3
"""Inject verified code into the scale book. data-src="root:relpath[#func Name|Lx-Ly]"."""
import html, re
from pathlib import Path
S = Path('/workspace/gocheatsheet/scale')
ROOTS = {'final': S/'code'/'gopherit', 'baseline': S/'code'/'baseline', 'hammer': S/'code'/'hammer'}
def extract(path, frag):
    lines = path.read_text().splitlines()
    if not frag: return '\n'.join(lines).rstrip()
    m = re.fullmatch(r'L(\d+)-L(\d+)', frag)
    if m: return '\n'.join(lines[int(m[1])-1:int(m[2])]).rstrip()
    kind, name = frag.split(' ', 1)
    if kind == 'func':
        pat = re.compile(rf'^func (\([^)]*\) )?{re.escape(name)}[ (\[]')
    else:
        pat = re.compile(rf'^{kind} {re.escape(name)}\b')
    start = next((i for i,l in enumerate(lines) if pat.match(l)), None)
    if start is None: raise SystemExit(f'NOT FOUND {frag} in {path}')
    first = start
    while first>0 and lines[first-1].lstrip().startswith('//'): first-=1
    depth, end = 0, start
    for i in range(start, len(lines)):
        depth += lines[i].count('{') - lines[i].count('}')
        if depth==0 and (i>start or ('{' in lines[i] and '}' in lines[i])): end=i; break
    return '\n'.join(lines[first:end+1]).rstrip()
CODE_RE = re.compile(r'(<code class="[^"]*" data-src="([^"]+)">)(.*?)(</code>)', re.S)
total=0
for page in sorted(S.glob('ch*.html')):
    t=page.read_text(); out=[]; pos=0; n=0
    for m in CODE_RE.finditer(t):
        root,rest=m[2].split(':',1); rel,frag=(rest.split('#',1)+[None])[:2]
        code=extract(ROOTS[root]/rel, frag)
        out.append(t[pos:m.start()]); out.append(m[1]+html.escape(code,quote=False)+m[4]); pos=m.end(); n+=1
    out.append(t[pos:]); page.write_text(''.join(out)); print(f'{page.name}: {n}'); total+=n
print('TOTAL',total)
