#!/usr/bin/env bash
#
# Create one epic issue per wave and attach the entity issues as real GitHub
# sub-issues. Idempotent: existing epics are reused, existing links are skipped.
#
# Usage:
#   scripts/create-epics.sh              # dry run
#   scripts/create-epics.sh --apply
#
# Requires: gh (authenticated). Run scripts/create-issues.sh first.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DATA="${REPO_ROOT}/scripts/entities.tsv"
APPLY=0

while [ $# -gt 0 ]; do
	case "$1" in
	--apply) APPLY=1 ;;
	--dry-run) APPLY=0 ;;
	-h | --help)
		sed -n '2,12p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
		exit 0
		;;
	*)
		echo "unknown argument: $1" >&2
		exit 2
		;;
	esac
	shift
done

[ -f "$DATA" ] || {
	echo "missing $DATA" >&2
	exit 1
}

repo="$(gh repo view --json nameWithOwner -q .nameWithOwner)"
echo "repository: $repo"

wave_title() {
	case "$1" in
	1) echo "Taxonomy and governance" ;;
	2) echo "Database stack" ;;
	3) echo "BI stack" ;;
	4) echo "Pipelines and messaging" ;;
	5) echo "Platform services" ;;
	6) echo "Drives and documents" ;;
	7) echo "API assets" ;;
	8) echo "Data quality and operations" ;;
	9) echo "Access and organisation" ;;
	10) echo "AI and learning" ;;
	esac
}

wave_goal() {
	case "$1" in
	1) echo "Settle the entities everything else references. Classifications, tags, glossary terms, domains, data products and metrics are attached to later assets by name, so the naming, lookup and import conventions get decided here rather than being re-litigated in every later wave." ;;
	2) echo "Build the deepest hierarchy OpenMetadata has: service to database to schema to table, plus stored procedures and queries. This wave proves nested fully-qualified-name handling and parent-child imports before the shape is copied everywhere else." ;;
	3) echo "Dashboards, charts and dashboard data models. Same service-to-asset shape as the database stack, but with BI-specific lineage fields and a flatter second level, so it tests how much of wave 2 genuinely generalises." ;;
	4) echo "Orchestration pipelines and messaging topics. Small wave: service-scoped assets with fewer nested levels, useful for confirming the service registration pattern is not database-specific." ;;
	5) echo "The remaining service types plus storage containers, ML models and search indexes. This wave settles service registration for every non-database provider and is the last place a structural surprise should appear." ;;
	6) echo "Drive tree: drives, directories, files, spreadsheets, worksheets and documents. Tree-shaped like the database stack, but with permission and ownership semantics that do not exist on data assets." ;;
	7) echo "API services, collections and endpoints: a third tree shape, and a small enough wave to check the pattern one more time before the bulk work lands." ;;
	8) echo "Data quality and operational automation. Test cases attach to tables from wave 2 and data contracts attach to data products from wave 1, so this wave cannot start earlier." ;;
	9) echo "Identity and authorisation: users, teams, roles, policies, bots and personas. High blast radius if a plan is wrong, which is why it lands after the asset model is stable." ;;
	10) echo "The newest and least stable API surface. Deliberately last so SDK churn cannot block the stable entities, and so any upstream fixes can be batched here." ;;
	esac
}

wave_depends() {
	case "$1" in
	1) echo "none" ;;
	2 | 3 | 4 | 5 | 6 | 7) echo "wave-01 (tags and classifications for asset metadata)" ;;
	8) echo "wave-01 and wave-02 (test cases attach to tables, contracts to data products)" ;;
	9) echo "wave-01" ;;
	10) echo "wave-01" ;;
	esac
}

# --- snapshot of existing issues -------------------------------------------

issues_json="$(gh issue list --repo "$repo" --state all --limit 300 \
	--json number,id,title,state)"
issue_number() { # title -> number
	printf '%s' "$issues_json" | jq -r --arg t "$1" \
		'.[] | select(.title == $t) | .number' | head -1
}
issue_node() { # title -> node id
	printf '%s' "$issues_json" | jq -r --arg t "$1" \
		'.[] | select(.title == $t) | .id' | head -1
}

if [ "$APPLY" -eq 1 ]; then
	gh label create epic --repo "$repo" --color "6f42c1" \
		--description "Wave-level tracking issue" --force >/dev/null 2>&1 || true
