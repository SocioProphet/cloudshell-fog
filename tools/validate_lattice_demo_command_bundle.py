#!/usr/bin/env python3
"""Validate the Lattice Studio/Data/GovernAI demo command bundle."""

from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[1]
FIXTURE = ROOT / "fixtures" / "lattice-data-governai" / "demo-command-bundle.v0.1.json"

NOTEBOOK = "runtime-asset:prophet-python-ml:0.1.0"
RAY = "runtime-asset:prophet-ray-ml:0.1.0"
BEAM = "runtime-asset:prophet-beam-dataops:0.1.0"
BINDING = "runtime-profile-binding:lattice-data-governai:0.1.0"
PROMOTION_MANIFEST = "runtime-promotion-manifest:lattice-runtime-promotion-manifest:0.1.0"

REQUIRED_SOURCE_REFS = {
    "demoReadinessPr",
    "demoReadinessTopologyPr",
    "runtimeForgePr",
    "runtimePromotionPr",
    "policyRuntimePromotionPr",
    "agentplaneRuntimeRefsPr",
    "runtimeRoutesPr",
}
REQUIRED_COMMANDS = [
    "/lattice data search community_truth_demo",
    "/lattice runtime pick prophet-python-ml",
    "/lattice runtime pick prophet-ray-ml",
    "/lattice runtime pick prophet-beam-dataops",
    "/lattice runtime bindings inspect",
    "/lattice notebook launch community_truth_demo --runtime prophet-python-ml",
    "/lattice data inspect urn:srcos:data-product:community_truth_demo",
    "/lattice mlops ray run community_truth_demo --runtime prophet-ray-ml --dry-run",
    "/lattice dataops beam run community_truth_demo --runtime prophet-beam-dataops --dry-run",
    "/lattice govern review urn:srcos:evaluation-bundle:community_truth_demo_model_eval",
    "/lattice publication inspect urn:srcos:publication-artifact:community_truth_demo_report",
    "/lattice publication export urn:srcos:publication-artifact:community_truth_demo_report",
]
EXECUTION_ACTION_IDS = {
    "demo.notebook.launch": NOTEBOOK,
    "demo.ray.run": RAY,
    "demo.beam.run": BEAM,
}
POLICY_ROUTES = {"demo.govern.review", "demo.publication.export.blocked"}


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
    require(isinstance(value, list) and bool(value), f"{key} must be a non-empty list")
    return value


def validate_command(command: dict[str, Any], expected_step: int, expected_command: str) -> str:
    require(command.get("step") == expected_step, f"step {expected_step} expected")
    command_id = require_str(command, "id")
    require(require_str(command, "command") == expected_command, f"step {expected_step} command mismatch")
    route = require_str(command, "route")
    mode = require_str(command, "expectedMode")
    output = command.get("expectedOutput")
    require(isinstance(output, dict), f"{command_id}: expectedOutput must be object")
    require_str(output, "status")

    if command_id in EXECUTION_ACTION_IDS:
        require(route == "agentplane", f"{command_id}: execution must route to AgentPlane")
        require(mode == "execution-dry-run", f"{command_id}: must be execution-dry-run")
        require(output.get("runtimeRef") == EXECUTION_ACTION_IDS[command_id], f"{command_id}: runtimeRef mismatch")
        require_str(output, "evidenceRef")
    if command_id in POLICY_ROUTES:
        require(route == "policy-fabric", f"{command_id}: governance must route to Policy Fabric")
    if command_id.startswith("demo.runtime.pick"):
        require(route == "lattice-forge", f"{command_id}: runtime picks must route to Lattice Forge")
        require(mode == "read-only", f"{command_id}: runtime picks must be read-only")
        require(output.get("assetRef") in {NOTEBOOK, RAY, BEAM}, f"{command_id}: assetRef must be runtime asset")
        require_str(output, "policyRef")
    if command_id == "demo.runtime.bindings.inspect":
        require(route == "prophet-platform", "runtime bindings inspect must route to Prophet Platform")
        require(output.get("assetRef") == BINDING, "runtime binding assetRef mismatch")
        runtime_refs = set(require_list(output, "containsRuntimeRefs"))
        require(runtime_refs == {NOTEBOOK, RAY, BEAM}, "runtime binding output must include all runtime refs")
    if command_id == "demo.publication.export.blocked":
        require(mode == "denied", "publication export must be denied")
        require(output.get("status") == "denied", "publication export status must be denied")
        because = require_list(output, "because")
        require(any("stable promotion remains blocked" in str(item) for item in because), "publication export must mention stable promotion blocker")
    return command_id


