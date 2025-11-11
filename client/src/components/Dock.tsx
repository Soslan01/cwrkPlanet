export function Dock() {
  const items = [
    { label: "Home", href: "#/" },
    { label: "Rooms", href: "#/rooms" },
    { label: "Auth", href: "#/auth" },
    { label: "Profile", href: "#/profile" },
  ];

  return (
    <div className="fixed bottom-4 left-1/2 -translate-x-1/2">
      <div className="flex gap-3 rounded-2xl px-4 py-2 shadow-lg bg-white/80 dark:bg-black/40 backdrop-blur border border-black/5 dark:border-white/10">
        {items.map((it) => (
          <a
            key={it.label}
            href={it.href}
            className="text-xs px-3 py-1 rounded-xl hover:bg-black/5 dark:hover:bg-white/10"
          >
            {it.label}
          </a>
        ))}
      </div>
    </div>
  );
}
