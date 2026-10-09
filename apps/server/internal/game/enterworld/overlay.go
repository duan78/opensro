/*
===========================================================================

overlay.go - the extended catalog overlay (port-only, not native)

Extended content (isro-live-2026) grafts the live tables onto the native
v1.150 ones: the native row wins every conflict (charter 4.1), the live
row adds what the native game never had - degrees 11-14, the 91-140
mobs, the modern skill rows. Both sides load through the ordinary
textdata readers over the sealed projection's textdata tree, so no
parser knows which side a row came from; only this file arbitrates.

===========================================================================
*/
package enterworld

import "sort"

/*
================
OverlayItems

The item and character catalog graft: every lookup asks the native table
first, so a row both files carry answers natively and a codename the
native game never shipped resolves through the extended table.
================
*/
type OverlayItems struct {
	native   *TextdataItems
	extended *TextdataItems
}

/*
================
NewOverlayItems
================
*/
func NewOverlayItems(native, extended *TextdataItems) *OverlayItems {
	return &OverlayItems{native: native, extended: extended}
}

// ItemRefByCodename implements ItemRefSource (native first).
func (o *OverlayItems) ItemRefByCodename(codename string) (*ItemRef, bool) {
	if ref, ok := o.native.ItemRefByCodename(codename); ok {
		return ref, ok
	}
	return o.extended.ItemRefByCodename(codename)
}

// ItemRefByID implements ItemRefSource (native first).
func (o *OverlayItems) ItemRefByID(id uint32) (*ItemRef, bool) {
	if ref, ok := o.native.ItemRefByID(id); ok {
		return ref, ok
	}
	return o.extended.ItemRefByID(id)
}

// CharacterRefByCodename implements CharacterRefSource (native first).
func (o *OverlayItems) CharacterRefByCodename(codename string) (*CharacterRef, bool) {
	if ref, ok := o.native.CharacterRefByCodename(codename); ok {
		return ref, ok
	}
	return o.extended.CharacterRefByCodename(codename)
}

/*
================
SummonableCharacterRefs

The union of both families; a codename both tables carry publishes once,
natively.
================
*/
func (o *OverlayItems) SummonableCharacterRefs() []CharacterRef {
	native := o.native.SummonableCharacterRefs()
	seen := make(map[string]struct{}, len(native))
	out := make([]CharacterRef, 0, len(native)+16)
	for index := range native {
		seen[native[index].Codename] = struct{}{}
		out = append(out, native[index])
	}
	for _, ref := range o.extended.SummonableCharacterRefs() {
		if _, dup := seen[ref.Codename]; dup {
			continue
		}
		seen[ref.Codename] = struct{}{}
		out = append(out, ref)
	}
	return out
}

/*
================
ItemCommandReferences

Both tables' composer rows merged by reference id; a shared id answers
natively, never twice.
================
*/
func (o *OverlayItems) ItemCommandReferences() []ItemCommandReference {
	byID := make(map[uint32]ItemCommandReference, o.native.Len()+o.extended.Len())
	for _, row := range o.native.ItemCommandReferences() {
		byID[row.RefObjID] = row
	}
	for _, row := range o.extended.ItemCommandReferences() {
		if _, dup := byID[row.RefObjID]; dup {
			continue
		}
		byID[row.RefObjID] = row
	}
	rows := make([]ItemCommandReference, 0, len(byID))
	for _, row := range byID {
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].RefObjID < rows[j].RefObjID })
	return rows
}

/*
================
Len

The readiness face: the sum of both tables' rows. The merged distinct
counts live in the projection's census, not here.
================
*/
func (o *OverlayItems) Len() int {
	return o.native.Len() + o.extended.Len()
}

/*
================
UseBoundedCache

Swap both sides onto their bounded record caches. The composition calls
this once after startup compilation, exactly as it does for the plain
native table.
================
*/
func (o *OverlayItems) UseBoundedCache(capacity int) error {
	if err := o.native.UseBoundedCache(capacity); err != nil {
		return err
	}
	return o.extended.UseBoundedCache(capacity)
}

/*
================
OverlaySkills

The skill catalog graft, same arbitration: a skill id both tables carry
answers natively, a modern row the native game never had resolves
through the extended table.
================
*/
type OverlaySkills struct {
	native   *TextdataSkills
	extended *TextdataSkills
}

/*
================
NewOverlaySkills
================
*/
func NewOverlaySkills(native, extended *TextdataSkills) *OverlaySkills {
	return &OverlaySkills{native: native, extended: extended}
}

// SkillByID implements SkillDataSource (native first).
func (o *OverlaySkills) SkillByID(id uint32) (SkillRow, bool) {
	if row, ok := o.native.SkillByID(id); ok {
		return row, ok
	}
	return o.extended.SkillByID(id)
}

// SkillByCodename resolves by codename (native first).
func (o *OverlaySkills) SkillByCodename(codename string) (SkillRow, bool) {
	if row, ok := o.native.SkillByCodename(codename); ok {
		return row, ok
	}
	return o.extended.SkillByCodename(codename)
}

// Len is the readiness face: the sum of both tables' rows.
func (o *OverlaySkills) Len() int {
	return o.native.Len() + o.extended.Len()
}

/*
================
Load

Eagerly load the extended side so the composition fails the boot on a
table that does not parse; the native side loads through its own
readiness path.
================
*/
func (o *OverlaySkills) Load() error {
	return o.extended.Load()
}

/*
================
UseBoundedCache

Both sides onto their bounded record caches, once after startup.
================
*/
func (o *OverlaySkills) UseBoundedCache(capacity int) error {
	if err := o.native.UseBoundedCache(capacity); err != nil {
		return err
	}
	return o.extended.UseBoundedCache(capacity)
}
