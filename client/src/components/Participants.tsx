import { useEffect, useState } from "react";
import { ParticipantsResponse, listParticipants } from "../lib/roomsApi";
import { roomWS } from "../lib/ws";

const ONLINE_WINDOW_SEC = 30; // считаем online, если last_seen свежее 30с

export default function Participants({ roomId }: { roomId: string }) {
  const [items, setItems] = useState<ParticipantsResponse["items"]>([]);

  async function reload() {
    try {
      const r = await listParticipants(roomId);
      setItems(r.items);
    } catch {/* noop */}
  }

  useEffect(() => {
    reload();
    const off = roomWS.on(ev => {
      if (ev.type === "state" || ev.type === "peer_joined" || ev.type === "peer_left") {
        reload();
      }
    });
    return () => { off(); };
    // eslint-disable-next-line
  }, [roomId]);

  const now = Date.now();

  return (
    <div className="glass border hairline rounded-2xl h-full flex flex-col">
      <div className="px-3 py-2 border-b hairline font-medium">Participants</div>
      <ul className="flex-1 overflow-auto divide-y divide-black/5 dark:divide-white/10 mac-scroll">
        {items.map(p => {
          const last = new Date(p.last_seen).getTime();
          const online = (now - last) <= ONLINE_WINDOW_SEC * 1000;
          return (
            <li key={p.user_id} className="p-2 flex items-center gap-3">
              <div className="relative">
                <img
                  src={p.avatar_url || "https://api.dicebear.com/7.x/initials/svg?seed=" + encodeURIComponent(p.display_name || p.user_id)}
                  alt=""
                  className="w-8 h-8 rounded-full border hairline"
                />
                <span className={`absolute -right-0.5 -bottom-0.5 w-3 h-3 rounded-full border
                    ${online ? "bg-emerald-500 border-white dark:border-neutral-900" : "bg-gray-300 border-white dark:border-neutral-900"}`} />
              </div>
              <div className="min-w-0">
                <div className="truncate font-medium">{p.display_name || `User ${p.user_id}`}</div>
                <div className="text-[12px] opacity-60">id: {p.user_id}</div>
              </div>
            </li>
          );
        })}
        {items.length === 0 && <li className="p-2 opacity-70">No participants</li>}
      </ul>
    </div>
  );
}
