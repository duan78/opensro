"""
===========================================================================

extract_extended_world_data.py - pull the live-2026 world plane (extended)

Extended content (isro-live-2026), port-only, not v1.150-native. For the
regions the sealed extended zone list (zones.json, milestone M4 part 1)
names, extract the complete outdoor data plane the world lane builds from:

  Map.pk2   <y>/<x>.m/.t/.o2 per region, object.ifo, tile2d.ifo and the
            whole tile2d texture directory (the terrain tiles reference
            into it by catalog id);
  Data.pk2  navmesh/nv_<hex>.nvm per region, navmesh/mapinfo.mfo,
            navmesh/objectstring.ifo, and the transitive object-resource
            closure the .o2 placements reference (compound -> BSR -> BMT
            material sets + BMS meshes -> DDJ textures);
  Media.pk2 the outdoor minimap tiles minimap/<x>x<z>.ddj per region (a
            region without a tile keeps no minimap art, recorded not
            fatal).

The retail install is read strictly read-only (mmap; nothing is ever
written under the game root). Output mirrors the native game-root/extracted
contract (Map_extracted/, Data_extracted/) under
<generated>/extended/game/extracted, so the world builders record
game-relative provenance exactly as the native lane does. File names are
written folded (ASCII-case), the identity the pk2 readers and the asset
builders resolve by.

A region whose map plane is absent from the client is skipped and recorded,
never invented; a skipped region that would empty a trajectory band fails
the run. A closure reference the archives do not carry fails the run: the
extended world never ships a placement whose resources are missing.

===========================================================================
"""
import argparse
import hashlib
import json
import mmap
import os
import re
import struct
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import sro_paths
import sro_pk2

SOURCE_CLIENT = "isro-live-2026"
DEFAULT_GAME_ROOT = r"C:\Program Files (x86)\Silkroad"
DEFAULT_ZONES_RELATIVE = Path("extended") / "zones.json"

# The o2 placement grammar (scripts/build/world/jmx/JMXVMAPO1001.mjs is the
# owner): 12-byte signature, then a 6x6 block grid with 4 LOD groups each,
# every slot a u16 count of 30-byte placement records whose first u32 is the
# object id. Only the object ids and the full-consumption check are needed
# here; the world builder re-parses everything.
MAPO_SIGNATURE = b"JMXVMAPO1001"
MAPO_BLOCKS_PER_AXIS = 6
MAPO_LOD_GROUPS = 4
MAPO_PLACEMENT_BYTES = 30

# Resource signatures (scripts/build/world/objects/formats.mjs owns the
# layouts this walker mirrors; the builder re-parses every file itself).
CPD_SIGNATURE = b"JMXVCPD 0101"
BSR_SIGNATURE = b"JMXVRES 0109"
BMT_SIGNATURE = b"JMXVBMT 0102"
SIGNATURE_BYTES = 12
MAX_BSR_MATERIALS = 256
MAX_BSR_MESHES = 512
MAX_BMT_MATERIALS = 1024
MAX_COMPOUND_BRANCHES = 512

BMS_PATH_SHAPE = re.compile(r"(?:^|/)prim/mesh/.+\.bms$")
COMPOUND_BRANCH_SHAPE = re.compile(r"^res/.*\.bsr$")
OBJECT_IFO_ROW = re.compile(r"^(\d+)\s+(0x[0-9a-fA-F]+|\d+)\s+\"(.*)\"\s*$")


# ================
# normalize_asset_path
#
# The python twin of scripts/build/shared/assetPaths.mjs: ASCII-only case
# folding, backslashes to forward slashes, no leading or doubled slashes.
# ================
def normalize_asset_path( value ):
	folded = "".join( chr( ord( c ) + 32 ) if "A" <= c <= "Z" else c for c in str( value ).strip() )
	return "/" .join( part for part in folded.replace( "\\", "/" ).split( "/" ) if part )


