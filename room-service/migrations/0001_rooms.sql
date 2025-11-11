CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Комнаты остаются с UUID
CREATE TABLE IF NOT EXISTS public.rooms (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name             text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 100),
  max_participants smallint NOT NULL DEFAULT 10 CHECK (max_participants >= 1 AND max_participants <= 10),
  created_at       timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_rooms_created_at_desc ON public.rooms (created_at DESC, id);

-- Участники: user_id bigint (под users.id)
CREATE TABLE IF NOT EXISTS public.room_participants (
  room_id   uuid NOT NULL REFERENCES public.rooms(id) ON DELETE CASCADE,
  user_id   bigint NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
  joined_at timestamptz NOT NULL DEFAULT now(),
  last_seen timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (room_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_participants_room      ON public.room_participants (room_id);
CREATE INDEX IF NOT EXISTS idx_participants_last_seen ON public.room_participants (last_seen);
