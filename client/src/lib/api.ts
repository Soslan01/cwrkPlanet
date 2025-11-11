import { tokenVault } from "./tokenVault";

type ApiOptions = {
  method?: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  headers?: Record<string, string>;
  body?: any;
  signal?: AbortSignal;
  noAuth?: boolean; // не добавлять Authorization/X-User-ID
};

const API_BASE = import.meta.env.VITE_API_BASE?.replace(/\/+$/, "") || "";

let tokenRefresher: null | (() => Promise<boolean>) = null;
export function setTokenRefresher(fn: () => Promise<boolean>) {
  tokenRefresher = fn;
}

export function apiUrl(path: string) {
  if (!path.startsWith("/")) path = "/" + path;
  return `${API_BASE}${path}`;
}

export async function apiFetch<T = unknown>(path: string, opts: ApiOptions = {}): Promise<T> {
  const url = apiUrl(path);
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(opts.headers || {}),
  };

  if (!opts.noAuth) {
    const at = tokenVault.getAccessToken();
    const uid = tokenVault.getUserId();
    if (at) headers["Authorization"] = `Bearer ${at}`;
    if (uid != null) headers["X-User-ID"] = String(uid);
  }

  const init: RequestInit = {
    method: opts.method || (opts.body ? "POST" : "GET"),
    headers,
    body: opts.body != null ? JSON.stringify(opts.body) : undefined,
    signal: opts.signal,
    credentials: "include",
  };

  let res = await fetch(url, init);
  if (res.status === 401 && !opts.noAuth && tokenRefresher) {
    const ok = await tokenRefresher();
    if (ok) {
      const at = tokenVault.getAccessToken();
      const uid = tokenVault.getUserId();
      if (at) headers["Authorization"] = `Bearer ${at}`;
      if (uid != null) headers["X-User-ID"] = String(uid);
      res = await fetch(url, { ...init, headers });
    }
  }

  if (!res.ok) {
    let payload: any = null;
    try { payload = await res.json(); } catch {}
    const err = new Error(payload?.error?.message || payload?.error || `HTTP ${res.status}`);
    (err as any).status = res.status;
    (err as any).payload = payload;
    throw err;
  }

  if (res.status === 204) return undefined as unknown as T;

  // Единая распаковка: если ответ вида { data: ... } → возвращаем именно data
  const json = await res.json();
  return (json && typeof json === "object" && "data" in json ? json.data : json) as T;
}

/* ====== AUTH helpers (с учётом data-обёртки) ====== */
export async function loginOrRegister(
  mode: "login" | "register",
  payload: { email: string; password: string; displayName?: string }
): Promise<{ userId: number | null }> {
  const path = mode === "login" ? "/auth/login" : "/auth/register";
  // apiFetch уже снимет обёртку {data: ...}
  const data: any = await apiFetch(path, { method: "POST", body: payload, noAuth: true });

  const access = data?.accessToken ?? data?.access_token ?? null;
  const refresh = data?.refreshToken ?? data?.refresh_token ?? null;
  if (typeof access === "string") tokenVault.setAccessToken(access);
  if (typeof refresh === "string") tokenVault.setRefreshToken(refresh);

  let userId: number | null = null;
  const u = data?.user ?? {};
  const raw = u?.id ?? u?.user_id ?? u?.userId ?? data?.user_id ?? data?.userId;
  if (raw != null) {
    const n = Number(raw);
    if (Number.isFinite(n)) userId = n;
  }
  if (userId != null) tokenVault.setUserId(userId);

  return { userId };
}

export function logout() {
  tokenVault.clearAll();
}
