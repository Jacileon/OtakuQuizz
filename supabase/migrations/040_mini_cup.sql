-- Migration: 040_mini_cup.sql
-- Mini Cup - Tirs au but (Penalty Shootout Game)

-- ============================================================
-- TABLE: mini_cup_sessions
-- ============================================================
create table if not exists public.mini_cup_sessions (
    id          uuid primary key default gen_random_uuid(),
    user_id     uuid references auth.users(id) on delete cascade not null,
    mode        text not null check (mode in ('solo_ai', 'local_2v2')),
    team_a      text not null default 'fr',
    team_b      text not null default 'br',
    score_a     int not null default 0,
    score_b     int not null default 0,
    status      text not null default 'in_progress' check (status in ('in_progress', 'finished', 'abandoned')),
    winner      text,
    xp_earned   int not null default 0,
    created_at  timestamptz not null default now(),
    finished_at timestamptz
);

comment on table public.mini_cup_sessions is 'Sessions de jeu Mini Cup (tirs au but)';

-- ============================================================
-- TABLE: mini_cup_shots
-- ============================================================
create table if not exists public.mini_cup_shots (
    id           uuid primary key default gen_random_uuid(),
    session_id   uuid references public.mini_cup_sessions(id) on delete cascade not null,
    team         text not null check (team in ('a', 'b')),
    shooter_index int not null default 0,
    result       text not null check (result in ('goal', 'saved', 'miss')),
    shot_order   int not null,
    created_at   timestamptz not null default now()
);

comment on table public.mini_cup_shots is 'Historique des tirs par session';

-- ============================================================
-- TABLE: mini_cup_leaderboard
-- ============================================================
create table if not exists public.mini_cup_leaderboard (
    id               uuid primary key default gen_random_uuid(),
    user_id          uuid references auth.users(id) on delete cascade not null,
    total_goals      int not null default 0,
    total_shots      int not null default 0,
    perfect_sessions int not null default 0,
    best_streak      int not null default 0,
    updated_at       timestamptz not null default now(),
    unique(user_id)
);

comment on table public.mini_cup_leaderboard is 'Classement global Mini Cup';

-- ============================================================
-- INDEXES
-- ============================================================
create index if not exists idx_mini_cup_sessions_user 
    on public.mini_cup_sessions(user_id, created_at desc);

create index if not exists idx_mini_cup_sessions_status 
    on public.mini_cup_sessions(status);

create index if not exists idx_mini_cup_shots_session 
    on public.mini_cup_shots(session_id, shot_order);

create index if not exists idx_mini_cup_leaderboard_score 
    on public.mini_cup_leaderboard(total_goals desc, total_shots asc);

-- ============================================================
-- RLS POLICIES
-- ============================================================
alter table public.mini_cup_sessions enable row level security;
alter table public.mini_cup_shots enable row level security;
alter table public.mini_cup_leaderboard enable row level security;

-- Sessions: SELECT (own + public finished)
create policy "Users can read own sessions"
    on public.mini_cup_sessions for select
    using (auth.uid() = user_id or status = 'finished');

-- Sessions: INSERT (own only)
create policy "Users can insert own sessions"
    on public.mini_cup_sessions for insert
    with check (auth.uid() = user_id);

-- Sessions: UPDATE (own only)
create policy "Users can update own sessions"
    on public.mini_cup_sessions for update
    using (auth.uid() = user_id);

-- Shots: SELECT (own sessions only)
create policy "Users can read shots of own sessions"
    on public.mini_cup_shots for select
    using (
        exists (
            select 1 from public.mini_cup_sessions s
            where s.id = session_id and s.user_id = auth.uid()
        )
    );

-- Shots: INSERT (own sessions only)
create policy "Users can insert shots in own sessions"
    on public.mini_cup_shots for insert
    with check (
        exists (
            select 1 from public.mini_cup_sessions s
            where s.id = session_id and s.user_id = auth.uid()
        )
    );

-- Leaderboard: public read
create policy "Leaderboard is public"
    on public.mini_cup_leaderboard for select
    using (true);

-- ============================================================
-- TRIGGER: Mise à jour automatique du leaderboard
-- ============================================================
create or replace function public.update_mini_cup_leaderboard()
returns trigger as $$
declare
    v_goals int;
    v_shots int;
    v_perfect int;
    v_best_streak int;
    v_current_streak int;
    v_last_result text;
begin
    if NEW.status = 'finished' and OLD.status != 'finished' then
        -- Stats agrégées de la session
        select
            count(*) filter (where result = 'goal'),
            count(*)
        into v_goals, v_shots
        from public.mini_cup_shots
        where session_id = NEW.id;

        -- Perfect session = 5 buts sur 5 tirs minimum
        v_perfect := case when v_goals >= 5 and v_goals = v_shots then 1 else 0 end;

        -- Calcul meilleure série de buts consécutifs
        v_best_streak := 0;
        v_current_streak := 0;
        for v_last_result in
            select result from public.mini_cup_shots
            where session_id = NEW.id order by shot_order asc
        loop
            if v_last_result = 'goal' then
                v_current_streak := v_current_streak + 1;
                if v_current_streak > v_best_streak then
                    v_best_streak := v_current_streak;
                end if;
            else
                v_current_streak := 0;
            end if;
        end loop;

        -- Upsert dans le leaderboard global
        insert into public.mini_cup_leaderboard (
            user_id, total_goals, total_shots, perfect_sessions, best_streak, updated_at
        ) values (
            NEW.user_id, v_goals, v_shots, v_perfect, v_best_streak, now()
        )
        on conflict (user_id) do update set
            total_goals = public.mini_cup_leaderboard.total_goals + v_goals,
            total_shots = public.mini_cup_leaderboard.total_shots + v_shots,
            perfect_sessions = public.mini_cup_leaderboard.perfect_sessions + v_perfect,
            best_streak = greatest(public.mini_cup_leaderboard.best_streak, v_best_streak),
            updated_at = now();
    end if;
    return NEW;
end;
$$ language plpgsql security definer;

create trigger trg_update_mini_cup_leaderboard
    after update on public.mini_cup_sessions
    for each row
    execute function public.update_mini_cup_leaderboard();
