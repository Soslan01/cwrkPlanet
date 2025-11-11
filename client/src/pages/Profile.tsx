import { useEffect, useState } from "react";
import { apiFetch, logout } from "../lib/api";
import { tokenVault } from "../lib/tokenVault";

export function Profile() {
  const [user, setUser] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let mounted = true;
    (async () => {
      try {
        const me = await apiFetch<any>("/auth/me", { method: "GET" });
        if (mounted) setUser(me?.user ?? null);
      } catch {
        setError("Не удалось загрузить профиль. Авторизуйтесь.");
      } finally {
        if (mounted) setLoading(false);
      }
    })();
    return () => { mounted = false; };
  }, []);

  const onLogout = () => {
    logout();
    location.hash = "#/auth";
  };

  if (loading) return <p>Загрузка…</p>;
  if (error) return (
    <div className="space-y-3">
      <p className="text-red-500">{error}</p>
      <a href="#/auth" className="underline">Перейти к авторизации</a>
    </div>
  );
  if (!user) return <p>Пользователь не найден</p>;

  const email = user.email;
  const createdAt = user.createdAt;
  const updatedAt = user.updatedAt;

  const avatarURL =
    user?.avatar_url ||
    `https://api.dicebear.com/9.x/identicon/svg?seed=${encodeURIComponent(email || "user")}`;

  return (
    <section className="space-y-6">
      <div className="flex items-center gap-4">
        <img src={avatarURL} alt="avatar" className="h-16 w-16 rounded-2xl border border-black/10 dark:border-white/15" />
        <div>
          <h2 className="text-xl font-semibold">{email}</h2>
          <p className="text-sm opacity-75">
            ID: {user?.id ?? user?.user_id} · Email verified: {String(user?.emailVerified ?? user?.email_verified)}
          </p>
        </div>
        <div className="flex-1" />
        <button
          onClick={onLogout}
          className="text-sm rounded-xl border border-black/10 dark:border-white/15 px-3 py-1 hover:bg-black/5 dark:hover:bg-white/10"
        >
          Выйти
        </button>
      </div>

      <div className="rounded-2xl border border-black/10 dark:border-white/10 p-4 bg-white/60 dark:bg-white/5">
        <h3 className="font-medium mb-2">Профиль</h3>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <Field label="Email" value={email} />
          <Field label="CreatedAt" value={formatTime(createdAt)} />
          <Field label="UpdatedAt" value={formatTime(updatedAt)} />
        </div>
        <p className="text-xs mt-3 opacity-70">
          Аватарка будет подтягиваться из avatar_url, когда бэкенд начнёт возвращать поле.
        </p>
      </div>
    </section>
  );
}

function Field({ label, value }: { label: string; value: any }) {
  return (
    <div className="rounded-xl border border-black/10 dark:border-white/15 p-3">
      <div className="text-xs opacity-60">{label}</div>
      <div className="text-sm break-all">{String(value ?? "—")}</div>
    </div>
  );
}

function formatTime(ts?: number | string) {
  if (!ts && ts !== 0) return "—";
  try {
    const date =
      typeof ts === "number"
        ? new Date(ts * 1000) // UNIX seconds → ms
        : new Date(ts);
    if (isNaN(date.getTime())) return String(ts);
    return date.toLocaleString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return String(ts);
  }
}
