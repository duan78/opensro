"""
===========================================================================

extract_extended_client_data.py - pull the live-2026 textdata (extended)

Extended content (isro-live-2026), port-only, not v1.150-native. The
retail install is read strictly read-only (mmap; nothing is ever written
under the game root and no executable is touched): the progression tables
and the item/character fragment shards the extended bundle build censuses
come out as lossless raw bytes, with an inventory that seals them to the
five PK2 archives by SHA-256. Archive digests are incremental - a rebuild
whose archive size and mtime did not move reuses the recorded digest
instead of re-hashing 5.7 GB.

Output lives under the generated root (scripts/sro_paths.py):
<generated>/extended/source, never in a source tree.

===========================================================================
"""
import argparse
import hashlib
import json
import mmap
import os
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import sro_paths
import sro_pk2

ARCHIVES = ("Data.pk2", "Media.pk2", "Map.pk2", "Particles.pk2", "Music.pk2")
SOURCE_CLIENT = "isro-live-2026"
TEXTDATA_PREFIX = "server_dep/silkroad/textdata/"
# The progression tables the extended level curve reads (mission M1/M2):
# leveldata for XP/SP/mob-basis/jobs, levelgold as the gold record,
# dg.txt as the gold-walk basis the native server reads (CDropGoldData).
# textdataname/magicoption are single tables the reference loaders read.
TABLES = (
	"leveldata.txt",
	"levelgold.txt",
	"dg.txt",
	"textdataname.txt",
	"magicoption.txt",
	"npcpos.txt",
	# The teleport plane the E7 graft reads: buildings (gate NPCs and their
	# bounds), destinations, links, and the fortress gate ownership column.
	"teleportbuilding.txt",
	"teleportdata.txt",
	"teleportlink.txt",
	"siegefortress.txt",
)
# Loader files whose listed shards are extracted for the item/character
# and skill censuses (degrees, requirement levels, mob levels, skills).
LOADERS = {
	"itemdata": "itemdata.txt",
	"characterdata": "characterdata.txt",
	"skilldata": "skilldata.txt",
}
DEFAULT_GAME_ROOT = r"C:\Program Files (x86)\Silkroad"
HASH_CHUNK = 4 * 1024 * 1024

# The live-2026 characterdata rows carry the v1.150 _RefObjChar layout
# minus six columns the modern client dropped (measured 2026-10-09 on
# MOB_CH_MANGNYANG: columns 77.. of the legacy row appear at 71.. in the
# live row, identical values; the six legacy cells 71..76 are gone and
# the legacy tail 110..119 no longer ships). The server's readers (and
# every provenance comment pinning them) stay on the legacy indexes, so
# the extraction reshapes each live row back: six empty cells reinserted
# at index 71, the row padded to CHARACTER_LEGACY_COLUMNS.
CHARACTER_REMOVED_AT = 71
CHARACTER_REMOVED_COUNT = 6
CHARACTER_LEGACY_COLUMNS = 120


# ================
# reshape_character_cells
#
# One live-2026 characterdata row back to the v1.150 column layout.
# ================
def reshape_character_cells( cells ):
	if len( cells ) < CHARACTER_REMOVED_AT:
		return cells
	# The reinserted and padded cells carry "0", not "": the server's
	# cell parsers refuse an empty cell where the legacy row carried a
	# number (an empty CanRide would drop the whole character row).
	reshaped = cells[:CHARACTER_REMOVED_AT] + ["0"] * CHARACTER_REMOVED_COUNT + cells[CHARACTER_REMOVED_AT:]
	while len( reshaped ) < CHARACTER_LEGACY_COLUMNS:
		reshaped.append( "0" )
	return reshaped


# ================
# sha256_file
#
# Chunked so the 3.2 GB Data.pk2 never lands in memory at once.
# ================
def sha256_file( path: Path ) -> str:
	digest = hashlib.sha256()
	with open( path, "rb" ) as handle:
		for chunk in iter( lambda: handle.read( HASH_CHUNK ), b"" ):
			digest.update( chunk )
	return digest.hexdigest()


# ================
# archive_digest
#
# Reuse the previous run's digest when size and mtime both match: the
# archive identity the manifest seals is content, and identical
# (size, mtime) on an immutable retail install is that content.
# ================
def archive_digest( path: Path, previous: dict ) -> str:
	stat = path.stat()
	if previous.get( path.name ) and previous[path.name]["bytes"] == stat.st_size and previous[path.name]["mtimeNs"] == stat.st_mtime_ns:
		return previous[path.name]["sha256"]
	return sha256_file( path )


