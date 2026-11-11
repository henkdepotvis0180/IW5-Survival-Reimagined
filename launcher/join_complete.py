from pathlib import Path
import hashlib
import json
import sys
import zipfile

root = Path(__file__).resolve().parent.parent / "downloads/complete-v0.3.0"
manifest = json.loads((root / "manifest.json").read_text())
output = Path(sys.argv[1]) / manifest["name"]
with output.open("wb") as target:
    for part in manifest["parts"]:
        data = (root / part["name"]).read_bytes()
        assert hashlib.sha256(data).hexdigest() == part["sha256"], part["name"]
        target.write(data)
assert hashlib.sha256(output.read_bytes()).hexdigest() == manifest["sha256"]
with zipfile.ZipFile(output) as z:
    assert z.testzip() is None
    exe = z.read("SurvivalReimagined.exe")
    assert hashlib.sha256(exe).hexdigest() == z.read("SHA256SUMS.txt").decode().split()[0]
output.with_suffix(".sha256").write_text(manifest["sha256"] + "  " + output.name + "\n")
print(output)
