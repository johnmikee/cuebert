DROP TRIGGER IF EXISTS update_os_user_time ON updates_os_users;
DROP TRIGGER IF EXISTS update_os_updates_time ON updates_os_updates;
DROP TRIGGER IF EXISTS update_os_tracking_time ON updates_os_tracking;
DROP TRIGGER IF EXISTS update_os_task_time ON updates_os_task;
DROP TRIGGER IF EXISTS update_os_exclusion_time ON updates_os_exclusions;

DROP INDEX IF EXISTS idx_os_user_email;

ALTER TABLE updates_os_users DROP CONSTRAINT IF EXISTS unique_os_user_email;
ALTER TABLE updates_os_tracking DROP CONSTRAINT IF EXISTS updates_os_tracking_pkey;
ALTER TABLE updates_os_task DROP CONSTRAINT IF EXISTS fk_user_service_id;
ALTER TABLE updates_os_users DROP CONSTRAINT IF EXISTS updates_os_users_pkey;
ALTER TABLE updates_os_task DROP CONSTRAINT IF EXISTS updates_os_task_pkey;
ALTER TABLE updates_os_exclusions DROP CONSTRAINT IF EXISTS updates_os_exclusions_pkey;
ALTER TABLE updates_os_updates DROP CONSTRAINT IF EXISTS updates_os_updates_pkey;

DROP TABLE IF EXISTS updates_os_tracking;
DROP TABLE IF EXISTS updates_os_task;
DROP TABLE IF EXISTS updates_os_exclusions;
DROP TABLE IF EXISTS updates_os_users;
DROP TABLE IF EXISTS updates_os_updates;
