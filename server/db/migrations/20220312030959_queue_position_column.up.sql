-- ALTER TABLE battle_queue
--     ADD COLUMN mech_position INTEGER NULL,
--     ADD CONSTRAINT bq_order_unique UNIQUE (faction_id, mech_position);
--
-- update battle_queue bq1
-- set mech_position = (select SUM(count(*),1) from battle_queue bq2 where bq1.faction_id = bq2.faction_id ORDER BY bq2.queued_at);
--
-- CREATE FUNCTION update_column() RETURNS TRIGGER AS
-- $BODY$
-- DECLARE
--     num INTEGER;
-- BEGIN
--     SELECT SUM(count(*),1) INTO num FROM battle_queue WHERE battle_queue.faction_id = mytable2.faction_id GROUP BY mytable1.field;
--     UPDATE mytable1 SET new_column = num;
-- END;
-- $BODY$
--     LANGUAGE PLPGSQL;
--
-- create trigger check
--     after insert or update on tab
--     for each row execute procedure check();



create view battle_queue as
select t1.*,
       (WITH bqpos AS (
           SELECT t.*, t.mech_id,
                  ROW_NUMBER() OVER(ORDER BY t.queued_at) AS position
           FROM battle_queue t WHERE t.faction_id = t1.faction_id)
        SELECT s.*
        FROM bqpos s
        WHERE s.mech_id = t1.mech_id) as queue_position
from battle_queue t1;