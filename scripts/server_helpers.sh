#!/usr/bin/env bash
# Shared "start my own arpg-server and prove I'm talking to it" helpers for the
# local CI / bot / benchmark wrappers. Source after quiet_helpers.sh.
#
# Concurrent worktrees share one machine, so a fixed port is unsafe: a second run
# fails to bind, and its readiness probe can pass against the *other* checkout's
# server. The default is therefore ARPG_ADDR=:0 (kernel-picked free port). The
# server logs `server listening` with the bound addr + its pid only after a
# successful bind; we read that line from *our* server's log, require the pid to
# be ours, and only then probe /readyz on that port while the pid is alive.
#
# Inputs (globals): ADDR, BASE_URL (either may be empty), SERVER_PID, SERVER_LOG.
#   ADDR empty, BASE_URL empty -> ADDR=:0, BASE_URL derived from the bound port
#   ADDR empty, BASE_URL set   -> ADDR=:<BASE_URL port>
#   ADDR set                   -> bound port must equal ADDR's port (unless :0)
#   BASE_URL set               -> its port must equal the bound port
# Output (globals): BASE_URL, SERVER_PORT.

arpg_url_port() {
  local re='^[a-z]+://[^/]*:([0-9]+)(/.*)?$'
  [[ "$1" =~ $re ]] && printf '%s\n' "${BASH_REMATCH[1]}"
}

# Database name only (no credentials) for log lines.
arpg_db_name() {
  local name="${1##*/}"
  printf '%s\n' "${name%%\?*}"
}

# Normalise ADDR from BASE_URL before the server starts.
arpg_resolve_server_addr() {
  if [[ -n "${ADDR:-}" ]]; then
    return 0
  fi
  if [[ -n "${BASE_URL:-}" ]]; then
    local port
    port="$(arpg_url_port "$BASE_URL")" || {
      echo "BASE_URL=$BASE_URL has no explicit port; set ARPG_ADDR too" >&2
      return 1
    }
    ADDR=":$port"
  else
    ADDR=":0"
  fi
}

# Sets LISTEN_PORT / LISTEN_PID from the server's own "server listening" line.
arpg_read_listening_line() {
  local line re_addr re_pid
  re_addr='"addr":"[^"]*:([0-9]+)"'
  re_pid='"pid":([0-9]+)'
  line="$(grep -m1 '"message":"server listening"' "$SERVER_LOG" 2>/dev/null)" || return 1
  [[ "$line" =~ $re_addr ]] || return 1
  LISTEN_PORT="${BASH_REMATCH[1]}"
  [[ "$line" =~ $re_pid ]] || return 1
  LISTEN_PID="${BASH_REMATCH[1]}"
}

arpg_server_died() {
  local label="$1"
  echo "[$label] server pid=$SERVER_PID exited before becoming ready; log:"
  show_log "$SERVER_LOG" "server"
  if grep -q "address already in use" "$SERVER_LOG" 2>/dev/null; then
    echo "[$label] port ${ADDR##*:} is taken (another worktree's run?). Drop the explicit" \
      "ARPG_ADDR/CI_ADDR/BOT_ADDR override to get a free per-run port."
  fi
}

# arpg_wait_own_server <label> [timeout_s]: returns 0 once our server is ready.
arpg_wait_own_server() {
  local label="$1"
  local timeout_s="${2:-90}"
  local deadline=$((SECONDS + timeout_s))
  LISTEN_PORT=""
  LISTEN_PID=""

  while ! arpg_read_listening_line; do
    if ! kill -0 "$SERVER_PID" >/dev/null 2>&1; then
      arpg_server_died "$label"
      return 1
    fi
    if (( SECONDS >= deadline )); then
      echo "[$label] server pid=$SERVER_PID did not log 'server listening' within ${timeout_s}s; log:"
      show_log "$SERVER_LOG" "server"
      return 1
    fi
    sleep 0.2
  done

  if [[ "$LISTEN_PID" != "$SERVER_PID" ]]; then
    echo "[$label] $SERVER_LOG reports pid=$LISTEN_PID, expected our server pid=$SERVER_PID"
    return 1
  fi
  local want_port="${ADDR##*:}"
  if [[ "$want_port" != "0" && "$want_port" != "$LISTEN_PORT" ]]; then
    echo "[$label] server bound port $LISTEN_PORT but ARPG_ADDR=$ADDR"
    return 1
  fi
  if [[ -n "${BASE_URL:-}" ]]; then
    local url_port
    url_port="$(arpg_url_port "$BASE_URL" || true)"
    if [[ "$url_port" != "$LISTEN_PORT" ]]; then
      echo "[$label] BASE_URL=$BASE_URL does not point at our server (port $LISTEN_PORT)"
      return 1
    fi
  else
    BASE_URL="http://127.0.0.1:$LISTEN_PORT"
  fi
  SERVER_PORT="$LISTEN_PORT"

  # The port is held by our live pid, so a /readyz answer on it is ours.
  until curl -fsS "${BASE_URL%/}/readyz" >/dev/null 2>&1; do
    if ! kill -0 "$SERVER_PID" >/dev/null 2>&1; then
      arpg_server_died "$label"
      return 1
    fi
    if (( SECONDS >= deadline )); then
      echo "[$label] $BASE_URL/readyz not ready within ${timeout_s}s; log:"
      show_log "$SERVER_LOG" "server"
      return 1
    fi
    sleep 0.5
  done
  if ! kill -0 "$SERVER_PID" >/dev/null 2>&1; then
    arpg_server_died "$label"
    return 1
  fi
  echo "[$label] server pid=$SERVER_PID ready at $BASE_URL"
}
