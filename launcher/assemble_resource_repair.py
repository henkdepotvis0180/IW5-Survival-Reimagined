"""Reassemble the verified repair parts on GitHub, without a ChatGPT download."""
from pathlib import Path
import hashlib
import json
import sys
import zipfile

parts = Path(__file__).resolve().parent.parent / "downloads/resource-herstel"
output = Path(sys.argv[1])
seen = {}
for line in (parts / "SHA256SUMS.txt").read_text().splitlines():
    checksum, name = line.split()
    path = parts / name
    if path.parent != parts or hashlib.sha256(path.read_bytes()).hexdigest() != checksum:
        raise RuntimeError(f"Part checksum mismatch: {name}")
    with zipfile.ZipFile(path) as archive:
        if archive.testzip() is not None:
            raise RuntimeError(f"Corrupt ZIP: {name}")
        for item in archive.infolist():
            if item.is_dir():
                continue
            data = archive.read(item)
            if item.filename in seen and seen[item.filename] != data:
                raise RuntimeError(f"Conflicting duplicate: {item.filename}")
            seen[item.filename] = data
manifest = json.loads(seen["HERKOMST-EN-CONTROLE.json"])
for item in manifest["added_files"]:
    if hashlib.sha256(seen[item["path"]]).hexdigest() != item["sha256"]:
        raise RuntimeError(f"Resource checksum mismatch: {item['path']}")
with zipfile.ZipFile(output, "w", zipfile.ZIP_DEFLATED) as archive:
    for name, data in sorted(seen.items()):
        archive.writestr(name, data)
with zipfile.ZipFile(output) as archive:
    if archive.testzip() is not None:
        raise RuntimeError("Combined ZIP integrity check failed")
output.with_suffix(".sha256").write_text(
    hashlib.sha256(output.read_bytes()).hexdigest() + "  " + output.name + "\n"
)
print(f"Verified and combined {len(seen)} entries into {output}")
