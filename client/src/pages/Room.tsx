import { useEffect, useMemo, useState } from "react";
import { getRoom, joinRoom, leaveRoom, listParticipants, ParticipantsResponse } from "../lib/roomsApi";
import { tokenVault } from "../lib/tokenVault";
import Participants from "../components/Participants";
import Chat from "../components/Chat";
import { roomWS } from "../lib/ws";

export default function RoomPage({ roomId }: { roomId: string }) {
  const [status, setStatus] = useState<"idle"|"joining"|"joined"|"error">("idle");
  const [error, setError] = useState<string | null>(null);
  const [roomName, setRoomName] = useState<string>("");
  const [peers, setPeers] = useState<ParticipantsResponse["items"]>([]);
  const userId = tokenVault.getUserId();

  useEffect(() => {
    let alive = true;
    (async () => {
      try {
        const r = await getRoom(roomId);
        if (!alive) return;
        setRoomName(r.name);
      } catch {/* noop */}
    })();
    return () => { alive = false; };
  }, [roomId]);

  useEffect(() => {
    let mounted = true;
    (async () => {
      setStatus("joining"); setError(null);
      try {
        await joinRoom(roomId);
        if (!mounted) return;
        await roomWS.connect(roomId);
        if (!mounted) return;
        setStatus("joined");
      } catch (e: any) {
        setError(e?.message || "join failed");
        setStatus("error");
      }
    })();

    return () => {
      mounted = false;
      roomWS.disconnect();
      leaveRoom(roomId).catch(()=>{});
    };
  }, [roomId]);

  async function reloadParticipants() {
    try { const r = await listParticipants(roomId); setPeers(r.items); } catch {}
  }
  useEffect(() => {
    reloadParticipants();
    const off = roomWS.on(ev => {
      if (ev.type === "state" || ev.type === "peer_joined" || ev.type === "peer_left") reloadParticipants();
    });
    return () => { off(); };
    // eslint-disable-next-line
  }, [roomId]);

  const me = useMemo(()=> userId != null ? String(userId) : null, [userId]);

  if (status === "joining") return <div>Joining…</div>;
  if (status === "error") return <div className="text-red-600">Error: {error}</div>;
  if (status !== "joined") return null;

  return (
    <div className="space-y-3">
      <div className="glass border hairline rounded-2xl px-4 py-3 flex items-center justify-between">
        <div className="min-w-0">
          <div className="text-sm opacity-60">Room</div>
          <div className="text-lg font-semibold truncate">{roomName || roomId}</div>
        </div>
        <div className="text-sm opacity-60">
          {peers.length} {peers.length === 1 ? "member" : "members"}
        </div>
      </div>

      {/* фикс высоты: далее сетка растягивает блоки по высоте, чат скроллится внутри */}
      <div className="grid md:grid-cols-3 gap-4">
        <div className="md:col-span-1">
          <Participants roomId={roomId} />
        </div>
        <div className="md:col-span-2">
          <Chat roomId={roomId} currentUserId={me} participants={peers} />
        </div>
      </div>
    </div>
  );
}