fi

# --- epic body -------------------------------------------------------------

epic_body() {
	local wave="$1" children="$2"
	cat <<EOF
## Goal

$(wave_goal "$wave")

## Scope

Tracks wave ${wave} of the entity rollout: $(wave_title "$wave").
See \`TASKS.md\` for the full 58-entity inventory and the parallel work contract.

Children: ${children} entity issue(s), listed below. Every child copies the
reference slice in \`internal/resources/glossary\` and touches no shared files
beyond one alphabetical line in \`internal/provider/provider.go\`.

## Dependencies

Depends on: $(wave_depends "$wave").

## Exit criteria

- [ ] Every child issue closed with its resource and data source merged
- [ ] \`go build ./...\`, \`go vet ./...\`, \`go test ./...\` green on \`main\`
- [ ] \`go generate ./...\` produces no docs diff
- [ ] Acceptance suite green in CI for the entities in this wave
- [ ] No shared-file edits accumulated: \`provider.go\` still has exactly one
      line per entity and nothing else moved
EOF
}

# --- create and link -------------------------------------------------------

created=0
reused=0
linked=0
already=0

for wave in 1 2 3 4 5 6 7 8 9 10; do
	title="[epic] Wave $(printf '%02d' "$wave") — $(wave_title "$wave")"

	entities=()
	while IFS=$'\t' read -r entity w base has_delete; do
		[ "$entity" = "entity" ] && continue
		[ "$w" = "$wave" ] && entities+=("$entity")
	done <"$DATA"

	epic_number="$(issue_number "$title")"

	if [ "$APPLY" -eq 0 ]; then
		if [ -n "$epic_number" ]; then
			echo "exists  #${epic_number}  ${title}"
		else
			echo "would create  ${title}  [epic, wave-$(printf '%02d' "$wave")]  (${#entities[@]} children)"
		fi
		continue
	fi

	if [ -z "$epic_number" ]; then
		epic_number="$(epic_body "$wave" "${#entities[@]}" |
			gh issue create --repo "$repo" --title "$title" \
				--label epic --label "wave-$(printf '%02d' "$wave")" \
				--body-file - 2>&1 | grep -oE '/issues/[0-9]+$' | grep -oE '[0-9]+' | tail -1)"
		echo "created  #${epic_number}  ${title}"
		created=$((created + 1))
		# The node id is only known after creation.
		epic_node="$(gh api graphql -f query="query { repository(owner: \"${repo%%/*}\", name: \"${repo##*/}\") { issue(number: ${epic_number}) { id } } }" --jq '.data.repository.issue.id')"
	else
		echo "reuse    #${epic_number}  ${title}"
		reused=$((reused + 1))
		epic_node="$(issue_node "$title")"
	fi

	# Which children are already attached?
	existing_children="$(gh api graphql \
		-f query='query($owner:String!,$name:String!,$num:Int!){ repository(owner:$owner,name:$name){ issue(number:$num){ subIssues(first:100){ nodes{ number } } } } }' \
		-f owner="${repo%%/*}" -f name="${repo##*/}" -F num="$epic_number" \
		--jq '[.data.repository.issue.subIssues.nodes[].number] | join(" ")' 2>/dev/null || true)"

	for entity in "${entities[@]}"; do
		child_title="[entity] openmetadata_${entity} resource and data source"
		child_number="$(issue_number "$child_title")"
		if [ -z "$child_number" ]; then
			echo "  missing child issue for ${entity}; run create-issues.sh first" >&2
			continue
		fi

		case " ${existing_children} " in
		*" ${child_number} "*)
			already=$((already + 1))
			continue
			;;
		esac

		child_node="$(issue_node "$child_title")"
		gh api graphql \
			-f query='mutation($parent:ID!,$child:ID!){ addSubIssue(input:{issueId:$parent, subIssueId:$child}){ subIssue{ number } } }' \
			-f parent="$epic_node" -f child="$child_node" >/dev/null
		linked=$((linked + 1))
	done
done

echo
if [ "$APPLY" -eq 0 ]; then
	echo "dry run: re-run with --apply to create the epics and link children."
else
	echo "done: ${created} epic(s) created, ${reused} reused, ${linked} sub-issue link(s) added, ${already} already linked."
fi
