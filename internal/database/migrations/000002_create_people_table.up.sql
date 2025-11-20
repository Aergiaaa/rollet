create table if not exists people (
  id serial primary key,
  name varchar(255) not null,
  role varchar(255) not null,
  team integer not null,
  created_at timestamp default current_timestamp
);