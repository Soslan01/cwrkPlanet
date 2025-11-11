import { useEffect, useMemo, useSyncExternalStore } from "react";

function getHashPath(): string {
  const raw = window.location.hash.replace(/^#/, "");
  return raw.startsWith("/") ? raw : "/" + raw;
}

let listeners: Set<() => void> = new Set();

function subscribe(cb: () => void) {
  listeners.add(cb);
  return () => listeners.delete(cb);
}

function notify() {
  for (const cb of listeners) cb();
}

window.addEventListener("hashchange", notify);
window.addEventListener("popstate", notify);

export function navigate(path: string) {
  if (!path.startsWith("#")) path = "#" + path;
  if (window.location.hash === path) return;
  window.location.hash = path;
}

export function useHashRoute() {
  const route = useSyncExternalStore(subscribe, getHashPath, getHashPath);

  const params = useMemo(() => {
    const m = route.match(/^\/rooms\/([^/]+)$/);
    return {
      roomId: m ? decodeURIComponent(m[1]) : null,
    };
  }, [route]);

  return { route, params, navigate };
}
