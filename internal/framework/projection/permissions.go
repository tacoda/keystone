package projection

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// Rules is the host-neutral permission set, in Claude Code rule syntax
// (`Tool` or `Tool(specifier)`): the union of every permission primitive.
type Rules struct {
	Allow, Ask, Deny []string
}

func (r Rules) empty() bool { return len(r.Allow)+len(r.Ask)+len(r.Deny) == 0 }

// buckets lists the rule lists by name, in evaluation-relevant order.
func (r Rules) buckets() []bucket {
	return []bucket{{"allow", r.Allow}, {"ask", r.Ask}, {"deny", r.Deny}}
}

type bucket struct {
	name  string
	rules []string
}

// CollectRules unions every permission primitive's lists, deduped in order.
func CollectRules(primitives []primitive.Primitive) Rules {
	var r Rules
	seen := map[string]bool{}
	for _, p := range primitives {
		if primitive.Kind(p.Kind) != primitive.KindPermission {
			continue
		}
		r.Allow = appendUnique(r.Allow, p.Allow, "allow", seen)
		r.Ask = appendUnique(r.Ask, p.Ask, "ask", seen)
		r.Deny = appendUnique(r.Deny, p.Deny, "deny", seen)
	}
	return r
}

func appendUnique(list, add []string, bucket string, seen map[string]bool) []string {
	for _, x := range add {
		if k := bucket + ":" + x; !seen[k] {
			seen[k] = true
			list = append(list, x)
		}
	}
	return list
}

// PermissionBridge translates the canonical rules for one host. It returns
// the output (nil Dest when the host has no file to write) and every rule
// the host cannot enforce exactly, as "bucket: rule". Translation may make a
// deny or ask stricter, never an allow broader.
type PermissionBridge func(r Rules, managed []string) (Output, []string)

// permissionOutputs plans each host's permission output, plus a Report for
// every host that cannot enforce some rule (never a silent drop).
func permissionOutputs(primitives []primitive.Primitive, hosts []Profile, managed map[string][]string) []Output {
	r := CollectRules(primitives)
	var outs []Output
	for _, h := range hosts {
		outs = append(outs, hostPermissionOutputs(h, r, managed)...)
	}
	return append(outs, disabledPermissionCleanup(hosts, managed)...)
}

// disabledPermissionCleanup strips keystone's recorded entries from the
// shared permission file of a host no longer in adapters; nothing else would
// remove them. (An owned file, like Codex's, is pruned instead.)
func disabledPermissionCleanup(hosts []Profile, managed map[string][]string) []Output {
	enabled := map[string]bool{}
	for _, h := range hosts {
		enabled[h.Host] = true
	}
	var outs []Output
	for _, mk := range profiles {
		h := mk()
		if enabled[h.Host] || h.Permissions == nil {
			continue
		}
		o, _ := h.Permissions(Rules{}, managed[permissionDest(h)])
		if o.Strategy == Merge && len(managed[o.Dest]) > 0 {
			o.Host, o.Kind = h.Host, string(primitive.KindPermission)
			outs = append(outs, o)
		}
	}
	return outs
}

// hostPermissionOutputs is one host's permission file (if any) plus a
// Report naming every rule it cannot enforce.
func hostPermissionOutputs(h Profile, r Rules, managed map[string][]string) []Output {
	if h.Permissions == nil {
		return reportRules(h, allRules(r))
	}
	o, unsupported := h.Permissions(r, managed[permissionDest(h)])
	var outs []Output
	if writesPermissionFile(o, r) {
		o.Host, o.Kind = h.Host, string(primitive.KindPermission)
		outs = append(outs, o)
	}
	return append(outs, reportRules(h, unsupported)...)
}

// writesPermissionFile: with no rules, only merges run (to strip stale
// entries); an owned file is pruned instead of being written empty.
func writesPermissionFile(o Output, r Rules) bool {
	if o.Dest == "" {
		return false
	}
	return !r.empty() || o.Strategy == Merge
}

func reportRules(h Profile, unsupported []string) []Output {
	if len(unsupported) == 0 {
		return nil
	}
	return []Output{{Host: h.Host, Kind: string(primitive.KindPermission), Strategy: Report,
		Note: "cannot enforce: " + strings.Join(unsupported, "; ")}}
}

