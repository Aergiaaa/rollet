create table if not exists people (
  id serial primary key,
  name varchar(255) not null,
  role varchar(255) not null,
  team integer not null,
  user_id integer references users(id) on delete cascade,
  created_at timestamp default current_timestamp
);

create index idx_people_user_id on people(user_id);
create index idx_people_team on people(team);