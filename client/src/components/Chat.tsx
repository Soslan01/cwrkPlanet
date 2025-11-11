import { useEffect, useMemo, useRef, useState } from "react";
import { getChatHistory, ChatMessageItem, ParticipantItem } from "../lib/roomsApi";
import { roomWS } from "../lib/ws";

type Props = {
  roomId: string;
  currentUserId?: string | null;
  participants?: ParticipantItem[]; // для имён/аватаров
};

function lookup(participants: ParticipantItem[] | undefined, uid: string) {
  if (!participants) return { name: `User ${uid}`, avatar: null as string | null };
  const p = participants.find(x => x.user_id === uid);
  return {
    name: p?.display_name || `User ${uid}`,
    avatar: p?.avatar_url || `https://api.dicebear.com/7.x/initials/svg?seed=${encodeURIComponent(p?.display_name || uid)}`,
  };
}

const GROUP_WINDOW_MIN = 5;
function isSameGroup(prev: ChatMessageItem | null, cur: ChatMessageItem) {
  if (!prev) return false;
  if (prev.user_id !== cur.user_id) return false;
  const dt = Math.abs(new Date(cur.created_at).getTime() - new Date(prev.created_at).getTime());
  return dt <= GROUP_WINDOW_MIN * 60 * 1000;
}

export default function Chat({ roomId, currentUserId, participants }: Props) {
  const [items, setItems] = useState<ChatMessageItem[]>([]);
  const [cursor, setCursor] = useState<string | undefined>();
  const [loading, setLoading] = useState(false);
  const [text, setText] = useState("");
  const seen = useRef<Set<string>>(new Set());

  async function load(first=false) {
    setLoading(true);
    try {
      const res = await getChatHistory(roomId, first ? undefined : { after: cursor });
      const merged = first ? res.items : [...items, ...res.items];
      const uniq: ChatMessageItem[] = [];
      const seenIds = new Set<string>();
      for (const m of merged) {
        if (seenIds.has(m.id)) continue;
        seenIds.add(m.id);
        uniq.push(m);
      }
      setItems(uniq.sort((a,b)=> new Date(a.created_at).getTime()-new Date(b.created_at).getTime()));
      setCursor(res.next_cursor);
    } finally { setLoading(false); }
  }

  useEffect(()=>{ load(true); /* eslint-disable-next-line */ }, [roomId]);

  useEffect(() => {
    const off = roomWS.on(ev => {
      if (ev.type === "chat") {
        const p = ev.payload;
        if (p?.msg_id && seen.current.has(p.msg_id)) return;
        if (p?.msg_id) seen.current.add(p.msg_id);
        const msg: ChatMessageItem = {
          id: p.msg_id || `${Date.now()}-${Math.random()}`,
          room_id: p.room_id,
          user_id: p.user_id,
          text: p.message,
          created_at: new Date((p.ts_unix || Math.floor(Date.now()/1000))*1000).toISOString(),
        };
        setItems(prev => [...prev, msg]);
      }
    });
    return () => { off(); };
  }, []);

  function onSend(e: React.FormEvent) {
    e.preventDefault();
    const t = text.trim();
    if (!t) return;
    roomWS.sendChat(t);
    setText("");
  }

  const tailRef = useRef<HTMLDivElement | null>(null);
  useEffect(()=>{ tailRef.current?.scrollIntoView({ behavior: "smooth" }); }, [items.length]);

  const me = useMemo(()=> String(currentUserId ?? ""), [currentUserId]);

  return (
    <div className="glass border hairline rounded-2xl h-full chat-panel flex flex-col">
      <div className="px-4 py-2 border-b hairline flex items-center justify-between">
        <div className="font-medium">Chat</div>
        <button className="text-sm opacity-70" disabled={loading || !cursor} onClick={()=>load(false)}>
          Load older
        </button>
      </div>

      <div className="flex-1 overflow-auto px-3 py-3 mac-scroll space-y-2">
        {items.map((m, idx) => {
          const mine = m.user_id === me;
          const prev = idx > 0 ? items[idx - 1] : null;
          const firstOfGroup = !isSameGroup(prev, m);
          const meta = lookup(participants, m.user_id);

          if (mine) {
            return (
              <div key={m.id} className="chat-appear w-full flex justify-end">
                <div className="max-w-[76%] px-3 py-2 leading-snug bg-black text-white dark:bg-white dark:text-black rounded-2xl rounded-br-sm">
                  {/* корректные переносы внутри пузыря */}
                  <div className="whitespace-pre-wrap break-words [hyphens:auto]">
                    {m.text}
                  </div>
                  <div className="text-[11px] mt-1 opacity-80">
                    {new Date(m.created_at).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
                  </div>
                </div>
              </div>
            );
          }

          return (
            <div key={m.id} className="chat-appear w-full flex justify-start">
              <div className="w-full flex flex-col items-start gap-1">
                {firstOfGroup && (
                  <img
                    src={meta.avatar || undefined}
                    alt=""
                    className="w-8 h-8 rounded-full border hairline"
                  />
                )}
                <div className="max-w-[76%] px-3 py-2 leading-snug bg-white/90 dark:bg-white/10 border hairline rounded-2xl rounded-bl-sm">
                  {firstOfGroup && (
                    <div className="text-[12px] opacity-70 mb-0.5 whitespace-pre-wrap break-words [hyphens:auto]">
                      {meta.name}
                    </div>
                  )}
                  {/* переносы «по словам» */}
                  <div className="whitespace-pre-wrap break-words [hyphens:auto]">
                    {m.text}
                  </div>
                  <div className="text-[11px] mt-1 opacity-60">
                    {new Date(m.created_at).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
                  </div>
                </div>
              </div>
            </div>
          );
        })}
        <div ref={tailRef} />
      </div>

      <form onSubmit={onSend} className="p-2 border-t hairline flex gap-2">
        <input
          className="flex-1 px-3 py-2 rounded-xl hairline bg-white/70 dark:bg-white/5 outline-none focus:ring-2 focus:ring-black/10 dark:focus:ring-white/20"
          value={text}
          onChange={e=>setText(e.target.value)}
          placeholder="Message…"
        />
        <button className="px-3 py-2 rounded-xl bg-black text-white dark:bg-white dark:text-black">Send</button>
      </form>
    </div>
  );
}