// permissionDest is where a host's merged permissions live (for the
// managed-entries record).
func permissionDest(h Profile) string {
	o, _ := h.Permissions(Rules{}, nil)
	return o.Dest
}

func allRules(r Rules) []string {
	var out []string
	for _, b := range r.buckets() {
		for _, x := range b.rules {
			out = append(out, b.name+": "+x)
		}
	}
	return out
}

// parseRule splits `Tool(spec)` into its parts; a bare `Tool` has no spec.
func parseRule(rule string) (tool, spec string) {
	open := strings.IndexByte(rule, '(')
	if open < 0 || !strings.HasSuffix(rule, ")") {
		return rule, ""
	}
	return rule[:open], rule[open+1 : len(rule)-1]
}

// bashPrefix reads a Bash specifier: `cmd args:*` is a prefix match on
// "cmd args"; anything else is an exact command. A `*` anywhere else is a
// wildcard no host here can express: no words, so it is reported.
func bashPrefix(spec string) (words []string, prefix bool) {
	p, prefix := strings.CutSuffix(spec, ":*")
	if strings.Contains(p, "*") {
		return nil, false
	}
	return strings.Fields(p), prefix
}

// claudePermissions writes the rules unchanged into .claude/settings.json.
func claudePermissions(r Rules, managed []string) (Output, []string) {
	want := map[string][]string{"allow": r.Allow, "ask": r.Ask, "deny": r.Deny}
	return bucketMergeOutput(".claude/settings.json", want, managed), nil
}

// cursorPermissions maps rules onto the Cursor CLI's allow/deny tokens.
// Cursor has no ask. A Bash rule translates only when it names a single
// command word (Shell(cmd) matches that command).
func cursorPermissions(r Rules, managed []string) (Output, []string) {
	want := map[string][]string{}
	var unsupported []string
	for _, b := range r.buckets() {
		for _, rule := range b.rules {
			tok, ok := cursorToken(b.name, rule)
			if !ok || b.name == "ask" {
				unsupported = append(unsupported, b.name+": "+rule)
				continue
			}
			want[b.name] = append(want[b.name], tok)
		}
	}
	return bucketMergeOutput(".cursor/cli.json", want, managed), unsupported
}

// cursorToken translates one rule. Shell(cmd) matches cmd with any
// arguments, so an allow translates only from a `cmd:*` prefix rule (an
// exact allow would broaden). Paths translate only when project-relative:
// Claude's `/`, `//`, and `~/` anchors have no verified Cursor meaning.
func cursorToken(bucket, rule string) (string, bool) {
	tool, spec := parseRule(rule)
	switch tool {
	case "Bash":
		return cursorShell(bucket, spec)
	case "Read":
		return "Read(" + spec + ")", relativePath(spec)
	case "Edit", "Write":
		return "Write(" + spec + ")", relativePath(spec)
	case "WebFetch":
		if d, ok := strings.CutPrefix(spec, "domain:"); ok {
			return "WebFetch(" + d + ")", true
		}
	}
	return "", false
}

// cursorShell is Shell(cmd) for a single-word Bash rule that does not
// broaden: an allow needs `cmd:*`, since Shell(cmd) takes any arguments.
func cursorShell(bucket, spec string) (string, bool) {
	words, prefix := bashPrefix(spec)
	if len(words) != 1 {
		return "", false
	}
	return "Shell(" + words[0] + ")", prefix || bucket != "allow"
}

func relativePath(spec string) bool {
	return spec != "" && !strings.HasPrefix(spec, "/") && !strings.HasPrefix(spec, "~")
}

// codexPermissions owns .codex/rules/keystone.rules (Starlark prefix rules;
// Codex applies the most restrictive match). Only Bash rules translate. A
// prefix rule matches more than an exact command, so an exact-command allow
// is not translated (it would broaden); an exact deny or ask may be.
func codexPermissions(r Rules, _ []string) (Output, []string) {
	var b strings.Builder
	b.WriteString("# Generated by keystone project from .agents/permissions/. Edit the sources, not this file.\n")
	var unsupported []string
	decision := map[string]string{"allow": "allow", "ask": "prompt", "deny": "forbidden"}
	for _, bk := range r.buckets() {
		for _, rule := range bk.rules {
			words, ok := codexPattern(bk.name, rule)
			if !ok {
				unsupported = append(unsupported, bk.name+": "+rule)
				continue
			}
			fmt.Fprintf(&b, "prefix_rule(pattern = [%s], decision = %q)\n", starlarkList(words), decision[bk.name])
		}
	}
	return Output{Dest: ".codex/rules/keystone.rules", Strategy: Shim, Content: []byte(b.String())}, unsupported
}

