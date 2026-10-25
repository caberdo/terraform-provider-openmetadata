#!/usr/bin/env bash
#
# Create one GitHub issue per OpenMetadata entity, from scripts/entities.tsv.
#
# The script is idempotent: an entity is skipped when an issue with the exact
# same title already exists in any state, so it can be re-run after a partial
# failure without producing duplicates.
#
# Usage:
#   scripts/create-issues.sh              # dry run, prints what it would do
#   scripts/create-issues.sh --apply      # creates labels and issues
#   scripts/create-issues.sh --only wave-1 --apply
#   scripts/create-issues.sh --entity tag --apply
#
# Requires: gh (authenticated), and the working tree rooted at the provider.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DATA="${REPO_ROOT}/scripts/entities.tsv"
APPLY=0
ONLY_WAVE=""
ONLY_ENTITY=""

while [ $# -gt 0 ]; do
	case "$1" in
	--apply) APPLY=1 ;;
	--dry-run) APPLY=0 ;;
	--only)
		shift
		case "${1:-}" in
		wave-*) ONLY_WAVE="${1#wave-}" ;;
		*) ONLY_ENTITY="${1:-}" ;;
		esac
		;;
	--entity)
		shift
		ONLY_ENTITY="${1:-}"
		;;
	-h | --help)
		sed -n '2,18p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
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

wave_name() {
	case "$1" in
	0) echo "foundations" ;;
	1) echo "taxonomy" ;;
	2) echo "database" ;;
	3) echo "bi" ;;
	4) echo "pipelines" ;;
	5) echo "platform-services" ;;
	6) echo "drives" ;;
	7) echo "api" ;;
	8) echo "quality-ops" ;;
	9) echo "access" ;;
	10) echo "ai" ;;
	*) echo "unknown" ;;
	esac
}

wave_depends() {
	case "$1" in
	0) echo "none (reference slice)" ;;
	1) echo "none" ;;
	2) echo "wave-1 (classification/tag for tagging assets)" ;;
	3) echo "wave-1" ;;
	4) echo "wave-1" ;;
	5) echo "wave-1" ;;
	6) echo "wave-1" ;;
	7) echo "wave-1" ;;
	8) echo "wave-1, wave-2 (test cases attach to tables)" ;;
	9) echo "wave-1" ;;
	10) echo "wave-1" ;;
	*) echo "unknown" ;;
	esac
}

# --- validation -------------------------------------------------------------

rows=0
seen=""
dupes=""
while IFS=$'\t' read -r entity wave base has_delete; do
	[ "$entity" = "entity" ] && continue
	[ -z "$entity" ] && continue
	rows=$((rows + 1))
	case " ${seen} " in
	*" ${entity} "*) dupes="${dupes} ${entity}" ;;
	esac
	seen="${seen} ${entity}"
	case "$wave" in
	0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10) ;;
	*)
		echo "invalid wave '$wave' for entity '$entity'" >&2
		exit 1
		;;
	esac
done <"$DATA"

if [ -n "$dupes" ]; then
	echo "duplicate entities in $DATA:$dupes" >&2
	exit 1
fi
if [ "$rows" -ne 58 ]; then
	echo "expected 58 entities in $DATA, found $rows" >&2
	exit 1
fi

# --- repo and labels --------------------------------------------------------

repo="$(gh repo view --json nameWithOwner -q .nameWithOwner)"
echo "repository: $repo"

labels=("entity")
for w in 0 1 2 3 4 5 6 7 8 9 10; do
	labels+=("wave-$(printf '%02d' "$w")")
done
labels+=("sdk-blocker")

label_color() {
	case "$1" in
	entity) echo "5319e7" ;;
	sdk-blocker) echo "b60205" ;;
	*) echo "0e8a16" ;;
	esac
}

if [ "$APPLY" -eq 1 ]; then
	for label in "${labels[@]}"; do
		gh label create "$label" --repo "$repo" --color "$(label_color "$label")" \
			--description "Parallel entity work item" --force >/dev/null 2>&1 || true
	done
fi

# --- issue body -------------------------------------------------------------

