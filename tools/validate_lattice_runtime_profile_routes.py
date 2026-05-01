#!/usr/bin/env python3
"""Validate Lattice runtime profile Developer Home route fixtures."""

from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[1]
FIXTURE = ROOT / "fixtures" / "lattice-data-governai" / "runtime-profile-routes.v0.1.json"
NOTEBOOK = "runtime-asset:prophet-python-ml:0.1.0"
RAY = "runtime-asset:prophet-ray-ml:0.1.0"
BEAM = "runtime-asset:prophet-beam-dataops:0.1.0"
BINDING = "runtime-profile-binding:lattice-data-governai:0.1.0"
REQUIRED_COMMANDS = {
    "lattice.runtime.pick.notebook",
    "lattice.runtime.pick.ray",
    "lattice.runtime.pick.beam",
    "lattice.runtime.bindings.inspect",
    "lattice.notebook.launch.with-profile",
    "lattice.mlops.ray.run.with-profile",
    "lattice.dataops.beam.run.with-profile",
}
REQUIRED_SOURCE_REFS = {
    "runtimeForgePr",
    "platformRuntimeCatalogPr",
    "agentplaneRuntimeRefsPr",
    "topologyRuntimePr",
    "sherlockRuntimeIndexPr",
    "slashRuntimeTopicPr",
    "newHopeRuntimeMembranePr",
    "policyRuntimePr",
}


def fail(message: str) -> int:
    print(f"ERR: {message}", file=sys.stderr)
    return 1


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def require_str(mapping: dict[str, Any], key: str) -> str:
    value = mapping.get(key)
    require(isinstance(value, str) and bool(value), f"{key} must be non-empty string")
    return value


def require_list(mapping: dict[str, Any], key: str) -> list[Any]:
    value = mapping.get(key)
    require(isinstance(value, list) and value, f"{key} must be non-empty list")
    return value


def validate_command(command: dict[str, Any]) -> str:
    command_id = require_str(command, "id")
    require(command_id in REQUIRED_COMMANDS, f"unexpected command {command_id}")
    require(require_str(command, "slashCommand").startswith("/lattice "), "slashCommand must start with /lattice")
    route = require_str(command, "route")
    action = require_str(command, "action")
    require_str(command, "assetRef")
    require_str(command, "policyRef")
    require_str(command, "evidenceCorrelationId")
    requires = require_list(command, "requires")
    require("PolicyFabric" in requires, f"{command_id} must require PolicyFabric")
    result_mode = require_str(command, "resultMode")
    if action in {"run-notebook-dry-run", "run-ray-runtime", "run-beam-runtime"}:
        require(route == "agentplane", f"{command_id} execution route must be AgentPlane")
        require("AgentPlane" in requires, f"{command_id} must require AgentPlane")
        require("RuntimeProfileBinding" in requires, f"{command_id} must require RuntimeProfileBinding")
        require(result_mode == "execution-dry-run", f"{command_id} must be execution-dry-run")
        runtime_ref = require_str(command, "runtimeRef")
        if action == "run-notebook-dry-run":
            require(runtime_ref == NOTEBOOK, "notebook command runtimeRef mismatch")
        if action == "run-ray-runtime":
            require(runtime_ref == RAY, "Ray command runtimeRef mismatch")
        if action == "run-beam-runtime":
            require(runtime_ref == BEAM, "Beam command runtimeRef mismatch")
    if action == "select-runtime":
        require(route == "lattice-forge", f"{command_id} runtime selection must route to lattice-forge")
        require("RuntimeAsset" in requires, f"{command_id} must require RuntimeAsset")
        require("RuntimeProfileBinding" in requires, f"{command_id} must require RuntimeProfileBinding")
        require(result_mode == "read-only", f"{command_id} runtime pick must be read-only")
    if action == "inspect-runtime-profile-bindings":
        require(command["assetRef"] == BINDING, "binding inspect assetRef mismatch")
        require("RuntimeProfileBinding" in requires, "binding inspect must require RuntimeProfileBinding")
    return command_id


def main() -> int:
    if not FIXTURE.exists():
        return fail(f"missing {FIXTURE}")
    try:
        data = json.loads(FIXTURE.read_text(encoding="utf-8"))
        require(data.get("apiVersion") == "cloudshell.socioprophet.dev/v0", "apiVersion mismatch")
        require(data.get("kind") == "LatticeRuntimeProfileRoutesFixture", "kind mismatch")
        refs = data.get("sourceRefs")
        require(isinstance(refs, dict), "sourceRefs must be object")
        missing = sorted(REQUIRED_SOURCE_REFS - set(refs))
        require(not missing, f"missing sourceRefs: {missing}")
        runtime_refs = data.get("runtimeRefs")
        require(isinstance(runtime_refs, dict), "runtimeRefs must be object")
        require(runtime_refs.get("notebookRuntimeRef") == NOTEBOOK, "notebook runtime ref mismatch")
        require(runtime_refs.get("rayRuntimeRef") == RAY, "ray runtime ref mismatch")
        require(runtime_refs.get("beamRuntimeRef") == BEAM, "beam runtime ref mismatch")
        require(runtime_refs.get("runtimeProfileBindingRef") == BINDING, "runtime profile binding ref mismatch")
        surface = data.get("surfaceModel")
        require(isinstance(surface, dict), "surfaceModel must be object")
        require(surface.get("mustRouteThroughPolicyFabric") is True, "mustRouteThroughPolicyFabric must be true")
        require(surface.get("mustRouteThroughAgentPlaneForExecution") is True, "mustRouteThroughAgentPlaneForExecution must be true")
        require(surface.get("mustPreserveRuntimeAssetRefs") is True, "mustPreserveRuntimeAssetRefs must be true")
        require(surface.get("mustNotBypassSlashTopicsOrNewHope") is True, "mustNotBypassSlashTopicsOrNewHope must be true")
        commands = data.get("commands")
        require(isinstance(commands, list) and commands, "commands must be non-empty list")
        seen = {validate_command(command) for command in commands if isinstance(command, dict)}
        require(seen == REQUIRED_COMMANDS, f"command set mismatch: {sorted(seen)}")
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
