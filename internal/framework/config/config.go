// Package config holds project-wide settings that are stable across the
// lifetime of an install: the charter directory name, framework version
// pin, and (in later phases) the policy tree declared in keystone.json.
package config

// DefaultAgentsRoot is the directory that holds the charter's primitives:
// rules, hooks, permissions, skills, commands, subagents, and the rest. At
// 5.0 this is `.agents` — the tool-neutral directory hosts already read
// (`.agents/skills/` is native in Codex, Cursor, Copilot, opencode, Goose,
// and Pi). Only standard primitives live here (ADR 0010).
//
// This is the one place to change the primitives location. `keystone.json`
// stays at the repo root as the entry-point config.
const DefaultAgentsRoot = ".agents"

// StateRoot holds everything keystone-private: INDEX.json, the lockfile,
// learning/, work/, archive/, state/, sources/, hosts/, and vendored
// policies/. Kept out of .agents/ so hosts that read .agents/ see only
// standard files.
const StateRoot = ".keystone"

// PoliciesDir is where `keystone policy add` vendors policy trees,
// repo-relative.
const PoliciesDir = StateRoot + "/policies"

// LockfileName is the basename of the per-install state record, written
// at <StateRoot>/lockfile.json.
const LockfileName = "lockfile.json"

// IndexName is the basename of the generated primitive descriptor index,
// written at <StateRoot>/INDEX.json.
const IndexName = "INDEX.json"

// IndexLiteName is the basename of the cheap-discovery sibling index
// (kind/id/description only). Written alongside INDEX.json by every
// `keystone index` run. Agents reference this for first-pass browsing;
// INDEX.json is opened only when a path, glob, or trigger is needed to
// activate a specific primitive.
const IndexLiteName = "INDEX.lite.json"