issue_body() {
	local entity="$1" wave="$2" base="$3" has_delete="$4"
	local blocker=""
	if [ "$has_delete" = "no" ]; then
		blocker="
## Blocker

The SDK service for this entity exposes **no \`Delete\` method**, so a normal
Terraform destroy path does not exist yet. Resolve this before writing CRUD:

- Check whether the REST API supports \`DELETE /api/v1/${base}/{id}\` or a
  soft delete through \`PATCH\` with \`{\"op\":\"add\",\"path\":\"/deleted\",\"value\":true}\`.
- If neither works, add a narrow raw-HTTP helper in \`internal/client\` or
  contribute the missing method upstream.

Do not ship a no-op \`Delete\`: it leaves orphaned metadata and breaks
\`terraform destroy\` guarantees.
"
	fi

	cat <<EOF
## Entity

\`openmetadata_${entity}\` — SDK service base path \`${base}\`.

## Scope

- [ ] \`internal/resources/${entity}/resource.go\` — CRUD + \`ImportState\`
- [ ] \`internal/resources/${entity}/model.go\` — API ⇄ Terraform mapping, \`patchOps\`
- [ ] \`internal/resources/${entity}/resource_test.go\` — metadata, schema, mapping, patch ops
- [ ] \`internal/datasources/${entity}/datasource.go\` — lookup by id or full name
- [ ] \`internal/datasources/${entity}/datasource_test.go\`
- [ ] \`internal/resources/${entity}/acceptance_test.go\` — \`package ${entity}_test\`
- [ ] \`internal/datasources/${entity}/acceptance_test.go\` — \`package ${entity}_test\`
- [ ] \`examples/resources/openmetadata_${entity}/resource.tf\` and \`import.sh\`
- [ ] \`examples/data-sources/openmetadata_${entity}/data-source.tf\`
- [ ] Exactly one alphabetical insertion in \`internal/provider/provider.go\`
- [ ] \`go generate ./...\` and commit the regenerated docs

## Pattern to copy

Do not invent a new architecture. Copy the reference slice verbatim and adapt:

- \`internal/resources/glossary/resource.go\`
- \`internal/datasources/glossary/datasource.go\`

Framework rules already established in this repo:

- Plugin **Framework only** — SDKv2 is not allowed.
- The SDK has no \`Update\`; use \`Patch\` with \`ometa.JSONPatchOp\`, sending only changed fields.
- \`Delete\` takes a params struct (\`Recursive\`/\`HardDelete\` where supported).
- Acceptance tests must live in \`package ${entity}_test\`; importing
  \`internal/acctest\` from the internal package creates an import cycle.
- \`Default\` on an attribute requires \`Computed: true\` as well.
- The provider only registers; resources reach the client through
  \`*providerdata.ProviderData\`.

## Definition of done

\`\`\`
go build ./...
go vet ./...
go test ./...
go generate ./...
git diff --exit-code docs
\`\`\`

Plus a passing acceptance run against the compose stack (CI only, no local Docker).

## Parallel work contract

- Own only the files listed above, plus your single line in \`provider.go\`.
- Do not touch \`internal/client\`, \`internal/convert\`, or another entity's package.
  If you need a shared helper, say so in the PR instead of editing shared files.

## Dependencies

Wave ${wave} ($(wave_name "$wave")) — depends on: $(wave_depends "$wave").
${blocker}
EOF
}

# --- create -----------------------------------------------------------------

created=0
skipped=0
planned=0

while IFS=$'\t' read -r entity wave base has_delete; do
	[ "$entity" = "entity" ] && continue
	[ -z "$entity" ] && continue

	# Wave 0 is the reference slice, already shipped.
	[ "$wave" = "0" ] && continue
	if [ -n "$ONLY_WAVE" ] && [ "$wave" != "$ONLY_WAVE" ]; then continue; fi
	if [ -n "$ONLY_ENTITY" ] && [ "$entity" != "$ONLY_ENTITY" ]; then continue; fi

	title="[entity] openmetadata_${entity} resource and data source"
	label_wave="wave-$(printf '%02d' "$wave")"
	extra_labels="$label_wave"
	if [ "$has_delete" = "no" ]; then
		extra_labels="$extra_labels,sdk-blocker"
	fi

	existing="$(gh issue list --repo "$repo" --state all --search "in:title \"${title}\"" \
		--json number,title --jq ".[] | select(.title == \"${title}\") | .number" 2>/dev/null | head -1 || true)"

	if [ -n "$existing" ]; then
		echo "skip  #${existing}  openmetadata_${entity} (already exists)"
		skipped=$((skipped + 1))
		continue
	fi

	if [ "$APPLY" -eq 0 ]; then
		echo "would create  ${title}  [${extra_labels}]"
		planned=$((planned + 1))
		continue
	fi

	url="$(issue_body "$entity" "$wave" "$base" "$has_delete" |
		gh issue create --repo "$repo" --title "$title" \
			--label "entity" --label "$label_wave" \
			$( [ "$has_delete" = "no" ] && printf '%s' '--label sdk-blocker' ) \
			--body-file - 2>&1 | tail -1)"

	echo "created  ${url}  openmetadata_${entity}"
	created=$((created + 1))
done <"$DATA"

echo
if [ "$APPLY" -eq 0 ]; then
	echo "dry run: ${planned} issue(s) to create, ${skipped} already present. Re-run with --apply."
else
	echo "done: ${created} created, ${skipped} already present."
fi