def main() -> int:
    if not FIXTURE.exists():
        return fail(f"missing {FIXTURE}")
    try:
        data = json.loads(FIXTURE.read_text(encoding="utf-8"))
        require(data.get("apiVersion") == "cloudshell.socioprophet.dev/v0", "apiVersion mismatch")
        require(data.get("kind") == "LatticeDemoCommandBundleFixture", "kind mismatch")
        refs = data.get("sourceRefs")
        require(isinstance(refs, dict), "sourceRefs must be object")
        missing_refs = sorted(REQUIRED_SOURCE_REFS - set(refs))
        require(not missing_refs, f"missing sourceRefs: {missing_refs}")
        require(refs.get("demoReadinessPr") == "SocioProphet/prophet-platform#307", "demoReadinessPr mismatch")
        require(refs.get("demoReadinessTopologyPr") == "SocioProphet/sociosphere#244", "demoReadinessTopologyPr mismatch")
        require(refs.get("policyRuntimePromotionPr") == "SocioProphet/policy-fabric#42", "policyRuntimePromotionPr mismatch")

        runtime_refs = data.get("runtimeRefs")
        require(isinstance(runtime_refs, dict), "runtimeRefs must be object")
        require(runtime_refs.get("notebookRuntimeRef") == NOTEBOOK, "notebook runtime mismatch")
        require(runtime_refs.get("rayRuntimeRef") == RAY, "ray runtime mismatch")
        require(runtime_refs.get("beamRuntimeRef") == BEAM, "beam runtime mismatch")
        require(runtime_refs.get("runtimeProfileBindingRef") == BINDING, "runtime binding mismatch")
        require(runtime_refs.get("runtimePromotionManifestRef") == PROMOTION_MANIFEST, "runtime promotion manifest mismatch")

        state = data.get("demoState")
        require(isinstance(state, dict), "demoState must be object")
        require(state.get("readiness") == "demo-ready", "readiness must be demo-ready")
        require(state.get("network") == "none", "network must be none")
        require(state.get("secrets") == "none", "secrets must be none")
        require(state.get("hostMutation") is False, "hostMutation must be false")
        require(state.get("devRuntimePromotion") == "allowed-with-generated-evidence", "dev runtime promotion mismatch")
        require(state.get("stableRuntimePromotion") == "blocked-pending-external-evidence", "stable runtime promotion mismatch")

        commands = require_list(data, "commands")
        require(len(commands) == len(REQUIRED_COMMANDS), "command count mismatch")
        seen = [validate_command(command, i, expected) for i, (command, expected) in enumerate(zip(commands, REQUIRED_COMMANDS), start=1)]
        require(len(set(seen)) == len(seen), "command ids must be unique")

        requirements = data.get("validationRequirements")
        require(isinstance(requirements, dict), "validationRequirements must be object")
        for key in [
            "requireSequentialSteps",
            "requireRuntimeProfileCommands",
            "requireAgentPlaneForExecution",
            "requirePolicyFabricForGovernance",
            "requireStablePromotionBlocked",
            "requireNoNetworkSecretsOrHostMutation",
        ]:
            require(requirements.get(key) is True, f"{key} must be true")
    except Exception as exc:  # noqa: BLE001
        return fail(str(exc))
    print(f"PASS {FIXTURE}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