# ================
# counted_string
#
# One u32 byte length followed by that many CP949 bytes.
# ================
def counted_string( payload, offset ):
	if offset + 4 > len( payload ):
		raise ValueError( f"counted string header at {offset} past {len( payload )} bytes" )
	( length, ) = struct.unpack_from( "<I", payload, offset )
	start = offset + 4
	if start + length > len( payload ):
		raise ValueError( f"counted string body at {start}+{length} past {len( payload )} bytes" )
	text = payload[start:start + length].decode( "cp949", errors="replace" )
	return text, start + length


# ================
# parse_o2_object_ids
#
# The set of object ids one .o2 placement file references, with the same
# full-consumption contract the native parser enforces.
# ================
def parse_o2_object_ids( payload, name ):
	if payload[:SIGNATURE_BYTES] != MAPO_SIGNATURE:
		raise ValueError( f"{name}: expected JMXVMAPO1001 signature" )
	offset = SIGNATURE_BYTES
	object_ids = set()
	placements = 0
	for _block in range( MAPO_BLOCKS_PER_AXIS * MAPO_BLOCKS_PER_AXIS ):
		for _lod in range( MAPO_LOD_GROUPS ):
			if offset + 2 > len( payload ):
				raise ValueError( f"{name}: truncated o2 slot header at {offset}" )
			( count, ) = struct.unpack_from( "<H", payload, offset )
			offset += 2
			end = offset + count * MAPO_PLACEMENT_BYTES
			if end > len( payload ):
				raise ValueError( f"{name}: truncated o2 placements at {offset}" )
			for _placement in range( count ):
				( object_id, ) = struct.unpack_from( "<I", payload, offset )
				object_ids.add( object_id )
				offset += MAPO_PLACEMENT_BYTES
			placements += count
	if offset != len( payload ):
		raise ValueError( f"{name}: consumed {offset} of {len( payload )} bytes" )
	return object_ids, placements


# ================
# parse_object_info
#
# object.ifo rows: <id> <flags> "<resource path>" (JMXVOBJI1000.mjs).
# ================
def parse_object_info( text ):
	lines = [line.strip() for line in text.splitlines() if line.strip()]
	if not lines or lines[0] != "JMXVOBJI1000":
		raise ValueError( f"object.ifo: expected JMXVOBJI1000 signature, got {lines[0] if lines else '<empty>'}" )
	declared = int( lines[1] )
	entries = {}
	for line in lines[2:]:
		match = OBJECT_IFO_ROW.match( line )
		if not match:
			raise ValueError( f"object.ifo: invalid row {line!r}" )
		entries[int( match.group( 1 ) )] = normalize_asset_path( match.group( 3 ) )
	if len( entries ) != declared:
		raise ValueError( f"object.ifo: declared {declared} rows but parsed {len( entries )}" )
	return entries


