BEGIN;

delete from mech_stats;

alter table mech_stats
    add column battles_survived int not null default 0,
    add column total_losses int not null default 0;

insert into mech_stats(mech_id, total_kills, total_deaths, battles_survived, total_wins, total_losses)
select m.id, 0, 0, 0 ,0 ,0 from mechs m;

-- we run the blow queries manually
-- WITH stats AS (select
--                    m.id,
--                    coalesce((select count(_bk.mech_id) as deaths from battle_kills _bk where _bk.mech_id = m.id group by _bk.mech_id), 0) as kills
--                from mechs m
--                         left join battle_kills bk on bk.mech_id = m.id
--                group by m.id)
-- update mech_stats ms set total_kills = stats.kills from stats where ms.mech_id = stats.id;
--
--
-- WITH stats AS (select
--                    m.id,
--                    coalesce((select count(_bh.war_machine_one_id) as deaths from battle_history _bh where _bh.war_machine_one_id = m.id group by _bh.war_machine_one_id), 0) as deaths
--                from mechs m
--                group by m.id)
-- update mech_stats ms set total_deaths = stats.deaths from stats where ms.mech_id = stats.id;
--
--
-- WITH stats AS (select
--                    m.id,
--                    coalesce((select count(_bw.mech_id) from battle_wins _bw where _bw.mech_id = m.id group by _bw.mech_id), 0) as battles_survived
--                from mechs m
--                         left join battle_wins bw on bw.mech_id = m.id
--                group by m.id)
-- update mech_stats ms set battles_survived = stats.battles_survived from stats where ms.mech_id = stats.id;
--
-- WITH stats AS (select
--                    m.id,
--                    coalesce((select count(_bm.mech_id) from battle_mechs _bm where _bm.mech_id = m.id and _bm.faction_won group by _bm.mech_id), 0) faction_wins
--                from mechs m
--                group by m.id)
-- update mech_stats ms set total_wins = stats.faction_wins from stats where ms.mech_id = stats.id;
--
-- WITH stats AS (select
--                    m.id,
--                    coalesce((select count(_bm.mech_id) from battle_mechs _bm where _bm.mech_id = m.id and not _bm.faction_won group by _bm.mech_id), 0) faction_losses
--                from mechs m
--                group by m.id)
-- update mech_stats ms set total_losses = stats.faction_losses from stats where ms.mech_id = stats.id;

COMMIT;
