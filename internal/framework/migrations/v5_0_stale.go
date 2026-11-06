package migrations

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
)

// staleScaffold_5_0 lists every shipped version (sha256 of SKILL.md) of the
// 4.x scaffold skills 5.0 replaced with one keystone-new skill. A skill
// folder whose only file matches a shipped version was never edited, so it
// is keystone's to move; any other folder is the user's and stays.
var staleScaffold_5_0 = map[string]string{
	"44dd19662926dff1ca40f0e04a9e3514b6e83f24bbd1581d4bb6c9e75d7dba5d": "keystone-new-guide",
	"637c2f9a65a502b277cfc61ee3e7cec98354b9f2359dbba60995aecbd676c3cc": "keystone-new-guide",
	"63868abc4de2d4016412670c661820db2812ca1e8114ef204e47348ce5eee3fb": "keystone-new-guide",
	"e05979e6cbc0def705ddedcad726e8fb5d3059f283355fac15c775cfb1cc1530": "keystone-new-guide",
	"e42649073de8b168e1ccc7e5b4dbd59c02a98d48b7fe4d01eee105d6b9cb4356": "keystone-new-guide",
	"0e134dcc1fbf57c1e99ba439d8df77fbd5220c389b4108886d1769f150d36435": "keystone-new-corpus",
	"bb5ef875537f2c17163941d9d54ef0412a554e0f066903049f38754c4f41f8d0": "keystone-new-corpus",
	"d7bf951b1c2e5d4c40e13a2539b63e7b29be1986566414abd9ef4786da13c533": "keystone-new-corpus",
	"f322bef74f0b234fab6ea6072b2114206598eec89f2ac5ebbfc7898b57d10c1b": "keystone-new-corpus",
	"1ec9848953c1a81210862f4968a202e641e9e628545fbf98f34b34031fa70197": "keystone-new-sensor",
	"976c311724542e3ca6938e98681a9861a449b9670f17b2b3920c5ac3bc85966e": "keystone-new-sensor",
	"c432d35fd5ccc601bd0d914cfa700070b620d7cf13dfea982b1d31dfcb8e8b6a": "keystone-new-sensor",
	"f8cd7b421844ca060cb7a035503330648a59b6c2a4ea0bcff256b4728407cf76": "keystone-new-sensor",
	"77da90e992cd520659a0851194bda4a134082ed2e41d66f56f1298d5225674bb": "keystone-new-playbook",
	"b0df87c7e67ec8f32b7d6929e9b8c6c468e00f3c8dc0db13053d515d92efaa1d": "keystone-new-playbook",
	"e28a23f0dccac3bbe7da7827b53317432854c1f75ef13515eaae123d578f64e9": "keystone-new-playbook",
	"e38b6897a1bc02b263ef2adcd43ada38b22c50d1b639b595f36fcc60c94bfb18": "keystone-new-playbook",
	"9e81a82d518825e1f7874299fe89bb61c0e492dade908616a284fb41e5946f63": "keystone-new-command",
	"b23af52df9154abce710d6094548705bb3235cdf4cefb20e2626104af160796f": "keystone-new-command",
	"c8c668add544f168f433180d3d12cfdf6d914fac32730ce57434ddf8276d52ea": "keystone-new-command",
	"ec2ef66091b88ebc82b0c48c0a0613c21905611594b2e5672f06aa3aeb1a2826": "keystone-new-command",
	"07929c26002b95ec85ab28e9ca5590f7e4b45a63cca04de3db11ed6ec1f9a8d1": "keystone-new-action",
}

// staleArchive_5_0 is where Up moves an unedited stale scaffold skill (a
// rename, iron law 3) and where Down takes it back from.
const staleArchive_5_0 = stateRoot_5_0 + "/archive/5.0/skills"

// archiveStaleScaffold_5_0 moves each unedited 4.x scaffold skill out of
// .agents/skills/. An edited one stays (iron law 2).
func archiveStaleScaffold_5_0(absDir string) error {
	skills := filepath.Join(absDir, agentsRoot_5_0, "skills")
	for _, name := range staleNames_5_0() {
		dir := filepath.Join(skills, name)
		if !unedited_5_0(dir, name) {
			continue
		}
		if err := mergeMove_5_0(dir, filepath.Join(absDir, staleArchive_5_0, name)); err != nil {
			return err
		}
	}
	return nil
}

// restoreStaleScaffold_5_0 is Down: archived skills return to .agents/skills/.
func restoreStaleScaffold_5_0(absDir string) error {
	archive := filepath.Join(absDir, staleArchive_5_0)
	for _, name := range staleNames_5_0() {
		if err := mergeMove_5_0(filepath.Join(archive, name), filepath.Join(absDir, agentsRoot_5_0, "skills", name)); err != nil {
			return err
		}
	}
	for _, dir := range []string{archive, filepath.Dir(archive)} {
		_ = removeEmptyDir_5_0(dir)
	}
	return nil
}

// unedited_5_0: dir holds only SKILL.md, matching a shipped version of name.
func unedited_5_0(dir, name string) bool {
	if !onlySkillFile_5_0(dir) {
		return false
	}
	data, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if err != nil {
		return false
	}
	sum := sha256.Sum256(data)
	return staleScaffold_5_0[hex.EncodeToString(sum[:])] == name
}

// onlySkillFile_5_0: dir holds SKILL.md and nothing else.
func onlySkillFile_5_0(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		return false
	}
	return entries[0].Name() == "SKILL.md"
}

func staleNames_5_0() []string {
	seen := map[string]bool{}
	var names []string
	for _, n := range staleScaffold_5_0 {
		if !seen[n] {
			seen[n] = true
			names = append(names, n)
		}
	}
	return names
}
