/*
===========================================================================

graftrefs_test.go - the extended reference graft keeps the native row

Extended content (isro-live-2026, port-only, not native): the native
reference wins every shared id, the live reference the native game never
shipped joins the template, and the nests are rebuilt unchanged.

===========================================================================
*/
package monster

import "testing"

func TestGraftRefsNativeRowWinsAndLiveRowJoins(t *testing.T) {
	base := TemplateFromParts(map[uint32]MonsterRef{
		1933: {RefObjID: 1933, Codename: "MOB_CH_MANGNYANG", WalkSpeed: 8},
	}, []NestRow{{
		SpawnPoint:   SpawnPoint{RefObjID: 1933, RegionID: 0x7e7e, X: 1020, Z: 980},
		PolicyPinned: true,
		MaxCount:     1,
	}})
	grafted := GraftRefs(base, map[uint32]MonsterRef{
		1933: {RefObjID: 1933, Codename: "LIVE_RETUNE", WalkSpeed: 99},
		4242: {RefObjID: 4242, Codename: "MOB_TQ_SNAKEWOMAN", WalkSpeed: 40},
	})
	if got := grafted.Refs[1933].Codename; got != "MOB_CH_MANGNYANG" {
		t.Fatalf("the native row wins the shared id, got %q", got)
	}
	if got, ok := grafted.Refs[4242]; !ok || got.Codename != "MOB_TQ_SNAKEWOMAN" {
		t.Fatalf("the live row joins the template, got %+v/%v", got, ok)
	}
	if len(grafted.Nests) != 1 || grafted.Nests[0].RefObjID != 1933 {
		t.Fatalf("the native nests are rebuilt unchanged, got %+v", grafted.Nests)
	}
}

func TestGraftRefsWithoutNewRowsReturnsTheBaseTemplate(t *testing.T) {
	base := TemplateFromParts(map[uint32]MonsterRef{1933: {RefObjID: 1933}}, nil)
	if got := GraftRefs(base, map[uint32]MonsterRef{1933: {RefObjID: 1933}}); got.EvidenceMatches != base.EvidenceMatches {
		t.Fatal("a graft that adds nothing must return the base template")
	}
}
