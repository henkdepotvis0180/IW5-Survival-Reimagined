"""Build a self-contained Windows launcher using Python 3 and Go >= 1.23."""
from pathlib import Path
import hashlib
import os
import subprocess
import zipfile

root = Path(__file__).resolve().parent
repo = root.parent
go = os.environ.get("GO", "go")
with zipfile.ZipFile(root / "payload.zip", "w", zipfile.ZIP_DEFLATED) as archive:
    for path in sorted((repo / "iw5").rglob("*")):
        if path.is_file():
            if path.is_symlink():
                raise RuntimeError(f"Unexpected symlink: {path}")
            # Ignore the original author's local test script, if present.
            if path.relative_to(repo).as_posix() == "iw5/lethalbeats/Survival/dev/test.gsc":
                continue
            info = zipfile.ZipInfo(path.relative_to(repo).as_posix(), (2026, 1, 1, 0, 0, 0))
            info.compress_type = zipfile.ZIP_DEFLATED
            archive.writestr(info, path.read_bytes())
subprocess.run([go, "test", "./..."], cwd=root, check=True)
out = root / "dist"
out.mkdir(exist_ok=True)
env = dict(os.environ, GOOS="windows", GOARCH="amd64", CGO_ENABLED="0")
exe = out / "SurvivalReimagined.exe"
subprocess.run([go, "build", "-trimpath", "-ldflags=-H=windowsgui -s -w", "-o", str(exe), "."], cwd=root, env=env, check=True)
(out / "SHA256SUMS.txt").write_text(hashlib.sha256(exe.read_bytes()).hexdigest() + "  " + exe.name + "\n")
print(exe)