def main() -> int:
	parser = argparse.ArgumentParser( description = "Extract the extended (live-2026) client textdata." )
	parser.add_argument( "--game-root", default = os.environ.get( "SRO_EXTENDED_GAME_ROOT", DEFAULT_GAME_ROOT ) )
	parser.add_argument( "--output", default = None, help = "Defaults to <generated>/extended/source." )
	args = parser.parse_args()

	game_root = Path( args.game_root )
	output = Path( args.output ) if args.output else sro_paths.GENERATED_ROOT / "extended" / "source"
	for name in ARCHIVES:
		if not ( game_root / name ).is_file():
			print( f"missing archive {game_root / name}", file = sys.stderr )
			print( f"point --game-root or SRO_EXTENDED_GAME_ROOT at the live-2026 install", file = sys.stderr )
			return 2

	previous = {}
	inventory_path = output / "inventory.json"
	if inventory_path.is_file():
		try:
			previous = json.loads( inventory_path.read_text( encoding = "utf-8" ) ).get( "archives", {} )
		except ( OSError, ValueError ):
			previous = {}

	archives = {}
	for name in ARCHIVES:
		path = game_root / name
		stat = path.stat()
		sha256 = archive_digest( path, previous )
		archives[name] = {"bytes": stat.st_size, "mtimeNs": stat.st_mtime_ns, "sha256": sha256}
		print( f"archive {name}: {stat.st_size} bytes, sha256 {sha256[:16]}..." )

	output.mkdir( parents = True, exist_ok = True )
	files = {}
	with open( game_root / "Media.pk2", "rb" ) as handle:
		memory = mmap.mmap( handle.fileno(), 0, access = mmap.ACCESS_READ )
		try:
			by_path = {sro_pk2.fold_ascii( entry.path ): entry for entry in sro_pk2.read_directory( memory )}

			def extract( name: str, transform = None ) -> None:
				entry = by_path.get( sro_pk2.fold_ascii( TEXTDATA_PREFIX + name ) )
				if entry is None:
					raise RuntimeError( f"{name} is absent from the live-2026 Media.pk2 textdata" )
				payload = sro_pk2.payload( memory, entry )
				if transform is not None:
					payload = transform( payload )
				# The loaders' shard lists carry mixed case; the server's
				# glob readers (monster.LoadMonsterRefs) and the native
				# projection's file names are lowercase, so every shard is
				# written under its folded name.
				name = sro_pk2.fold_ascii( name )
				( output / name ).write_bytes( payload )
				files[name] = {
					"bytes": len( payload ),
					"sha256": hashlib.sha256( payload ).hexdigest(),
				}

			def reshape_character_shard( payload: bytes ) -> bytes:
				text = payload.decode( "utf-16-le", errors = "replace" ).lstrip( "\ufeff" )
				lines = []
				for line in text.splitlines():
					if not line.strip() or line.startswith( "//" ):
						continue
					lines.append( "\t".join( reshape_character_cells( line.split( "\t" ) ) ) )
				return ( "\ufeff" + "\r\n".join( lines ) + "\r\n" ).encode( "utf-16-le" )

			for table in TABLES:
				if table == "textdataname.txt":
					continue  # synthesized below from the object-name shards
				extract( table )
			shards = {}
			for label, loader in LOADERS.items():
				extract( loader )
				text = ( output / loader ).read_bytes().decode( "utf-16-le" ).lstrip( "\ufeff" )
				names = [line.strip() for line in text.splitlines() if line.strip()]
				for shard in names:
					extract( shard, reshape_character_shard if label == "characterdata" else None )
				shards[label] = len( names )
			# The skill loader feeds skilldata_virtual.txt through the same
			# parser; ship it when the client carries one, absence is legal.
			virtual = by_path.get( sro_pk2.fold_ascii( TEXTDATA_PREFIX + "skilldata_virtual.txt" ) )
			if virtual is not None:
				extract( "skilldata_virtual.txt" )
			# The live client's textdataname.txt is an empty BOM: object
			# names live in the textdata_object shards and equipment/skill
			# names in the textdata_equip&skill shards (the same measured
			# layout: SN_ symbol column 2, English column 9, 2026-10-09).
			# Project them onto the legacy layout the server's name reader
			# expects (symbol column 1, English column 8).
			name_rows = []
			name_shard_count = 0
			for family in ( "textdata_object", "textdata_equip&skill" ):
				loader_entry = by_path.get( sro_pk2.fold_ascii( TEXTDATA_PREFIX + family + ".txt" ) )
				if loader_entry is None:
					continue
				loader_text = sro_pk2.payload( memory, loader_entry ).decode( "utf-16-le", errors = "replace" ).lstrip( "\ufeff" )
				family_shards = [line.strip() for line in loader_text.splitlines() if line.strip()]
				name_shard_count += len( family_shards )
				for shard in family_shards:
					entry = by_path.get( sro_pk2.fold_ascii( TEXTDATA_PREFIX + shard ) )
					if entry is None:
						continue
					shard_text = sro_pk2.payload( memory, entry ).decode( "utf-16-le", errors = "replace" ).lstrip( "\ufeff" )
					for line in shard_text.splitlines():
						cells = line.split( "\t" )
						if len( cells ) < 10 or not cells[2].strip().startswith( "SN_" ):
							continue
						symbol = cells[2].strip()
						english = cells[9].strip()
						if not english or english in ( "0", "xxx" ):
							continue
						legacy = ["1", symbol, "0", "0", "0", "0", "0", "0", english]
						name_rows.append( "\t".join( legacy ) )
			shards["textdata_names"] = name_shard_count
			name_bytes = ( "\ufeff" + "\r\n".join( name_rows ) + "\r\n" ).encode( "utf-16-le" )
			( output / "textdataname.txt" ).write_bytes( name_bytes )
			files["textdataname.txt"] = {
				"bytes": len( name_bytes ),
				"sha256": hashlib.sha256( name_bytes ).hexdigest(),
			}
		finally:
			memory.close()

	inventory = {
		"sourceClient": SOURCE_CLIENT,
		"gameRoot": str( game_root ),
		"archives": archives,
		"files": files,
		"shardCounts": shards,
	}
	inventory_path.write_text( json.dumps( inventory, indent = 2, sort_keys = True ) + "\n", encoding = "utf-8" )
	print( f"extracted {len( files )} files, {shards['itemdata']} itemdata shards, {shards['characterdata']} characterdata shards" )
	print( f"inventory: {inventory_path}" )
	return 0


if __name__ == "__main__":
	raise SystemExit( main() )
