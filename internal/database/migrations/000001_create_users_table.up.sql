create table if not exists users (
  id serial primary key,
  google_id varchar(255) unique,
  name varchar(255) not null,
  password varchar(255),
  created_at timestamp default current_timestamp,
  updated_at timestamp default current_timestamp
);

create index idx_users_google_id on users(google_id);