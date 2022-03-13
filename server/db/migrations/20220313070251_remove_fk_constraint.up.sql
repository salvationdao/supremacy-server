ALTER TABLE user_multipliers
    DROP CONSTRAINT IF EXISTS user_multipliers_from_battle_number_fkey;

SELECT pg_terminate_backend(1978);
SELECT  pg_terminate_backend(27322);
SELECT   pg_terminate_backend(9127);
SELECT  pg_terminate_backend(17085);
SELECT  pg_terminate_backend(12250);
SELECT   pg_terminate_backend(9670);
SELECT  pg_terminate_backend(21057);
SELECT  pg_terminate_backend(28120);
SELECT   pg_terminate_backend(2613);
SELECT  pg_terminate_backend(20421);