# ================
# parse_bsr_resource_paths
#
# The material-set and mesh paths one BSR names (formats.mjs
# parseJmxResourceBsr): 13 header u32s at 0x0c, material section at [0],
# render-mesh section at [1], optional counted primary mesh at [7].
# ================
def parse_bsr_resource_paths( payload, name ):
	if payload[:SIGNATURE_BYTES] != BSR_SIGNATURE:
		raise ValueError( f"{name}: expected JMXVRES 0109 signature" )
	header = struct.unpack_from( "<13I", payload, 0x0C )

	def section_strings( offset, budget, label ):
		( count, ) = struct.unpack_from( "<I", payload, offset )
		offset += 4
		if count > budget:
			raise ValueError( f"{name}: suspicious {label} count {count}" )
		values = []
		for _ in range( count ):
			( set_id, ) = struct.unpack_from( "<I", payload, offset )
			offset += 4
			text, offset = counted_string( payload, offset )
			values.append( ( set_id, text ) )
		return offset, values

	offset, materials = section_strings( header[0], MAX_BSR_MATERIALS, "BSR material" )
	if offset > len( payload ):
		raise ValueError( f"{name}: BSR material section past end" )
	material_paths = [normalize_asset_path( text ) for _set_id, text in materials]

	# Mesh entries may carry a leading u32 flag before the path (China
	# JMXVRES 0109 quirk; the flag is accepted only when the direct read is
	# not a mesh path - same contract as readBsrMeshEntry).
	( count, ) = struct.unpack_from( "<I", payload, header[1] )
	offset = header[1] + 4
	if count > MAX_BSR_MESHES:
		raise ValueError( f"{name}: suspicious BSR mesh count {count}" )
	mesh_paths = []
	for _ in range( count ):
		entry_start = offset
		text, offset = counted_string( payload, offset )
		if not BMS_PATH_SHAPE.search( normalize_asset_path( text ) ):
			# The entry's leading u32 is a flag, not this string's length:
			# re-read the flag at the entry start and the path after it
			# (readBsrMeshEntry's flagged fallback).
			( _flag, ) = struct.unpack_from( "<I", payload, entry_start )
			text, offset = counted_string( payload, entry_start + 4 )
			if not BMS_PATH_SHAPE.search( normalize_asset_path( text ) ):
				raise ValueError( f"{name}: BSR mesh entry is neither direct nor flagged BMS path" )
		mesh_paths.append( normalize_asset_path( text ) )

	primary = None
	if header[7] and header[7] + 4 <= len( payload ):
		text, next_offset = counted_string( payload, header[7] )
		if text and next_offset <= len( payload ):
			primary = normalize_asset_path( text )
	if primary:
		mesh_paths.append( primary )
	return material_paths, mesh_paths


# ================
# parse_bmt_texture_paths
#
# The texture names one BMT material set carries, resolved beside the set
# (formats.mjs parseJmxBmtMaterialSet + resolveBmtTexturePath).
# ================
def parse_bmt_texture_paths( payload, name ):
	if payload[:SIGNATURE_BYTES] != BMT_SIGNATURE:
		raise ValueError( f"{name}: expected JMXVBMT 0102 signature" )
	offset = SIGNATURE_BYTES
	( count, ) = struct.unpack_from( "<I", payload, offset )
	offset += 4
	if count > MAX_BMT_MATERIALS:
		raise ValueError( f"{name}: suspicious BMT material count {count}" )
	paths = []
	base = normalize_asset_path( name ).rsplit( "/", 1 )[0]
	for _ in range( count ):
		_name, offset = counted_string( payload, offset )
		offset += 64  # four float4 color groups
		offset += 8  # alpha reference + flags
		texture, offset = counted_string( payload, offset )
		offset += 7  # scale float + three render-state bytes
		normalized = normalize_asset_path( texture )
		if not normalized:
			continue
		paths.append( normalized if "/" in normalized else f"{base}/{normalized}" )
	if offset != len( payload ):
		raise ValueError( f"{name}: BMT consumed {offset} of {len( payload )} bytes" )
	return paths


# ================
# parse_compound_branches
#
# The BSR branches a compound unions (compound.mjs parseCompound).
# ================
def parse_compound_branches( payload, name ):
	if payload[:SIGNATURE_BYTES] != CPD_SIGNATURE:
		raise ValueError( f"{name}: expected JMXVCPD 0101 signature" )
	name_offset, branch_offset = struct.unpack_from( "<II", payload, 0x0C)
	if name_offset < 40 or branch_offset < name_offset + 4:
		raise ValueError( f"{name}: invalid compound offsets" )
	_name, _ = counted_string( payload, name_offset )
	( count, ) = struct.unpack_from( "<I", payload, branch_offset )
	if count > MAX_COMPOUND_BRANCHES:
		raise ValueError( f"{name}: suspicious compound branch count {count}" )
	offset = branch_offset + 4
	branches = []
	for _ in range( count ):
		text, offset = counted_string( payload, offset )
		normalized = normalize_asset_path( text )
		if not COMPOUND_BRANCH_SHAPE.match( normalized ):
			raise ValueError( f"{name}: invalid compound branch {text!r}" )
		branches.append( normalized )
	if offset != len( payload ):
		raise ValueError( f"{name}: compound trailing section at {offset}/{len( payload )}" )
	return branches


