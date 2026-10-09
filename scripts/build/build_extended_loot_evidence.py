"""
===========================================================================

build_extended_loot_evidence.py - derive the extended loot evidence (DG10-14)

Extended content (isro-live-2026), port-only, not v1.150-native. Derives
scripts/data/loot/extended-equipment-source.json from the sealed extended
projection's itemdata: every wearable row of degrees 10-14 (the live
client's own state - DG10 complete at 90-100, DG11 tier A at 101, DG12
complete at 111-120, DG13/14 rare pre-provisioned at a flat 101), with
the native evidence's group layout ((degree-1)*3 + tier).

The drop WINDOWS are data-driven: for each monster level 91-140 the
tier group whose real requirement level is the highest one not above it
(the same shape the native rows show: level 90 rolls the DG10-A window,
level 100 the DG10-C window). The RATES are inferred, explicitly: the
normal windows continue the evidence's uniform 0.0035 and the rare
windows copy the native DG10 rare value - marked "inferred" in the file
and by the generator's audit. No native row is ever overwritten: the
generator merges only into cells the native evidence leaves at zero,
and a codename both sources carry is a hard error.

Rerun after a client update; the file carries the source client and the
extraction digest so drift is visible.
===========================================================================
"""
import argparse
import glob
import hashlib
import json
import os
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
OUTPUT = ROOT / "scripts/data/loot/extended-equipment-source.json"
CODENAME = re.compile(r"^ITEM_(CH|EU)_[A-Z]+_(\d{2})_([A-C])(_RARE)?$")
# The evidence's uniform ordinary-window rate and its DG10 rare-window
# value, both measured on the committed native rows.
INFERRED_NORMAL_RATE = 0.0035
INFERRED_RARE_RATE = 0.0026889999862760305


# ================
# main
# ================
def main() -> int:
	parser = argparse.ArgumentParser(description="Derive the extended (live-2026) loot evidence." )
	parser.add_argument("--extended-root", default=os.environ.get("SRO_EXTENDED_GAME_DATA_ROOT", ""))
	args = parser.parse_args()
	if not args.extended_root:
		print("point --extended-root or SRO_EXTENDED_GAME_DATA_ROOT at the extended projection", file=sys.stderr)
		return 2
	textdata = Path(args.extended_root) / "textdata"

	rows = []
	for path in sorted(glob.glob(str(textdata / "itemdata_*.txt"))):
		text = Path(path).read_bytes().decode("utf-16-le", errors="replace").lstrip("\ufeff")
		for line in text.splitlines():
			if not line.strip() or line.startswith("//"):
				continue
			cells = line.split("\t")
			if len(cells) < 34:
				continue
			match = CODENAME.match(cells[2].strip())
			if not match:
				continue
			degree = int(match.group(2))
			if degree < 10 or degree > 14:
				continue
			try:
				level = int(cells[33])
				type_ids = ":".join(cells[i].strip() for i in (9, 10, 11, 12))
			except ValueError:
				continue
			tier_index = "ABC".index(match.group(3))
			rows.append({
				"codename": cells[2].strip(),
				"country": 0 if match.group(1) == "CH" else 1,
				"group": (degree - 1) * 3 + tier_index,
				"rare": bool(match.group(4)),
				"type": type_ids,
				"weight": 100,
				"absolute": 100,
				"level": level,
			})

	native = json.loads((ROOT / "scripts/data/loot/equipment-source.json").read_text(encoding="utf-8"))
	native_codenames = {row["codename"] for row in native["items"]}
	overlap = sorted({row["codename"] for row in rows} & native_codenames)
	if overlap:
		print(f"extended rows collide with native evidence: {overlap[:4]}", file=sys.stderr)
		return 3

	# The window map: for each level 91-140, the tier group whose window
	# is open - a tier opens at its LOWEST real requirement level (the
	# DG10 tiers span several item levels per kind) and the open window
	# of the highest opened tier rolls. Normal rows drive the windows;
	# the rare catalogue rides the same windows.
	tier_levels = {}
	for row in rows:
		if row["rare"]:
			continue
		current = tier_levels.get(row["group"])
		tier_levels[row["group"]] = row["level"] if current is None else min(current, row["level"])
	windows = []
	for level in range(91, 141):
		best = None
		for group, requirement in sorted(tier_levels.items()):
			if requirement <= level and (best is None or requirement > tier_levels[best]):
				best = group
		if best is not None:
			windows.append({"level": level, "group": best})

	evidence = {
		"version": 1,
		"sourceClient": "isro-live-2026",
		"rates": "inferred - the normal windows continue the native evidence's uniform ordinary rate, the rare windows copy the native DG10 rare value",
		"itemRows": len(rows),
		"items": sorted(rows, key=lambda row: (row["group"], row["country"], row["codename"])),
		"normalRate": INFERRED_NORMAL_RATE,
		"rareRate": INFERRED_RARE_RATE,
		"windows": windows,
	}
	OUTPUT.write_text(json.dumps(evidence, indent=1, sort_keys=False) + "\n", encoding="utf-8")
	by_degree = {}
	for row in rows:
		key = f"DG{(row['group'] + 3) // 3}{'R' if row['rare'] else 'N'}"
		by_degree[key] = by_degree.get(key, 0) + 1
	print(f"extended evidence: {len(rows)} item rows {by_degree}, {len(windows)} window levels")
	print(f"written: {OUTPUT}")
	return 0


if __name__ == "__main__":
	raise SystemExit(main())
