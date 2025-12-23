create table if not exists users (
  id serial primary key,
  email varchar(255) unique not null,
  google_id varchar(255) unique,
  name varchar(255) not null,
  password varchar(255),
  created_at timestamp default current_timestamp,
  updated_at timestamp default current_timestamp
);

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_google_id_key;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_google_id_not_null
  ON users (google_id) WHERE google_id IS NOT NULL;

create index idx_users_google_id on users(google_id);
