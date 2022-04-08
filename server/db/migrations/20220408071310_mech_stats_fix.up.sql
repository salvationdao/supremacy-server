BEGIN;

delete from mech_stats;

alter table mech_stats
    add column battles_survived int not null default 0,
    add column total_losses int not null default 0;

insert into mech_stats(mech_id, total_kills, total_deaths, battles_survived, total_wins, total_losses)
    select
        m.id,
        coalesce((select count(_bk.mech_id) as deaths from battle_kills _bk where _bk.mech_id = m.id group by _bk.mech_id), 0) as kills,
        coalesce((select count(_bh.war_machine_one_id) as deaths from battle_history _bh where _bh.war_machine_one_id = m.id group by _bh.war_machine_one_id), 0) as deaths,
        coalesce((select count(_bw.mech_id) from battle_wins _bw where _bw.mech_id = m.id group by _bw.mech_id), 0) as battles_survived,
        coalesce((select count(_bm.mech_id) from battle_mechs _bm where _bm.mech_id = m.id and _bm.faction_won group by _bm.mech_id), 0) faction_wins,
        coalesce((select count(_bm.mech_id) from battle_mechs _bm where _bm.mech_id = m.id and not _bm.faction_won group by _bm.mech_id), 0) faction_losses,
    from mechs m
    left join battle_kills bk on bk.mech_id = m.id
    left join battle_wins bw on bw.mech_id = m.id
    group by m.id
    order by kills desc;

COMMIT;
