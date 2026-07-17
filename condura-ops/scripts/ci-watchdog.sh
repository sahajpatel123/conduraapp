#!/bin/bash
# CI watchdog wrapper — invoked by launchd every 10 minutes.
# Runs `claude` headlessly with the watchdog prompt, logs all output.
set -uo pipefail

REPO="/Users/sahajpatel/synaptic"
LOG_DIR="/Users/sahajpatel/Library/Logs/ci-watchdog"
LOG_FILE="$LOG_DIR/watchdog.log"
LOCK_FILE="$LOG_DIR/.watchdog.lock"
PROMPT_FILE="$LOG_DIR/prompt.txt"
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S %Z')

mkdir -p "$LOG_DIR"

# Refuse to overlap with a previous run that is still going.
if [ -f "$LOCK_FILE" ]; then
  LOCK_PID=$(cat "$LOCK_FILE" 2>/dev/null || echo "")
  if [ -n "$LOCK_PID" ] && kill -0 "$LOCK_PID" 2>/dev/null; then
    echo "[$TIMESTAMP] watchdog: another run is in progress (pid=$LOCK_PID), skipping" >> "$LOG_FILE"
    exit 0
  fi
  # Stale lock — remove it.
  rm -f "$LOCK_FILE"
fi
echo $$ > "$LOCK_FILE"
trap 'rm -f "$LOCK_FILE"' EXIT

cd "$REPO" || {
  echo "[$TIMESTAMP] watchdog: failed to cd to $REPO" >> "$LOG_FILE"
  exit 1
}

# Verify the working tree is clean (never commit/push dirty work).
if [ -n "$(git status --porcelain 2>/dev/null)" ]; then
  echo "[$TIMESTAMP] watchdog: working tree is dirty, skipping to avoid losing changes" >> "$LOG_FILE"
  exit 0
fi

# Verify gh is authenticated.
if ! gh auth status >/dev/null 2>&1; then
  echo "[$TIMESTAMP] watchdog: gh not authenticated, skipping" >> "$LOG_FILE"
  exit 0
fi

# Write the prompt to a file so the launchd invocation stays short.
cat > "$PROMPT_FILE" <<'EOF'
CI watchdog loop: Check all GitHub Actions / CI checks on the current repo (use `gh` CLI). For the current branch and any open PRs, find failing or in-progress runs. For each failure, investigate the logs, fix the underlying issue in the code, then commit and push the fix so CI re-runs. Always commit AND push every change you make. If everything is already green, report that and make no changes. Keep the working tree clean and never force-push over shared history.
EOF

echo "[$TIMESTAMP] watchdog: starting run (pid=$$)" >> "$LOG_FILE"

# Run claude headlessly with the prompt. -p = print/non-interactive mode.
# Stream the output into the log so the user can review it later.
claude -p --model claude-haiku-4-5-20251001 --effort low "$(cat $PROMPT_FILE)" \
  >> "$LOG_FILE" 2>&1
EXIT_CODE=$?

echo "[$TIMESTAMP] watchdog: run complete (exit=$EXIT_CODE)" >> "$LOG_FILE"
exit $EXIT_CODE
