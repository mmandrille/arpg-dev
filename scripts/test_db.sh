#!/usr/bin/env bash
# Per-checkout test database for CI/bot/benchmark runs.
#
# Every git worktree gets its own database inside the shared arpg-postgres
# container, so concurrent runs cannot share bot accounts/characters, see each
# other's coop sessions or market rows, reset each other's connection flags on
# server boot, or migrate one schema from two different branches.
#
#   test_db.sh url            print this checkout's default test database URL
#   test_db.sh ensure [URL]   create the database if missing (local container only)
#   test_db.sh prune          drop arpg_test_* databases whose checkout is gone
#
# The name is arpg_test_<checkout basename>_<cksum of absolute path>; the path is
# stored as the database comment so `prune` can tell which ones are orphaned.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONTAINER="${ARPG_PG_CONTAINER:-arpg-postgres}"
PG_USER="${ARPG_PG_USER:-arpg}"

default_name() {
  local slug hash
  slug="$(basename "$ROOT" | tr 'A-Z' 'a-z' | tr -c 'a-z0-9\n' '_' | cut -c1-32)"
  hash="$(printf '%s' "$ROOT" | cksum | cut -d' ' -f1)"
  printf 'arpg_test_%s_%s\n' "$slug" "$hash"
}

default_url() {
  printf 'postgres://arpg:arpg@localhost:5432/%s?sslmode=disable\n' "$(default_name)"
}

psql_admin() {
  docker exec -i "$CONTAINER" psql -v ON_ERROR_STOP=1 -U "$PG_USER" -d postgres -tAq "$@"
}

ensure() {
  local url="${1:-$(default_url)}"
  local re='^postgres(ql)?://[^@/]*@(localhost|127\.0\.0\.1)(:5432)?/([a-z0-9_]+)(\?.*)?$'
  if [[ ! "$url" =~ $re ]]; then
    # Not the local container (or an unusual name): the caller owns provisioning.
    return 0
  fi
  local name="${BASH_REMATCH[4]}"
  if [[ "$name" == "arpg" ]]; then
    return 0
  fi
  if [[ -n "$(psql_admin -c "SELECT 1 FROM pg_database WHERE datname = '$name'")" ]]; then
    return 0
  fi
  # A concurrent run in the same checkout may win the race; that is fine.
  if psql_admin -c "CREATE DATABASE \"$name\"" >/dev/null 2>&1; then
    local comment="${ROOT//\'/\'\'}"
    psql_admin -c "COMMENT ON DATABASE \"$name\" IS '$comment'" >/dev/null
    echo "[test-db] created $name"
  elif [[ -z "$(psql_admin -c "SELECT 1 FROM pg_database WHERE datname = '$name'")" ]]; then
    echo "[test-db] failed to create database $name" >&2
    return 1
  fi
}

prune() {
  local name path dropped=0
  while IFS='|' read -r name path; do
    [[ -z "$name" ]] && continue
    if [[ -n "$path" && -d "$path" ]]; then
      continue
    fi
    echo "[test-db] dropping $name (checkout ${path:-<unknown>} is gone)"
    psql_admin -c "DROP DATABASE IF EXISTS \"$name\" WITH (FORCE)" >/dev/null
    dropped=$((dropped + 1))
  done < <(psql_admin -F '|' -c "SELECT d.datname, coalesce(shobj_description(d.oid, 'pg_database'), '')
                                  FROM pg_database d WHERE d.datname LIKE 'arpg\_test\_%' ORDER BY 1")
  echo "[test-db] pruned $dropped database(s)"
}

case "${1:-}" in
  url) default_url ;;
  ensure) ensure "${2:-}" ;;
  prune) prune ;;
  *)
    echo "usage: $0 url | ensure [DATABASE_URL] | prune" >&2
    exit 2
    ;;
esac
