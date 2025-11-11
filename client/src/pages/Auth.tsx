// Общая форма: режимы "login" и "register". После успеха — редирект на #/profile.
import { useState } from "react";
import { loginOrRegister } from "../lib/api";

export function Auth() {
  const [mode, setMode] = useState<"login" | "register">("login");
  const [email, setEmail] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [info, setInfo] = useState("");

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true); setError(""); setInfo("");
    try {
      const payload = mode === "register"
        ? { email, displayName, password }
        : { email, password };
      await loginOrRegister(mode, payload);
      setInfo("Успешно! Перехожу в профиль…");
      location.hash = "#/profile";
    } catch (err: any) {
      setError(typeof err?.message === "string" ? err.message : "Ошибка");
    } finally {
      setLoading(false);
    }
  }

  return (
    <section className="max-w-md mx-auto">
      <div className="rounded-2xl border border-black/10 dark:border-white/10 p-6 bg-white/60 dark:bg-white/5">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold">{mode === "login" ? "Вход" : "Регистрация"}</h2>
          <button
            className="text-xs underline opacity-80 hover:opacity-100"
            onClick={() => setMode(mode === "login" ? "register" : "login")}
          >
            {mode === "login" ? "Нужен аккаунт?" : "Уже есть аккаунт?"}
          </button>
        </div>

        <form onSubmit={onSubmit} className="space-y-3">
          <label className="block">
            <span className="text-sm">Email</span>
            <input
              type="email" value={email} onChange={e => setEmail(e.target.value)} required
              className="mt-1 w-full rounded-xl border border-black/10 dark:border-white/15 bg-transparent px-3 py-2 outline-none focus:ring-2 focus:ring-black/10 dark:focus:ring-white/20"
            />
          </label>

          {mode === "register" && (
            <label className="block">
              <span className="text-sm">Display name</span>
              <input
                value={displayName} onChange={e => setDisplayName(e.target.value)} required
                className="mt-1 w-full rounded-xl border border-black/10 dark:border-white/15 bg-transparent px-3 py-2 outline-none focus:ring-2 focus:ring-black/10 dark:focus:ring-white/20"
              />
            </label>
          )}

          <label className="block">
            <span className="text-sm">Пароль</span>
            <input
              type="password" value={password} onChange={e => setPassword(e.target.value)} required
              className="mt-1 w-full rounded-xl border border-black/10 dark:border-white/15 bg-transparent px-3 py-2 outline-none focus:ring-2 focus:ring-black/10 dark:focus:ring-white/20"
            />
          </label>

          {error && <p className="text-sm text-red-500 whitespace-pre-wrap">{error}</p>}
          {info && <p className="text-sm text-emerald-600">{info}</p>}

          <button
            disabled={loading}
            className="w-full rounded-xl border border-black/10 dark:border-white/15 px-4 py-2 text-sm hover:bg-black/5 dark:hover:bg-white/10 disabled:opacity-50"
          >
            {loading ? "Обработка…" : (mode === "login" ? "Войти" : "Зарегистрироваться")}
          </button>
        </form>
      </div>
    </section>
  );
}