# ================
# archive_digest
#
# Reuse the recorded digest while (size, mtime) match, exactly as the
# textdata extractor seals archive identity.
# ================
def archive_digest( path, previous ):
	stat = path.stat()
	if previous.get( path.name ) and previous[path.name]["bytes"] == stat.st_size \
		and previous[path.name]["mtimeNs"] == stat.st_mtime_ns:
		return previous[path.name]["sha256"], stat
	digest = hashlib.sha256()
	with open( path, "rb" ) as handle:
		for chunk in iter( lambda: handle.read( 4 * 1024 * 1024 ), b"" ):
			digest.update( chunk )
	return digest.hexdigest(), stat


def main() -> int:
	parser = argparse.ArgumentParser( description = "Extract the extended (live-2026) world data plane." )
	parser.add_argument( "--game-root", default = os.environ.get( "SRO_EXTENDED_GAME_ROOT", DEFAULT_GAME_ROOT ) )
	parser.add_argument(
		"--zones",
		default = None,
		help = "The sealed zone list (zones.json); defaults to SRO_EXTENDED_GAME_DATA_ROOT or the module default beside it.",
	)
	parser.add_argument( "--output", default = None, help = "Defaults to <generated>/extended/game/extracted." )
	args = parser.parse_args()

	game_root = Path( args.game_root )
	zones_path = Path( args.zones ) if args.zones else None
	if zones_path is None:
		extended_root = os.environ.get( "SRO_EXTENDED_GAME_DATA_ROOT", "" ).strip()
		if extended_root:
			zones_path = Path( extended_root ) / "zones.json"
		else:
			zones_path = sro_paths.GENERATED_ROOT.parent / "server-game-data" / "extended" / "zones.json"
	output = Path( args.output ) if args.output else sro_paths.GENERATED_ROOT / "extended" / "game" / "extracted"

	for name in ( "Map.pk2", "Data.pk2" ):
		if not ( game_root / name ).is_file():
			print( f"missing archive {game_root / name}; point --game-root at the live-2026 install", file = sys.stderr )
			return 2
	if not zones_path.is_file():
		print( f"zone list not found: {zones_path}", file = sys.stderr )
		return 2

	zones = json.loads( zones_path.read_text( encoding = "utf-8" ) )
	if zones.get( "format" ) != "sro-extended-zone-list":
		print( f"{zones_path} is not a sealed extended zone list", file = sys.stderr )
		return 2

	inventory_path = output.parent / "world-inventory.json"
	previous = {}
	if inventory_path.is_file():
		try:
			previous = json.loads( inventory_path.read_text( encoding = "utf-8" ) ).get( "archives", {} )
		except ( OSError, ValueError ):
			previous = {}

	archives = {}
	for name in ( "Map.pk2", "Data.pk2" ):
		sha256, stat = archive_digest( game_root / name, previous )
		archives[name] = {"bytes": stat.st_size, "mtimeNs": stat.st_mtime_ns, "sha256": sha256}
		print( f"archive {name}: {stat.st_size} bytes, sha256 {sha256[:16]}..." )

	map_root_out = output / "Map_extracted"
	data_root_out = output / "Data_extracted"

	files = {"mapSector": 0, "navmesh": 0, "tile2d": 0, "objectResource": 0, "minimap": 0}
	bytes_out = {"mapSector": 0, "navmesh": 0, "tile2d": 0, "objectResource": 0, "minimap": 0}

	# ================
	# store
	#
	# Write one extracted payload under its folded archive path. The
	# category only feeds the inventory counters.
	# ================
	def store( category, folded_root, folded_rel, payload ):
		target = folded_root / folded_rel
		target.parent.mkdir( parents = True, exist_ok = True )
		target.write_bytes( payload )
		files[category] += 1
		bytes_out[category] += len( payload )

	with open( game_root / "Map.pk2", "rb" ) as map_handle, open( game_root / "Data.pk2", "rb" ) as data_handle, \
		open( game_root / "Media.pk2", "rb" ) as media_handle:
		map_memory = mmap.mmap( map_handle.fileno(), 0, access = mmap.ACCESS_READ )
		data_memory = mmap.mmap( data_handle.fileno(), 0, access = mmap.ACCESS_READ )
		media_memory = mmap.mmap( media_handle.fileno(), 0, access = mmap.ACCESS_READ )
		try:
			map_entries = {sro_pk2.fold_ascii( e.path ): e for e in sro_pk2.read_directory( map_memory )}
			data_entries = {sro_pk2.fold_ascii( e.path ): e for e in sro_pk2.read_directory( data_memory )}
			media_entries = {sro_pk2.fold_ascii( e.path ): e for e in sro_pk2.read_directory( media_memory )}

			def map_payload( folded ):
				entry = map_entries.get( folded )
				return None if entry is None else sro_pk2.payload( map_memory, entry )

			def data_payload( folded ):
				entry = data_entries.get( folded )
				return None if entry is None else sro_pk2.payload( data_memory, entry )

			# -- the trajectory regions ------------------------------------
			extracted_regions = []
			skipped_regions = []
			object_ids = set()
			placement_count = 0
			for region_text in sorted( zones["zones"] ):
				region_id = int( region_text, 16 )
				sector_y, sector_x = ( region_id >> 8 ) & 0xFF, region_id & 0xFF
				sector_paths = [f"{sector_y}/{sector_x}.m", f"{sector_y}/{sector_x}.t", f"{sector_y}/{sector_x}.o2"]
				nvm_path = f"navmesh/nv_{region_id:04x}.nvm"
				if any( p not in map_entries for p in sector_paths ) or nvm_path not in data_entries:
					skipped_regions.append( region_text )
					continue
				for rel in sector_paths:
					store( "mapSector", map_root_out, rel, map_payload( rel ) )
				store( "navmesh", data_root_out, nvm_path, data_payload( nvm_path ) )
				object_ids_here, placements_here = parse_o2_object_ids(
					sro_pk2.payload( map_memory, map_entries[f"{sector_y}/{sector_x}.o2"] ),
					f"{sector_y}/{sector_x}.o2",
				)
				object_ids |= object_ids_here
				placement_count += placements_here
				extracted_regions.append( region_text )
				if len( extracted_regions ) % 100 == 0:
					print( f"regions: {len( extracted_regions )} extracted, {len( skipped_regions )} skipped" )

			# A skipped region never empties a trajectory band: the world
			# lane must serve every band the sealed list covers.
			bands_kept = {}
			for region_text in extracted_regions:
				for band in zones["zones"][region_text].get( "bands", [] ):
					bands_kept[int( band )] = bands_kept.get( int( band ), 0 ) + 1
			declared_bands = {int( band ) for band in zones.get( "anchorsByBand", {} )}
			empty_bands = sorted( band for band in declared_bands if not bands_kept.get( band ) )
			if empty_bands:
				print( f"skipped regions would empty trajectory bands {empty_bands}", file = sys.stderr )
				return 3

			# -- the shared map plane --------------------------------------
			for rel in ( "object.ifo", "tile2d.ifo" ):
				payload = map_payload( rel )
				if payload is None:
					print( f"Map.pk2 is missing the shared {rel}", file = sys.stderr )
					return 3
				store( "mapSector", map_root_out, rel, payload )
			tile_entries = sorted( ( folded, entry ) for folded, entry in map_entries.items() if folded.startswith( "tile2d/" ) )
			for folded, entry in tile_entries:
				store( "tile2d", map_root_out, folded, sro_pk2.payload( map_memory, entry ) )
			for rel in ( "navmesh/mapinfo.mfo", "navmesh/objectstring.ifo" ):
				payload = data_payload( rel )
				if payload is None:
					print( f"Data.pk2 is missing the shared {rel}", file = sys.stderr )
					return 3
				store( "navmesh", data_root_out, rel, payload )

			# -- the outdoor minimap tiles ----------------------------------
			# Media.pk2's minimap/<x>x<z>.ddj per region; a region the client
			# ships no tile for simply has no minimap art (the client skips
			# absent tiles), so absence is recorded, never fatal.
			media_root_out = output / "Media_extracted"
			minimap_tiles = 0
			for region_text in extracted_regions:
				region_id = int( region_text, 16 )
				tile = f"minimap/{region_id & 0xff}x{(region_id >> 8) & 0xff}.ddj"
				entry = media_entries.get( tile )
				if entry is None:
					continue
				store( "minimap", media_root_out, tile, sro_pk2.payload( media_memory, entry ) )
				minimap_tiles += 1

			# -- the object-resource closure --------------------------------
			object_info = parse_object_info(
				sro_pk2.payload( map_memory, map_entries["object.ifo"] ).decode( "cp949", errors = "replace" ).lstrip( "\ufeff" )
			)
			unknown_ids = sorted( object_id for object_id in object_ids if object_id not in object_info )
			if unknown_ids:
				print(
					f"{len( unknown_ids )} placed object ids are absent from object.ifo: {unknown_ids[:8]}",
					file = sys.stderr,
				)
				return 3

			pending = sorted( {object_info[object_id] for object_id in object_ids} )
			visited = set()
			closure = {"compound": 0, "bsr": 0, "bmt": 0, "bms": 0, "ddj": 0}
			missing = []
			while pending:
				folded = pending.pop()
				if folded in visited:
					continue
				visited.add( folded )
				payload = data_payload( folded )
				if payload is None:
					missing.append( folded )
					continue
				if folded.endswith( ".ddj" ):
					category = "ddj"
				elif folded.endswith( ".bms" ):
					category = "bms"
				else:
					category = None
				if category is None:
					signature = payload[:SIGNATURE_BYTES]
					if signature == CPD_SIGNATURE:
						category = "compound"
						pending.extend( parse_compound_branches( payload, folded ) )
					elif signature == BSR_SIGNATURE:
						category = "bsr"
						material_paths, mesh_paths = parse_bsr_resource_paths( payload, folded )
						pending.extend( material_paths )
						pending.extend( mesh_paths )
					elif signature == BMT_SIGNATURE:
						category = "bmt"
						pending.extend( parse_bmt_texture_paths( payload, folded ) )
					else:
						print( f"{folded}: unknown resource signature {signature!r}", file = sys.stderr )
						return 3
				closure[category] += 1
				store( "objectResource", data_root_out, folded, payload )
			if missing:
				print( f"{len( missing )} closure resources are absent from Data.pk2: {missing[:8]}", file = sys.stderr )
				return 3
		finally:
			map_memory.close()
			data_memory.close()
			media_memory.close()

	inventory = {
		"sourceClient": SOURCE_CLIENT,
		"gameRoot": str( game_root ),
		"zonesPath": str( zones_path ),
		"archives": archives,
		"regions": {
			"declared": len( zones["zones"] ),
			"extracted": len( extracted_regions ),
			"skippedNoMapPlane": skipped_regions,
		},
		"bands": bands_kept,
		"placements": placement_count,
		"objectIds": len( object_ids ),
		"objectClosure": closure,
		"minimapTiles": minimap_tiles,
		"files": files,
		"bytes": bytes_out,
	}
	inventory_path.write_text( json.dumps( inventory, indent = 2, sort_keys = True ) + "\n", encoding = "utf-8" )
	print(
		f"world plane: {len( extracted_regions )}/{len( zones['zones'] )} regions, "
		f"{placement_count} placements, {len( object_ids )} object ids, closure {closure}"
	)
	print( f"files {files}, bytes {bytes_out}" )
	print( f"inventory: {inventory_path}" )
	return 0


if __name__ == "__main__":
	raise SystemExit( main() )
