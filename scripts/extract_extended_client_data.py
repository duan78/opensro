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
# The progression tables the extended level curve reads (mission M1/M2).
TABLES = ("leveldata.txt", "levelgold.txt")
# Loader files whose listed shards are extracted for the item/character
# censuses (degrees, requirement levels, mob levels).
LOADERS = {
	"itemdata": "itemdata.txt",
	"characterdata": "characterdata.txt",
}
DEFAULT_GAME_ROOT = r"C:\Program Files (x86)\Silkroad"
HASH_CHUNK = 4 * 1024 * 1024


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

			def extract( name: str ) -> None:
				entry = by_path.get( sro_pk2.fold_ascii( TEXTDATA_PREFIX + name ) )
				if entry is None:
					raise RuntimeError( f"{name} is absent from the live-2026 Media.pk2 textdata" )
				payload = sro_pk2.payload( memory, entry )
				( output / name ).write_bytes( payload )
				files[name] = {
					"bytes": len( payload ),
					"sha256": hashlib.sha256( payload ).hexdigest(),
				}

			for table in TABLES:
				extract( table )
			shards = {}
			for label, loader in LOADERS.items():
				extract( loader )
				text = ( output / loader ).read_bytes().decode( "utf-16-le" ).lstrip( "\ufeff" )
				names = [line.strip() for line in text.splitlines() if line.strip()]
				for shard in names:
					extract( shard )
				shards[label] = len( names )
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
