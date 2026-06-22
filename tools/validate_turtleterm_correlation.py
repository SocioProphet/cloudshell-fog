#!/usr/bin/env python3
"""Validate CloudShell FOG ↔ TurtleTerm correlation fixture."""

from __future__ import annotations

import json
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[1]
FIXTURE = REPO_ROOT / "fixtures" / "turtleterm" / "cloudshell-turtleterm-correlation.v0.1.json"


def main() -> int:
    data = json.loads(FIXTURE.read_text(encoding="utf-8"))
    assert data["schema"] == "cloudshell-fog.turtleterm.correlation.v0.1"

    cloudshell = data["cloudshell_session"]
    env = data["turtleterm_environment"]
    expected = data["expected_receipt_context"]

    assert env["SOURCEOS_TERMINAL_SESSION_ID"] == cloudshell["session_id"]
    assert env["SOURCEOS_WORKSPACE"] == cloudshell["workspace_id"]
    assert env["SOURCEOS_ACTOR_ID"] == cloudshell["subject"]
    assert env["SOURCEOS_POLICY_BUNDLE_ID"] == cloudshell["policy_bundle_id"]
    assert env["SOURCEOS_EXECUTION_DOMAIN"] == cloudshell["runtime"]["execution_domain"]

    assert expected["session_id"] == env["SOURCEOS_TERMINAL_SESSION_ID"]
    assert expected["workspace_id"] == env["SOURCEOS_WORKSPACE"]
    assert expected["actor_id"] == env["SOURCEOS_ACTOR_ID"]
    assert expected["policy_bundle_id"] == env["SOURCEOS_POLICY_BUNDLE_ID"]
    assert expected["execution_domain"] == env["SOURCEOS_EXECUTION_DOMAIN"]

    placement = cloudshell["placement"]
    assert placement["region"]
    assert placement["node_id"]
    assert placement["tier"]
    assert isinstance(placement["reasons"], list) and placement["reasons"]

    runtime = cloudshell["runtime"]
    assert runtime["namespace"].startswith("cloudshell-")
    assert runtime["pod"]
    assert "@sha256:" in runtime["image_ref"]

    for key in data["correlation_keys"]:
        assert key in expected, f"missing correlation key in expected context: {key}"

    print(f"validated {FIXTURE}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
