"""Build a self-contained Windows launcher using Python 3 and Go >= 1.23."""
from pathlib import Path
import hashlib
import os
import subprocess
import zipfile
import io

root = Path(__file__).resolve().parent
repo = root.parent
go = os.environ.get("GO", "go")
# Build a coherent full official resource baseline; loose project scripts remain.
import json
baseline = Path(os.environ["IW5_OFFICIAL_DIR"])
expected = json.loads((root / "official-files.json").read_text())
files = {"iw5/" + p.relative_to(repo / "iw5").as_posix(): p
         for p in (repo / "iw5").rglob("*") if p.is_file() and not p.is_symlink()}
for name, checksum in expected.items():
    path = baseline / name
    if hashlib.sha256(path.read_bytes()).hexdigest() != checksum:
        raise RuntimeError(f"Official baseline mismatch: {name}")
    files["iw5/" + name] = path
    if name.startswith("mods/survival/") and name.endswith(".ff"):
        files["iw5/zone/" + path.name] = path
# Package project scripts in a high-priority IWD. Loose files alone can lose
# to the original LB_Survival.iwd in the engine's archive-first search order.
overrides = io.BytesIO()
with zipfile.ZipFile(overrides, "w", zipfile.ZIP_DEFLATED) as patch:
    for path in sorted((repo / "iw5").rglob("*.gsc")):
        if path.is_symlink():
            raise RuntimeError(f"Unexpected script symlink: {path}")
        name = path.relative_to(repo / "iw5").as_posix().lower()
        info = zipfile.ZipInfo(name, (2026, 1, 1, 0, 0, 0))
        info.compress_type = zipfile.ZIP_DEFLATED
        patch.writestr(info, path.read_bytes())
with zipfile.ZipFile(root / "payload.zip", "w", zipfile.ZIP_DEFLATED) as archive:
    for name, path in sorted(files.items()):
        info = zipfile.ZipInfo(name, (2026, 1, 1, 0, 0, 0))
        info.compress_type = zipfile.ZIP_DEFLATED
        archive.writestr(info, path.read_bytes())
    for name in ["iw5/zz_survival_reimagined.iwd", "iw5/mods/survival/zz_survival_reimagined.iwd"]:
        archive.writestr(name, overrides.getvalue())
subprocess.run([go, "test", "./..."], cwd=root, check=True)
out = root / "dist"
out.mkdir(exist_ok=True)
env = dict(os.environ, GOOS="windows", GOARCH="amd64", CGO_ENABLED="0")
exe = out / "SurvivalReimagined.exe"
subprocess.run([go, "build", "-trimpath", "-buildvcs=false", "-ldflags=-H=windowsgui", "-o", str(exe), "."], cwd=root, env=env, check=True)
(out / "SHA256SUMS.txt").write_text(hashlib.sha256(exe.read_bytes()).hexdigest() + "  " + exe.name + "\n")
print(exe)
package = out / "SurvivalReimagined-Windows-v0.5.0.zip"
with zipfile.ZipFile(package, "w", zipfile.ZIP_DEFLATED) as archive:
    for source, name in [
        (exe, exe.name),
        (out / "SHA256SUMS.txt", "SHA256SUMS.txt"),
        (root / "README.md", "LEESMIJ.md"),
        (repo / "TESTING.md", "TESTING.md"),
    ]:
        archive.write(source, name)
with zipfile.ZipFile(package) as archive:
    if archive.testzip() is not None:
        raise RuntimeError("ZIP integrity check failed")
print(package)
