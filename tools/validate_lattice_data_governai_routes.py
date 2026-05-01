#!/usr/bin/env python3
"""Validate Lattice Studio/Data/GovernAI Developer Home route fixtures."""

from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[1]
FIXTURE = ROOT / "fixtures" / "lattice-data-governai" / "developer-home-routes.v0.1.json"

REQUIRED_COMMANDS = {
    "lattice.catalog.search.community-truth-demo",
    "lattice.data.inspect.community-truth-demo",
    "lattice.runtime.pick.prophet-python-ml",
    "lattice.notebook.launch.community-truth-demo",
    "lattice.mlops.run.ray-beam-demo",
    "lattice.govern.review.model-eval",
    "lattice.publication.inspect.report",
    "lattice.publication.export.report.blocked",
}
REQUIRED_SOURCE_REFS = {
    "platformPr",
    "runtimePr",
    "mlopsPr",
    "policyPr",
    "topologyPr",
    "sherlockPr",
    "slashTopicsPr",
    "newHopePr",
    "agentplanePr",
}
EXECUTION_ACTIONS = {"run-notebook-dry-run", "run-governed-mlops-dry-run"}


def fail(message: str) -> int:
    print(f"ERR: {message}", file=sys.stderr)
    return 1


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def require_str(mapping: dict[str, Any], key: str) -> str:
    value = mapping.get(key)
    require(isinstance(value, str) and bool(value), f"{key} must be a non-empty string")
    return value


def require_list(mapping: dict[str, Any], key: str) -> list[Any]:
    value = mapping.get(key)
    require(isinstance(value, list) and value, f"{key} must be a non-empty list")
    return value


def validate_command(command: dict[str, Any]) -> str:
    command_id = require_str(command, "id")
    require(command_id in REQUIRED_COMMANDS, f"unexpected command id {command_id}")
    require_str(command, "label")
    require(require_str(command, "slashCommand").startswith("/lattice "), f"{command_id} slashCommand must start with /lattice")
    route = require_str(command, "route")
    action = require_str(command, "action")
    require_str(command, "assetRef")
    require_str(command, "policyRef")
    require_str(command, "evidenceCorrelationId")
    requires = require_list(command, "requires")
    result_mode = require_str(command, "resultMode")
    if action in EXECUTION_ACTIONS:
        require(route == "agentplane", f"{command_id} execution commands must route through AgentPlane")
        require("AgentPlane" in requires, f"{command_id} must require AgentPlane")
        require(result_mode == "execution-dry-run", f"{command_id} must be execution-dry-run")
    if action == "export-artifact":
        require(route == "policy-fabric", "publication export must route through Policy Fabric")
        require(result_mode == "denied", "publication export fixture must be denied")
        require("review.accepted" in requires, "publication export must require accepted review")
    if route == "agentplane":
        require("PolicyFabric" in requires, f"{command_id} AgentPlane route must still require PolicyFabric")
    if route == "policy-fabric":
        require("PolicyFabric" in requires, f"{command_id} Policy Fabric route must declare PolicyFabric requirement")
    return command_id


def main() -> int:
    if not FIXTURE.exists():
        return fail(f"missing {FIXTURE}")
    try:
        data = json.loads(FIXTURE.read_text(encoding="utf-8"))
        require(isinstance(data, dict), "fixture root must be object")
        require(data.get("apiVersion") == "cloudshell.socioprophet.dev/v0", "apiVersion mismatch")
        require(data.get("kind") == "LatticeDeveloperHomeRoutesFixture", "kind mismatch")
        metadata = data.get("metadata")
        require(isinstance(metadata, dict), "metadata must be object")
        require(metadata.get("name") == "lattice-data-governai-developer-home-routes", "metadata.name mismatch")

        refs = data.get("sourceRefs")
        require(isinstance(refs, dict), "sourceRefs must be object")
        missing_refs = sorted(REQUIRED_SOURCE_REFS - set(refs))
        require(not missing_refs, f"missing sourceRefs: {missing_refs}")

        surface = data.get("surfaceModel")
        require(isinstance(surface, dict), "surfaceModel must be object")
        require(surface.get("shellSurface") == "cloudshell-fog", "shellSurface mismatch")
        require(surface.get("developerHomeSurface") == "fog-shell-command-palette", "developerHomeSurface mismatch")
        require(surface.get("mustRouteThroughPolicyFabric") is True, "mustRouteThroughPolicyFabric must be true")
        require(surface.get("mustRouteThroughAgentPlaneForExecution") is True, "mustRouteThroughAgentPlaneForExecution must be true")
        require(surface.get("mustPreservePlatformAssetRecordIdentity") is True, "mustPreservePlatformAssetRecordIdentity must be true")
        require(surface.get("mustNotBypassSlashTopicsOrNewHope") is True, "mustNotBypassSlashTopicsOrNewHope must be true")

        assets = data.get("assets")
        require(isinstance(assets, dict), "assets must be object")
        require(assets.get("dataProductRef") == "urn:srcos:data-product:community_truth_demo", "dataProductRef mismatch")
        require(assets.get("runtimeAssetRef") == "runtime-asset:prophet-python-ml:0.1.0", "runtimeAssetRef mismatch")
        require(assets.get("policyPackRef") == "SocioProphet/policy-fabric#39", "policyPackRef mismatch")
        require(assets.get("agentplaneReplayRef") == "SocioProphet/agentplane#76", "agentplaneReplayRef mismatch")
        require(assets.get("topicPackRef") == "slash-topics://packs/lattice-data-governai@0.1.0", "topicPackRef mismatch")
        require(assets.get("semanticMembraneRef") == "newhope://membranes/lattice-data-governai-admission@0.1.0", "semanticMembraneRef mismatch")

        commands = data.get("commands")
        require(isinstance(commands, list) and commands, "commands must be non-empty list")
        seen = {validate_command(command) for command in commands if isinstance(command, dict)}
        missing_commands = sorted(REQUIRED_COMMANDS - seen)
        require(not missing_commands, f"missing commands: {missing_commands}")

        safety = data.get("safety")
        require(isinstance(safety, dict), "safety must be object")
        require(safety.get("fixtureOnly") is True, "fixtureOnly must be true")
        require(safety.get("hostMutation") is False, "hostMutation must be false")
        require(safety.get("network") == "none", "network must be none")
        require(safety.get("secrets") == "none", "secrets must be none")
        require(safety.get("runtimeConnectorBypassAllowed") is False, "runtimeConnectorBypassAllowed must be false")
        require(safety.get("policyBypassAllowed") is False, "policyBypassAllowed must be false")
    except Exception as exc:  # noqa: BLE001
        return fail(str(exc))
    print(f"PASS {FIXTURE}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
