CREATE TABLE IF NOT EXISTS overdue_users (
    user_id INT GENERATED ALWAYS AS IDENTITY,
    user_service_id INT NOT NULL,
    user_email VARCHAR(255) NOT NULL,
    task VARCHAR(255) NOT NULL
);
