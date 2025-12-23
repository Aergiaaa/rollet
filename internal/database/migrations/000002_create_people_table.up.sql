create table if not exists people (
  id serial primary key,
  name varchar(255) not null,
  role varchar(255) not null,
  team integer not null,
  user_id integer references users(id) on delete cascade,
  created_at timestamp default current_timestamp
);

alter table people 
  add column if not exists people_data_id 
  integer references people_data(id)
  on delete cascade;