// codexPattern is a Bash rule's prefix words, unless translating it would
// broaden an allow (an exact command allow) or it is not a Bash rule.
func codexPattern(bucket, rule string) ([]string, bool) {
	tool, spec := parseRule(rule)
	if tool != "Bash" {
		return nil, false
	}
	words, prefix := bashPrefix(spec)
	if len(words) == 0 {
		return nil, false
	}
	return words, prefix || bucket != "allow"
}

func starlarkList(words []string) string {
	quoted := make([]string, len(words))
	for i, w := range words {
		quoted[i] = strconv.Quote(w)
	}
	return strings.Join(quoted, ", ")
}

// bucketMergeOutput merges allow/ask/deny lists under "permissions" in a
// shared JSON config file. managed is what keystone added last time
// ("bucket:entry"), so a rule dropped from the charter is removed — but an
// entry the user had before keystone added the same one stays.
func bucketMergeOutput(dest string, want map[string][]string, managed []string) Output {
	return Output{
		Dest:     dest,
		Strategy: Merge,
		MergeFn:  func(prev []byte) ([]byte, error) { return mergeBuckets(prev, want, managed) },
		RecordFn: func(prev []byte) []string { return recordBuckets(prev, want, managed) },
	}
}

var bucketNames = []string{"allow", "ask", "deny"}

func mergeBuckets(prev []byte, want map[string][]string, managed []string) ([]byte, error) {
	doc, err := readJSONObject(prev)
	if err != nil {
		return nil, err
	}
	perm, _ := doc["permissions"].(map[string]any)
	if perm == nil {
		perm = map[string]any{}
	}
	recorded := toSet(managed)
	changed := false
	for _, name := range bucketNames {
		changed = mergeBucket(perm, name, want[name], recorded) || changed
	}
	if !changed {
		return prev, nil // no rewrite, no reformat, no new file
	}
	doc["permissions"] = perm
	if len(perm) == 0 {
		delete(doc, "permissions")
	}
	return marshalJSON(doc)
}

// mergeBucket updates one permission list in perm; reports whether it changed.
func mergeBucket(perm map[string]any, name string, want []string, recorded map[string]bool) bool {
	cur := jsonStrings(perm[name])
	next := mergeList(cur, want, func(e string) bool { return recorded[name+":"+e] })
	if equalStrings(cur, next) {
		return false
	}
	if len(next) == 0 {
		delete(perm, name)
	} else {
		perm[name] = next
	}
	return true
}

// mergeList keeps cur's order, drops entries keystone recorded that are no
// longer wanted, and appends wanted entries not yet present.
func mergeList(cur, want []string, recorded func(string) bool) []string {
	return appendMissing(dropStale(cur, want, recorded), want)
}

// dropStale removes entries keystone recorded that are no longer wanted.
func dropStale(cur, want []string, recorded func(string) bool) []string {
	wanted := toSet(want)
	var out []string
	for _, e := range cur {
		if !recorded(e) || wanted[e] {
			out = append(out, e)
		}
	}
	return out
}

// appendMissing appends wanted entries not yet present, in order.
func appendMissing(cur, want []string) []string {
	have := toSet(cur)
	for _, e := range want {
		if !have[e] {
			cur = append(cur, e)
			have[e] = true
		}
	}
	return cur
}

// recordBuckets is the next managed record: wanted entries keystone owns —
// ones it recorded before, or ones absent from the file until now.
func recordBuckets(prev []byte, want map[string][]string, managed []string) []string {
	doc, _ := readJSONObject(prev)
	perm, _ := doc["permissions"].(map[string]any)
	recorded := toSet(managed)
	var out []string
	for _, name := range bucketNames {
		existing := toSet(jsonStrings(perm[name]))
		for _, e := range want[name] {
			if k := name + ":" + e; recorded[k] || !existing[e] {
				out = append(out, k)
			}
		}
	}
	return out
}

func jsonStrings(v any) []string {
	raw, _ := v.([]any)
	var out []string
	for _, e := range raw {
		if s, ok := e.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
