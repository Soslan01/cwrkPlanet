import { apiFetch } from "./api";

export type RoomItem = {
  id: string;
  name: string;
  max_participants: number;
  created_at: string; // ISO
};

export type RoomsListResponse = {
  items: RoomItem[];
  next_cursor?: string;
};

export type JoinRoomResponse = {
  room_id: string;
  peer_id: string;
};

export type ParticipantItem = {
  user_id: string;
  display_name?: string | null;
  avatar_url?: string | null;
  joined_at: string;  // ISO
  last_seen: string;  // ISO
};

export type ParticipantsResponse = {
  items: ParticipantItem[];
};

export type ChatMessageItem = {
  id: string;
  room_id: string;
  user_id: string;
  text: string;
  created_at: string; // ISO
  reply_to?: string | null;
};

export type ChatHistoryResponse = {
  items: ChatMessageItem[];
  next_cursor?: string;
};

// --- API ---

export async function createRoom(name: string, max?: number) {
  return apiFetch<RoomItem>("/rooms", {
    method: "POST",
    body: { name, max },
  });
}

export async function listRooms(params?: { limit?: number; cursor?: string }) {
  const qs = new URLSearchParams();
  if (params?.limit) qs.set("limit", String(params.limit));
  if (params?.cursor) qs.set("cursor", params.cursor);
  const path = "/rooms" + (qs.toString() ? `?${qs.toString()}` : "");
  return apiFetch<RoomsListResponse>(path);
}

export async function getRoom(roomId: string) {
  return apiFetch<RoomItem>(`/rooms/${encodeURIComponent(roomId)}`);
}

export async function joinRoom(roomId: string) {
  return apiFetch<JoinRoomResponse>(`/rooms/${encodeURIComponent(roomId)}/join`, {
    method: "POST",
  });
}

export async function leaveRoom(roomId: string) {
  return apiFetch<{ status: string }>(`/rooms/${encodeURIComponent(roomId)}/leave`, {
    method: "POST",
  });
}

export async function listParticipants(roomId: string) {
  return apiFetch<ParticipantsResponse>(`/rooms/${encodeURIComponent(roomId)}/participants`);
}

export async function getChatHistory(
  roomId: string,
  args?: { after?: string; limit?: number }
) {
  const qs = new URLSearchParams();
  if (args?.after) qs.set("after", args.after);
  if (args?.limit) qs.set("limit", String(args.limit));
  const path =
    `/rooms/${encodeURIComponent(roomId)}/chat` +
    (qs.toString() ? `?${qs.toString()}` : "");
  return apiFetch<ChatHistoryResponse>(path);
}
