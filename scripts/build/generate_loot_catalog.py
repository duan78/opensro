"""
===========================================================================
generate_loot_catalog.py - compile immutable v1.150 loot and eligibility audit

Only committed normalized evidence is read. Inferred ordinary consumable rates
are deliberately explicit here; they are not represented as recovered rates.
===========================================================================
"""

import argparse
import copy
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
SOURCE = ROOT / "scripts/data/loot"
OUTPUT = ROOT / "apps/server/internal/game/item/loot/.generated"
INFERRED_RATES = {2: 0.10, 3: 0.02, 4: 0.05, 5: 0.05, 6: 0.01}
SPECIAL_ITEMS = ["ITEM_ETC_ARCHEMY_REINFORCE_RECIPE_" + kind + "_B"
	for kind in ("WEAPON", "SHIELD", "ARMOR", "ACCESSARY")] + ["ITEM_ETC_SCROLL_RETURN_02"]


# ================
# read_source
# ================
def read_source(name):
	return json.loads((SOURCE / (name + "-source.json")).read_text(encoding="utf-8"))


# ================
# level_band
# ================
def level_band(level):
	return sum(level >= boundary for boundary in (20, 40, 60, 80, 90))


# ================
# compile_catalogs
# ================
def compile_catalogs():
	client = read_source("client")
	equipment = read_source("equipment")
	consumables = read_source("consumables")
	vsro, isro = read_source("vsro-rewards"), read_source("isro-rewards")
	if vsro["custom"] or isro["custom"]:
		raise ValueError("New applicable custom loot rules need a versioned implementation")
	items = client["items"]
	equipment_rebindings = []
	for row in equipment["items"]:
		if not row["rare"]:
			continue
		group = row["group"]
		remaining = equipment["rare"][row["level"] - 1:]
		if any(level[group] > 0.000001 for level in remaining):
			continue
		# Client-required torso levels outlive the donor's A-rare class window.
		# Reconstruction: join those assignments to the next enabled class of
		# the same degree. Preserve all class rates and the original item weight.
		for candidate in range(group + 1, (group // 3 + 1) * 3):
			if any(level[candidate] > 0.000001 for level in remaining):
				equipment_rebindings.append({"item": row["codename"], "from": group, "to": candidate,
					"reason": "client-required-level-outlives-source-class-window"})
				row["group"] = candidate
				break
		else:
			raise ValueError("Rare equipment has no compatible class: " + row["codename"])
	degrees = {(items[row["codename"]]["class"] + 2) // 3 for row in equipment["items"]}
	excluded = []
	filtered = []
	for row in consumables["items"]:
		ref = items[row["codename"]]
		if row["family"] in (8, 9) and ref["param1"] not in degrees:
			excluded.append({"item": row["codename"], "reason": "material-degree-has-no-client-equipment"})
		else:
			filtered.append(row)
	consumables["items"] = filtered
	for family, probability in INFERRED_RATES.items():
		rows = consumables["classes"][str(family)]
		for level in range(1, len(rows) + 1):
			if any(rows[level - 1]):
				raise ValueError("Authored probability conflicts with inferred family " + str(family))
			group = level_band(level)
			if family == 3:
				group = min(group, 3)
			if family == 6:
				group = 0
			rows[level - 1][group] = probability
	# An assignment alone does not enable acquisition. Preserve the source's
	# disabled classes (for example the A elixir bucket) in the audit, without
	# advertising them as reachable ordinary loot.
	active = []
	for row in consumables["items"]:
		classes = consumables["classes"][str(row["family"])]
		if any(row["group"] < len(level) and level[row["group"]] > 0.000001 for level in classes):
			active.append(row)
		else:
			excluded.append({"item": row["codename"], "reason": "source-class-disabled"})
	consumables["items"] = active
	consumables["fixed"] = []
	for row in vsro["fixed"]:
		ref = items[row["item"]]
		if ref["type"][2:] in ([11, 1], [11, 2], [11, 7]) and ref["param1"] not in degrees:
			excluded.append({"item": row["item"], "monster": row["monster"], "reason": "material-degree-has-no-client-equipment"})
		else:
			consumables["fixed"].append(copy.deepcopy(row))
	# v1.150 characterdata columns 99..108 give each monster up to five of its
	# own alchemy materials with a drop probability; no newer server table
	# carries them. INFERENCE: the server rolls each pair as a fixed
	# per-monster reward (724A00), one item at the authored probability.
	assigned = {(row["monster"], row["item"]) for row in consumables["fixed"]}
	for row in client.get("materialDrops", []):
		if (row["monster"], row["item"]) in assigned:
			continue
		assigned.add((row["monster"], row["item"]))
		consumables["fixed"].append({"monster": row["monster"], "item": row["item"], "plus": 0, "min": 1, "max": 1,
			"probability": row["probability"]})
	consumables["random"] = []
	consumables["groups"] = {}
	seen = set()
	for source in (vsro, isro):
		for row in source["random"]:
			key = (row["monster"], row["group"])
			if key in seen:
				continue
			seen.add(key)
			group = str(row["group"])
			members = []
			for member in source["groups"][group]:
				ref = items[member["codename"]]
				if ref["type"][2:] in ([11, 1], [11, 2], [11, 7]) and ref["param1"] not in degrees:
					excluded.append({"item": member["codename"], "group": int(group), "reason": "material-degree-has-no-client-equipment"})
				else:
					members.append(member)
			if not members:
				raise ValueError("Applicable random group has no eligible members")
			if group in consumables["groups"] and consumables["groups"][group] != members:
				raise ValueError("Conflicting group identities across source versions")
			consumables["groups"][group] = members
			consumables["random"].append({key: value for key, value in row.items() if key not in ("page", "record")})
	ordinary = {row["codename"] for row in equipment["items"] + consumables["items"]}
	special = set(SPECIAL_ITEMS) | {row["item"] for row in consumables["fixed"]}
	for members in consumables["groups"].values():
		special.update(row["codename"] for row in members)
	for code in ordinary | special:
		if code not in items:
			raise ValueError("Missing client item " + code)
	eligibility = []
	unavailable = {row["item"] for row in excluded}
	for code, ref in sorted(items.items()):
		if code in ordinary:
			category = "ordinary-monster"
		elif code in special:
			category = "assigned-or-special-monster"
		elif code in unavailable:
			category = next(row["reason"] for row in excluded if row["item"] == code)
		elif ref["type"][1:3] == [3, 8] or code.startswith("ITEM_EVENT_"):
			category = "quest-or-event-acquisition"
		else:
			category = "other-acquisition-no-monster-assignment"
		eligibility.append({"codename": code, "category": category})
	properties = {"items": {code: items[code] for code in sorted(ordinary | special) if items[code]["type"][1] == 1},
		"magic": client["magic"], "assignments": client["magicAssignments"]}
	audit = {"version": 1, "generatedBy": "scripts/build/generate_loot_catalog.py", "sourceHashes": client["hashes"],
		"equipmentClassRebindings": equipment_rebindings,
		"backupHashes": {"vsro": vsro["sha256"], "isro": isro["sha256"]},
		"inferredRates": INFERRED_RATES, "bands": [1, 20, 40, 60, 80, 90],
		"inferences": ["Ordinary family 7 follows the native ordinary categories using its authored class table.",
			"Monster material pairs (characterdata 99..108) roll as fixed per-monster rewards, one item each."],
		"specialItems": SPECIAL_ITEMS, "eligibility": eligibility, "excludedMaterials": excluded,
		"excludedSourceRows": {"vsro": vsro["excluded"], "isro": isro["excluded"]}}
	return {"equipment.json": equipment, "consumables.json": consumables, "properties.json": properties, "audit.json": audit}


# ================
# extended_equipment_supplement
#
# Extended content (isro-live-2026), port-only, not v1.150-native. The
# live client's degree 10-14 wearables ship as a SEPARATE supplement the
# server installs only behind the extended flag - the native equipment
# catalogue stays byte-identical, so a flag-off world drops exactly the
# native set. The supplement is validated against the native rows here:
# no codename overlap, groups within the rate-row width, windows only in
# cells the native evidence leaves at zero. Degrees 13-14 ship only
# flat-101 rare rows (the client's own pre-provisioning): no window can
# roll them, so they stay out - grant and equip still work through the
# item overlay. The window rates are inferred (the evidence file says
# so) and the audit records the whole supplement.
# ================
def extended_equipment_supplement(catalogs):
	extended_path = ROOT / "scripts/data/loot/extended-equipment-source.json"
	if not extended_path.exists():
		return None
	extended = json.loads(extended_path.read_text(encoding="utf-8"))
	equipment = catalogs["equipment.json"]
	native_codenames = {row["codename"] for row in equipment["items"]}
	overlap = sorted({row["codename"] for row in extended["items"]} & native_codenames)
	if overlap:
		raise ValueError("Extended equipment collides with native rows: " + ", ".join(overlap[:4]))
	groups = {len(row) for row in equipment["normal"]} | {len(row) for row in equipment["rare"]}
	if len(groups) != 1:
		raise ValueError("Native equipment rate rows have inconsistent widths")
	width = groups.pop()
	items = []
	pre_provisioned = 0
	for row in extended["items"]:
		if row["codename"] in native_codenames:
			continue
		if not 0 <= row["group"] < width:
			pre_provisioned += 1
			continue
		items.append(dict(row))
	windows = []
	native_windows = 0
	for window in extended["windows"]:
		level, group = window["level"], window["group"]
		if not 1 <= level <= len(equipment["normal"]):
			raise ValueError("Extended window level outside the rows: " + str(level))
		# Per kind, with the runtime's own activity threshold: the native
		# rows above 101 carry ~1e-9 epsilons (provisioned-inert windows)
		# the selection never sees, and the native rare rows stop at 97
		# while the ordinary ones run to 101 - only a genuinely active
		# native cell keeps its native rate, per kind.
		active = lambda value: value > 0.000001
		normal_kept = active(equipment["normal"][level - 1][group])
		rare_kept = active(equipment["rare"][level - 1][group])
		if normal_kept and rare_kept:
			native_windows += 1
			continue
		windows.append({"level": level, "group": group,
			"applyNormal": not normal_kept, "normalRate": extended["normalRate"],
			"applyRare": not rare_kept, "rareRate": extended["rareRate"]})
	catalogs["audit.json"]["extendedEquipment"] = {
		"sourceClient": extended["sourceClient"],
		"rates": extended["rates"],
		"itemRows": len(items),
		"preProvisionedUncatalogued": pre_provisioned,
		"windows": len(windows),
		"windowsNativeKept": native_windows,
	}
	print(
		f"extended equipment supplement: {len(items)} items, {len(windows)} window rows (rates inferred), "
		f"{native_windows} fully native windows kept, {pre_provisioned} pre-provisioned rows uncatalogued"
	)
	return {"version": 1, "sourceClient": extended["sourceClient"], "rates": extended["rates"],
		"items": items, "windows": windows}


# ================
# main
# ================
def main():
	parser = argparse.ArgumentParser(description=__doc__)
	parser.add_argument("--check", action="store_true")
	args = parser.parse_args()
	catalogs = compile_catalogs()
	supplement = extended_equipment_supplement(catalogs)
	if supplement is not None:
		catalogs["equipment-extended.json"] = supplement
	for name, value in catalogs.items():
		data = (json.dumps(value, separators=(",", ":"), ensure_ascii=True) + "\n").encode("utf-8")
		path = OUTPUT / name
		if args.check:
			if not path.exists() or path.read_bytes() != data:
				raise SystemExit("Stale loot catalog: " + str(path))
		else:
			OUTPUT.mkdir(parents=True, exist_ok=True)
			path.write_bytes(data)
	print("Loot catalogs verified" if args.check else "Loot catalogs generated")


if __name__ == "__main__":
	main()
