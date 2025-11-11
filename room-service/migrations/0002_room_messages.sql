-- Сообщения в комнатах

CREATE TABLE IF NOT EXISTS public.room_messages (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  room_id    uuid   NOT NULL REFERENCES public.rooms(id) ON DELETE CASCADE,
  user_id    bigint NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
  text       text   NOT NULL CHECK (char_length(text) BETWEEN 1 AND 4000),
  reply_to   uuid       NULL REFERENCES public.room_messages(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_room_messages_room_created_desc
  ON public.room_messages (room_id, created_at DESC, id DESC);

-- Потенциально для фильтра по конкретному пользователю
CREATE INDEX IF NOT EXISTS idx_room_messages_room_user_created_desc
  ON public.room_messages (room_id, user_id, created_at DESC, id DESC);
