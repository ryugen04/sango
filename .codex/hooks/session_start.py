#!/usr/bin/env python3
from __future__ import annotations

import json
import os
import re
import subprocess
from pathlib import Path

REQUIRED_SECTIONS = [
    "## Context",
    "## 完了基準",
    "## Phase Checklist",
    "## Agent Assignment",
    "## Review Loop",
]


def repo_root() -> Path:
    value = os.environ.get("CODEX_PROJECT_ROOT", "").strip()
    if value:
        return Path(value).resolve()
    result = subprocess.run(["git", "rev-parse", "--show-toplevel"], check=False, capture_output=True, text=True)
    if result.returncode == 0 and result.stdout.strip():
        return Path(result.stdout.strip()).resolve()
    return Path.cwd().resolve()


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


def unchecked_tasks(text: str) -> list[str]:
    return re.findall(r"^- \[ \] (.+)$", text, re.MULTILINE)


def main() -> None:
    plan = latest_plan()
    if not plan:
        print(json.dumps({"hookSpecificOutput": {"hookEventName": "SessionStart", "additionalContext": "No active plan found. Create one in .codex/plans/active before implementation."}}))
        return

    text = plan.read_text()
    meta = frontmatter(text)
    lines = [
        f"Active plan: {plan.relative_to(repo_root())}",
        f"Status: {meta.get('status', 'unknown')}",
    ]
    missing = missing_sections(text)
    if missing:
        lines.append("Missing required sections: " + ", ".join(missing))
    tasks = unchecked_tasks(text)[:5]
    if tasks:
        lines.append("Next unchecked tasks:")
        lines.extend(f"- {task}" for task in tasks)
    print(json.dumps({"hookSpecificOutput": {"hookEventName": "SessionStart", "additionalContext": "\n".join(lines)}}))


if __name__ == "__main__":
    main()
