import { useHashRoute } from "../lib/router";
import { useTheme } from "../lib/theme";

export function TopBar() {
  const { theme, setTheme } = useTheme();
  const { navigate } = useHashRoute();

  return (
    <header className="fixed inset-x-0 top-0 backdrop-blur supports-[backdrop-filter]:bg-white/60 dark:supports-[backdrop-filter]:bg-black/30 border-b hairline z-40">
      <div className="mx-auto max-w-3xl px-4 h-12 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="h-3 w-3 rounded-full bg-red-400" />
          <div className="h-3 w-3 rounded-full bg-yellow-400" />
          <div className="h-3 w-3 rounded-full bg-green-400" />
          <button onClick={() => navigate("/")} className="ml-3 text-sm font-medium opacity-80 hover:opacity-100">
            CWRK Planet
          </button>
        </div>
        <div className="flex items-center gap-2">
          <a href="#/rooms" className="text-sm opacity-80 hover:opacity-100">Rooms</a>
          <a href="#/auth" className="text-sm opacity-80 hover:opacity-100">Sign in</a>
          <a href="#/profile" className="text-sm opacity-80 hover:opacity-100">Profile</a>
          <div className="w-px h-5 bg-black/10 dark:bg-white/10 mx-2" />
          <button
            onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
            className="text-xs rounded-lg px-2 py-1 border hairline hover:bg-black/5 dark:hover:bg-white/5"
            aria-label="Toggle theme"
          >
            {theme === "dark" ? "Light" : "Dark"}
          </button>
        </div>
      </div>
    </header>
  );
}
