#!/usr/bin/env python3
from __future__ import annotations

import json
import os
import re
import subprocess
import sys
from pathlib import Path

ALLOWED_WITHOUT_PLAN_PREFIXES = [
    ".codex/plans/",
    ".codex/artifacts/",
    ".codex/hooks/",
    ".codex/config.toml",
    ".codex/rules/",
    "AGENTS.md",
    "README.md",
]

REQUIRED_SECTIONS = [
    "## Context",
    "## 完了基準",
    "## Phase Checklist",
    "## Agent Assignment",
    "## Review Loop",
]

READ_ONLY_COMMAND_RE = re.compile(
    r"^\s*(pwd|ls|find|rg|sed|cat|head|tail|wc|git status|git diff|git show|git branch|git log|git fetch|date|which|echo)\b"
)
CLAUDE_COMMAND_RE = re.compile(r"(^|\s)(?:\S*/)?claude(?:\s|$)|claude-plan-review|claude-code-review|claude-analysis")
FORBIDDEN_COMMAND_RE = re.compile(r"\b(git push|rm -rf /|rm -rf \.|DROP TABLE|TRUNCATE)\b", re.IGNORECASE)


def repo_root() -> Path:
    value = os.environ.get("CODEX_PROJECT_ROOT", "").strip()
    if value:
        return Path(value).resolve()
    result = subprocess.run(["git", "rev-parse", "--show-toplevel"], check=False, capture_output=True, text=True)
    if result.returncode == 0 and result.stdout.strip():
        return Path(result.stdout.strip()).resolve()
    return Path.cwd().resolve()


def load_input() -> dict:
    raw = sys.stdin.read().strip()
    if not raw:
        return {}
    try:
        return json.loads(raw)
    except json.JSONDecodeError:
        return {}


def deep_get(data, *paths):
    for path in paths:
        current = data
        ok = True
        for part in path:
            if isinstance(current, dict) and part in current:
                current = current[part]
            else:
                ok = False
                break
        if ok:
            return current
    return None


def frontmatter(text: str) -> dict[str, str]:
    if not text.startswith("---\n"):
        return {}
    parts = text.split("\n---\n", 1)
    if len(parts) != 2:
        return {}
    data = {}
    for line in parts[0].splitlines()[1:]:
        if ":" not in line:
            continue
        key, value = line.split(":", 1)
        data[key.strip()] = value.strip().strip('"')
    return data


def latest_plan() -> Path | None:
    active_dir = repo_root() / ".codex" / "plans" / "active"
    candidates = sorted([p for p in active_dir.glob("*.md") if p.name != "README.md"], key=lambda p: p.stat().st_mtime, reverse=True)
    return candidates[0] if candidates else None


def missing_sections(text: str) -> list[str]:
    return [section for section in REQUIRED_SECTIONS if section not in text]


def tool_name(payload: dict) -> str:
    return str(deep_get(payload, ("tool_name",)) or deep_get(payload, ("toolName",)) or deep_get(payload, ("hook_event", "tool_name")) or "")


def command_text(payload: dict) -> str:
    return str(deep_get(payload, ("tool_input", "command")) or deep_get(payload, ("toolInput", "command")) or deep_get(payload, ("arguments", "command")) or "")


def target_paths(payload: dict) -> list[str]:
    values = []
    for path in [
        ("tool_input", "path"),
        ("tool_input", "file_path"),
        ("tool_input", "paths"),
        ("toolInput", "path"),
        ("toolInput", "file_path"),
        ("toolInput", "paths"),
        ("arguments", "path"),
        ("arguments", "file_path"),
        ("arguments", "paths"),
    ]:
        value = deep_get(payload, path)
        if not value:
            continue
        if isinstance(value, list):
            values.extend(str(v) for v in value)
        else:
            values.append(str(value))
    return values


def is_bootstrap_target(paths: list[str]) -> bool:
    if not paths:
        return False
    for path in paths:
        normalized = path.replace("\\", "/")
        if ".codex/" in normalized and not normalized.startswith(".codex/"):
            normalized = normalized[normalized.index(".codex/"):]
        if normalized.startswith("./"):
            normalized = normalized[2:]
        normalized = normalized.lstrip("/")
        if any(normalized.startswith(prefix) for prefix in ALLOWED_WITHOUT_PLAN_PREFIXES):
            return True
    return False


def emit_block(reason: str) -> None:
    print(json.dumps({"decision": "block", "reason": reason}))
    raise SystemExit(0)


def main() -> None:
    payload = load_input()
    name = tool_name(payload)
    cmd = command_text(payload)
    paths = target_paths(payload)

    if name == "Bash" and CLAUDE_COMMAND_RE.search(cmd):
        emit_block("This repository is Codex-only. Do not execute Claude commands here.")
    if name == "Bash" and FORBIDDEN_COMMAND_RE.search(cmd):
        emit_block("Forbidden dangerous command detected. Update the plan and use a safer workflow.")
    if name == "Bash" and READ_ONLY_COMMAND_RE.match(cmd):
        return
    if name in {"Edit", "Write", "MultiEdit"} and is_bootstrap_target(paths):
        return

    plan = latest_plan()
    if not plan:
        emit_block("Active plan is required before implementation work. Create one in .codex/plans/active first.")

    text = plan.read_text()
    meta = frontmatter(text)
    missing = missing_sections(text)
    if missing:
        emit_block("Active plan is missing required sections: " + ", ".join(missing))

    status = meta.get("status", "").strip()
    if status not in {"approved", "in_progress"}:
        emit_block(f"Active plan status must be approved or in_progress. Current: {status or 'unset'}.")


if __name__ == "__main__":
    main()
