#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
FIXTURE = ROOT / "fixtures" / "lattice-data-governai" / "runtime-release-command-bundle.v0.1.json"
MANIFEST = "runtime-promotion-manifest:lattice-runtime-promotion-manifest:0.2.0"
RUNTIME_REFS = {
    "runtime-asset:prophet-python-ml:0.1.0",
    "runtime-asset:prophet-ray-ml:0.1.0",
    "runtime-asset:prophet-beam-dataops:0.1.0",
}
COMMANDS = [
    "/lattice runtime release manifest inspect --manifest lattice-runtime-promotion-manifest:0.2.0",
    "/lattice runtime release policy inspect --manifest lattice-runtime-promotion-manifest:0.2.0",
    "/lattice runtime release readiness inspect",
]


def fail(message: str) -> int:
    print(f"ERR: {message}", file=sys.stderr)
    return 1


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def main() -> int:
    if not FIXTURE.exists():
        return fail(f"missing {FIXTURE}")
    try:
        data = json.loads(FIXTURE.read_text(encoding="utf-8"))
        require(data.get("apiVersion") == "cloudshell.socioprophet.dev/v0", "apiVersion mismatch")
        require(data.get("kind") == "LatticeRuntimeReleaseCommandBundleFixture", "kind mismatch")
        refs = data.get("sourceRefs")
        require(isinstance(refs, dict), "sourceRefs must be object")
        require(refs.get("runtimeEvidencePr") == "SocioProphet/lattice-forge#13", "runtimeEvidencePr mismatch")
        require(refs.get("runtimePolicyPr") == "SocioProphet/policy-fabric#43", "runtimePolicyPr mismatch")
        require(refs.get("platformRuntimeReleaseReadinessPr") == "SocioProphet/prophet-platform#308", "platformRuntimeReleaseReadinessPr mismatch")
        require(data.get("manifestRef") == MANIFEST, "manifestRef mismatch")
        require(set(data.get("runtimeRefs", [])) == RUNTIME_REFS, "runtimeRefs mismatch")
        commands = data.get("commands")
        require(isinstance(commands, list) and len(commands) == 3, "commands mismatch")
        for expected_step, expected_command in enumerate(COMMANDS, start=1):
            command = commands[expected_step - 1]
            require(command.get("step") == expected_step, "step mismatch")
            require(command.get("command") == expected_command, "command mismatch")
            require(command.get("expectedMode") == "read-only", "expectedMode mismatch")
            output = command.get("expectedOutput")
            require(isinstance(output, dict), "expectedOutput must be object")
            require(output.get("status") == "ok", "expected status mismatch")
        require(commands[0].get("route") == "lattice-forge", "manifest route mismatch")
        require(commands[1].get("route") == "policy-fabric", "policy route mismatch")
        require(commands[2].get("route") == "prophet-platform", "readiness route mismatch")
        require(commands[0]["expectedOutput"].get("manifestRef") == MANIFEST, "manifest output mismatch")
        require(commands[1]["expectedOutput"].get("policyRef") == "SocioProphet/policy-fabric#43", "policy output mismatch")
        require(commands[2]["expectedOutput"].get("readinessRef") == "SocioProphet/prophet-platform#308", "readiness output mismatch")
        requirements = data.get("validationRequirements")
        require(isinstance(requirements, dict), "validationRequirements must be object")
        for key in ["requireSequentialSteps", "requireManifestV020", "requirePolicyFabricForReleaseDecision", "requireAllRuntimeRefs", "requireNoNetworkSecretsOrHostMutation"]:
            require(requirements.get(key) is True, f"{key} must be true")
        require(data.get("safety") == {"network": "none", "secrets": "none", "hostMutation": False}, "safety mismatch")
    except Exception as exc:  # noqa: BLE001
        return fail(str(exc))
    print(f"PASS {FIXTURE}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
