// Singleton WS-клиент (один сокет на комнату), без дублей сообщений.
// WS подключается напрямую к room-service.
// Конфиг через VITE_ROOM_WS_BASE, например: ws://localhost:8082
import { tokenVault } from "./tokenVault";

export type WsEvent =
  | { type: "open" }
  | { type: "close" }
  | { type: "error"; error: any }
  | { type: "state"; payload: any }
  | { type: "peer_joined"; payload: any }
  | { type: "peer_left"; payload: any }
  | { type: "chat"; payload: any }
  | { type: "chat_ack"; payload: any };

type Listener = (ev: WsEvent) => void;

const ROOM_WS_BASE = (import.meta as any).env?.VITE_ROOM_WS_BASE?.replace(/\/+$/, "") || "";

function buildDirectWsUrl(roomId: string) {
  const at = tokenVault.getAccessToken() || "";
  const uid = tokenVault.getUserId();
  if (!ROOM_WS_BASE) throw new Error("VITE_ROOM_WS_BASE is not set");
  const qs = new URLSearchParams({
    access_token: at,
    user_id: uid != null ? String(uid) : "",
  }).toString();
  return `${ROOM_WS_BASE}/ws/rooms/${encodeURIComponent(roomId)}?${qs}`;
}

class RoomWS {
  private ws: WebSocket | null = null;
  private roomId: string | null = null;
  private listeners = new Set<Listener>();
  private reconnecting = false;
  private stop = false;
  private backoff = 500; // ms
  private seen = new Set<string>(); // msg_id дедуп

  on(fn: Listener) { this.listeners.add(fn); return () => this.listeners.delete(fn); }
  private emit(ev: WsEvent) { for (const fn of this.listeners) fn(ev); }

  async connect(roomId: string) {
    this.stop = false;
    this.roomId = roomId;
    this.seen.clear();

    const url = buildDirectWsUrl(roomId);
    this.ws = new WebSocket(url);

    this.ws.onopen = () => {
      this.backoff = 500;
      this.emit({ type: "open" });
    };
    this.ws.onclose = () => {
      this.emit({ type: "close" });
      if (!this.stop) this.scheduleReconnect();
    };
    this.ws.onerror = (e) => {
      this.emit({ type: "error", error: e });
    };
    this.ws.onmessage = (e) => {
      try {
        const msg = JSON.parse(e.data);
        switch (msg?.type) {
          case "state":
            this.emit({ type: "state", payload: msg.payload });
            break;
          case "peer_joined":
          case "peer_left":
            this.emit({ type: msg.type, payload: msg.payload });
            break;
          case "chat": {
            const id: string | undefined = msg?.payload?.msg_id;
            if (id && this.seen.has(id)) return;
            if (id) this.seen.add(id);
            this.emit({ type: "chat", payload: msg.payload });
            break;
          }
          case "chat_ack":
            this.emit({ type: "chat_ack", payload: msg.payload });
            break;
          default:
            // ignore
        }
      } catch {/* ignore */}
    };
  }

  async disconnect() {
    this.stop = true;
    this.seen.clear();
    if (this.ws) {
      try { this.ws.close(); } catch {/* noop */}
      this.ws = null;
    }
  }

  private scheduleReconnect() {
    if (this.reconnecting || this.stop || !this.roomId) return;
    this.reconnecting = true;
    const delay = this.backoff;
    setTimeout(() => {
      this.reconnecting = false;
      if (this.stop || !this.roomId) return;
      this.backoff = Math.min(this.backoff * 2, 10_000);
      this.connect(this.roomId).catch(() => this.scheduleReconnect());
    }, delay);
  }

  sendChat(text: string) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return;
    const payload = { type: "chat", payload: { message: text } };
    this.ws.send(JSON.stringify(payload));
  }
}

export const roomWS = new RoomWS();
