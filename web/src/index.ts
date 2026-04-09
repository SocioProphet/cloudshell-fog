import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";

// Gateway base URL — override window.GATEWAY_URL from a server-rendered config block.
declare global {
  interface Window {
    GATEWAY_URL?: string;
  }
}

const GATEWAY_HTTP =
  window.GATEWAY_URL ||
  `${location.protocol}//${location.host}`;

const SESSION_API = `${GATEWAY_HTTP}/v1/sessions`;

// ─── Message frame types (interfaces-v1 §1.3) ─────────────────────────────────

interface ResizeFrame {
  type: "resize";
  cols: number;
  rows: number;
}
interface StdinFrame {
  type: "stdin";
  data_b64: string;
}
interface StdoutFrame {
  type: "stdout";
  data_b64: string;
}
interface ExitFrame {
  type: "exit";
  code: number;
}
type Frame = ResizeFrame | StdinFrame | StdoutFrame | ExitFrame;

// ─── Base64 helpers ────────────────────────────────────────────────────────────

function b64encode(s: string): string {
  return btoa(unescape(encodeURIComponent(s)));
}

function b64decode(s: string): string {
  return decodeURIComponent(escape(atob(s)));
}

// ─── State ────────────────────────────────────────────────────────────────────

let term: Terminal;
let fitAddon: FitAddon;
let ws: WebSocket | null = null;
let sessionId: string | null = null;

// ─── UI helpers ───────────────────────────────────────────────────────────────

function setStatus(msg: string): void {
  const el = document.getElementById("status");
  if (el) el.textContent = msg;
}

function setTerminateEnabled(enabled: boolean): void {
  const btn = document.getElementById("btn-terminate") as HTMLButtonElement | null;
  if (!btn) return;
  if (enabled) {
    btn.removeAttribute("disabled");
  } else {
    btn.setAttribute("disabled", "");
  }
}

// ─── OIDC access-token retrieval ──────────────────────────────────────────────
// In production: exchange auth-code+PKCE for tokens and store in sessionStorage.
// Here we read the token if previously stored (e.g. by an OIDC redirect handler).
function getAccessToken(): string | null {
  return sessionStorage.getItem("access_token");
}

// ─── API calls ────────────────────────────────────────────────────────────────

interface CreateSessionResponse {
  session_id: string;
  attach: {
    ws_url: string;
    token: string;
    expires_at: string;
  };
  placement: string;
}

async function createSession(): Promise<CreateSessionResponse> {
  const token = getAccessToken();
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (token) headers["Authorization"] = `Bearer ${token}`;

  const resp = await fetch(SESSION_API, {
    method: "POST",
    headers,
    body: JSON.stringify({ profile: "default", ttl_seconds: 3600 }),
  });
  if (!resp.ok) {
    throw new Error(`create session: ${resp.status} ${await resp.text()}`);
  }
  return resp.json() as Promise<CreateSessionResponse>;
}

async function terminateSession(): Promise<void> {
  if (!sessionId) return;
  const token = getAccessToken();
  const headers: Record<string, string> = {};
  if (token) headers["Authorization"] = `Bearer ${token}`;
  await fetch(`${SESSION_API}/${sessionId}`, { method: "DELETE", headers });
  sessionId = null;
}

// ─── WebSocket ────────────────────────────────────────────────────────────────

function sendResize(cols: number, rows: number): void {
  if (ws?.readyState === WebSocket.OPEN) {
    const f: ResizeFrame = { type: "resize", cols, rows };
    ws.send(JSON.stringify(f));
  }
}

function connectWebSocket(attach: CreateSessionResponse["attach"]): void {
  // Build the wss:// URL — replace the HTTP base with ws base.
  const wsUrl = attach.ws_url.replace(/^http/, "ws");
  const fullUrl = `${wsUrl}?token=${encodeURIComponent(attach.token)}`;

  ws = new WebSocket(fullUrl);

  ws.onopen = () => {
    setStatus("Connected");
    sendResize(term.cols, term.rows);
  };

  ws.onclose = () => {
    setStatus("Disconnected");
    ws = null;
    setTerminateEnabled(false);
  };

  ws.onerror = (e) => console.error("WebSocket error", e);

  ws.onmessage = (ev: MessageEvent<string>) => {
    const f = JSON.parse(ev.data) as Frame;
    if (f.type === "stdout") {
      term.write(b64decode(f.data_b64));
    } else if (f.type === "exit") {
      term.writeln(`\r\n[Process exited with code ${f.code}]`);
      ws?.close();
    }
  };
}

// ─── Connect / terminate actions ──────────────────────────────────────────────

async function connect(): Promise<void> {
  setStatus("Creating session…");
  try {
    const resp = await createSession();
    sessionId = resp.session_id;
    setStatus(`Session ${resp.session_id} · placement: ${resp.placement}`);
    connectWebSocket(resp.attach);
    setTerminateEnabled(true);
  } catch (e) {
    setStatus(`Error: ${String(e)}`);
  }
}

// ─── Entrypoint ───────────────────────────────────────────────────────────────

window.addEventListener("DOMContentLoaded", () => {
  term = new Terminal({
    cursorBlink: true,
    theme: { background: "#1e1e1e" },
    fontFamily: "monospace",
  });
  fitAddon = new FitAddon();
  term.loadAddon(fitAddon);

  const termEl = document.getElementById("terminal");
  if (termEl) {
    term.open(termEl);
    fitAddon.fit();
  }

  // Forward keyboard input to PTY stdin.
  term.onData((data) => {
    if (ws?.readyState === WebSocket.OPEN) {
      const f: StdinFrame = { type: "stdin", data_b64: b64encode(data) };
      ws.send(JSON.stringify(f));
    }
  });

  // Resize terminal and notify PTY when the browser window is resized.
  window.addEventListener("resize", () => {
    fitAddon.fit();
    sendResize(term.cols, term.rows);
  });

  document.getElementById("btn-connect")?.addEventListener("click", () => {
    void connect();
  });

  document.getElementById("btn-terminate")?.addEventListener("click", async () => {
    ws?.close();
    await terminateSession();
    setStatus("Terminated");
    setTerminateEnabled(false);
  });

  // Auto-connect if a gateway URL has been injected (e.g. embedded console).
  if (window.GATEWAY_URL) {
    void connect();
  }
});
