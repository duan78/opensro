from pathlib import Path

p = Path("internal/game/action/hwan_test.go")
text = p.read_text(encoding="utf-8")
old = '\tcharacter.Skills = []uint32{100}'
new = ('\tcharacter.Skills = []uint32{100}\n'
       '\tskills := staticSkillSource{100: skill100()}\n'
       '\trt.deps = &enterworld.Deps{Characters: enterworld.StaticCharacterSource{testDivision: {character, inMode, idle}}, Items: testItems(), Skills: skills}')
assert old in text
text = text.replace(old, new, 1)
# remove the earlier deps assignment (before Skills was set)
old2 = '\trt.deps = &enterworld.Deps{Characters: enterworld.StaticCharacterSource{testDivision: {character, inMode, idle}},\n\t\tItems: testItems(), Skills: staticSkillSource{}}\n'
assert old2 in text
text = text.replace(old2, "", 1)
# helper
old3 = 'func TestHwanAuraExtendsRunningBerserkOnly(t *testing.T) {'
new3 = ('func skill100() enterworld.SkillRow {\n'
        '\treturn enterworld.SkillRow{ID: 100, Group: 9, EffectDurationMs: 300000, RequiredWeaponKinds: [2]uint8{0xff, 0xff},\n'
        '\t\tConsumption: enterworld.SkillConsumption{Pinned: true}, TimingPinned: true,\n'
        '\t\tTimedEffect: enterworld.SkillTimedEffect{Pinned: true, HwanDurationMs: 30000,\n'
        '\t\t\tArea: enterworld.SkillRecipientArea{Present: true, Radius: 300, MaxTargets: 8, Select: 5}}}\n'
        '}\n\n'
        'func TestHwanAuraExtendsRunningBerserkOnly(t *testing.T) {')
assert old3 in text
text = text.replace(old3, new3, 1)
# the initial fixture row can be empty now
old4 = 'rt, character := newActiveEffectTestRuntime(t, staticSkillSource{\n\t\t100: {ID: 100, Group: 9, EffectDurationMs: 300000, RequiredWeaponKinds: [2]uint8{0xff, 0xff},\n\t\t\tConsumption: enterworld.SkillConsumption{Pinned: true}, TimingPinned: true,\n\t\t\tTimedEffect: enterworld.SkillTimedEffect{Pinned: true, HwanDurationMs: 30000,\n\t\t\t\tArea: enterworld.SkillRecipientArea{Present: true, Radius: 300, MaxTargets: 8, Select: 5}}},\n\t)'
assert old4 in text
text = text.replace(old4, 'rt, character := newActiveEffectTestRuntime(t, staticSkillSource{})', 1)
p.write_text(text, encoding="utf-8", newline="\n")
print("fixture rebuilt")
