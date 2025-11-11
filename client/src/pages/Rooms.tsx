import { useEffect, useState } from "react";
import { listRooms, createRoom, RoomsListResponse } from "../lib/roomsApi";
import { navigate } from "../lib/router";

export default function RoomsPage() {
  const [items, setItems] = useState<RoomsListResponse["items"]>([]);
  const [cursor, setCursor] = useState<string | undefined>();
  const [loading, setLoading] = useState(false);
  const [name, setName] = useState("");
  const [max, setMax] = useState<number>(10);
  const [error, setError] = useState<string | null>(null);

  async function load(first = false) {
    setLoading(true); setError(null);
    try {
      const res = await listRooms(first ? undefined : { cursor });
      setItems(first ? res.items : [...items, ...res.items]);
      setCursor(res.next_cursor);
    } catch (e: any) {
      setError(e?.message || "load failed");
    } finally { setLoading(false); }
  }

  useEffect(() => { load(true); /* eslint-disable-next-line */ }, []);

  async function onCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim()) return;
    try {
      const room = await createRoom(name.trim(), max);
      navigate(`/rooms/${room.id}`);
    } catch (e: any) {
      setError(e?.message || "create failed");
    }
  }

  return (
    <div className="space-y-8">
      {/* hero */}
      <div className="text-center mt-2">
        <h1 className="text-[28px] font-semibold tracking-[-0.02em]">Rooms</h1>
        <p className="text-sm opacity-70 mt-1">Create a room or join an existing one</p>
      </div>

      {/* create form — macOS minimal */}
      <form onSubmit={onCreate} className="glass border hairline rounded-2xl p-4 flex flex-wrap gap-3 items-end">
        <div className="flex flex-col min-w-[220px]">
          <label className="text-[12px] opacity-70 mb-1">Name</label>
          <input
            className="px-3 py-2 rounded-xl hairline bg-white/70 dark:bg-white/5 outline-none focus:ring-2 focus:ring-black/10 dark:focus:ring-white/20"
            value={name}
            onChange={e=>setName(e.target.value)}
            placeholder="My room"
          />
        </div>
        <div className="flex flex-col">
          <label className="text-[12px] opacity-70 mb-1">Max</label>
          <input
            className="px-3 py-2 rounded-xl hairline bg-white/70 dark:bg-white/5 w-24 outline-none focus:ring-2 focus:ring-black/10 dark:focus:ring-white/20"
            type="number" min={2} max={50}
            value={max}
            onChange={e=>setMax(Number(e.target.value||10))}
          />
        </div>
        <button className="px-4 py-2 rounded-xl bg-black text-white dark:bg-white dark:text-black">
          Create
        </button>
        {error && <div className="text-red-600 text-sm ml-auto">{error}</div>}
      </form>

      {/* list */}
      <div className="space-y-3">
        {items.length === 0 && !loading ? (
          <div className="glass border hairline rounded-2xl p-6 text-center opacity-80">
            No rooms yet
          </div>
        ) : (
          <ul className="space-y-2">
            {items.map(it => (
              <li
                key={it.id}
                className="glass border hairline rounded-2xl px-4 py-3 hover:shadow-md transition-shadow flex items-center justify-between"
              >
                <div className="min-w-0">
                  <div className="font-medium truncate">{it.name}</div>
                  <div className="text-[12px] opacity-70">max: {it.max_participants}</div>
                </div>
                <button
                  onClick={()=>navigate(`/rooms/${it.id}`)}
                  className="px-3 py-1.5 rounded-xl border hairline"
                >
                  Enter
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>

      <div className="flex gap-2">
        <button
          disabled={!cursor || loading}
          onClick={()=>load(false)}
          className="px-3 py-1.5 rounded-xl border hairline disabled:opacity-40"
        >
          Load more
        </button>
        <button
          disabled={loading}
          onClick={()=>load(true)}
          className="px-3 py-1.5 rounded-xl border hairline"
        >
          Refresh
        </button>
      </div>
    </div>
  );
}
