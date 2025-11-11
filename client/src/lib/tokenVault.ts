const K_ACCESS = "cwrp_access_token";
const K_REFRESH = "cwrp_refresh_token";
const K_USERID = "cwrp_user_id";

function read(k: string): string | null {
  try { return localStorage.getItem(k); } catch { return null; }
}
function write(k: string, v: string | null) {
  try {
    if (v == null) localStorage.removeItem(k);
    else localStorage.setItem(k, v);
  } catch { /* noop */ }
}

export const tokenVault = {
  getAccessToken(): string | null {
    return read(K_ACCESS);
  },
  setAccessToken(v: string | null) {
    write(K_ACCESS, v);
  },

  getRefreshToken(): string | null {
    return read(K_REFRESH);
  },
  setRefreshToken(v: string | null) {
    write(K_REFRESH, v);
  },

  getUserId(): number | null {
    const raw = read(K_USERID);
    if (!raw) return null;
    const n = Number(raw);
    return Number.isFinite(n) ? n : null;
  },
  setUserId(v: number | null) {
    write(K_USERID, v == null ? null : String(v));
  },

  clearAll() {
    write(K_ACCESS, null);
    write(K_REFRESH, null);
    write(K_USERID, null);
  },
};
