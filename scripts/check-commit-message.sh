#!/usr/bin/env sh
# Reject commit messages that contain AI attribution or co-author trailers.
set -eu

pattern='(claude|deepseek|anthropic|co[ -]?author(ed)?)'

check_file() {
	file=$1
	# Drop comment lines and the `git commit -v` scissors section so only the
	# real message is checked.
	hits=$(sed -e '/^# ------------------------ >8 ------------------------$/,$d' -e '/^#/d' "$file" |
		grep -Ein "$pattern" || true)
	if [ -n "$hits" ]; then
		echo "commit message rejected: AI attribution and co-author trailers are forbidden." >&2
		echo "offending text:" >&2
		printf '%s\n' "$hits" | sed 's/^/  /' >&2
		echo "See the commit-message policy in CONTRIBUTING.md." >&2
		return 1
	fi
}

case "${1:-}" in
--history)
	limit=${2:-50}
	tmp=$(mktemp)
	trap 'rm -f "$tmp"' EXIT
	for sha in $(git rev-list --max-count="$limit" HEAD); do
		git show -s --format=%B "$sha" >"$tmp"
		if ! check_file "$tmp"; then
			echo "commit: $sha" >&2
			exit 1
		fi
	done
	;;
"")
	echo "usage: $0 <message-file> | --history [count]" >&2
	exit 2
	;;
*)
	for file in "$@"; do
		check_file "$file" || exit 1
	done
	;;
esac
