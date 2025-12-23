create table if not exists people_data (
  id serial primary key,
  user_id integer references users(id) on delete cascade,
  created_at timestamp default current_timestamp
);

create index idx_people_data_user_id on people_data(user_id);
create index idx_people_user_id on people(user_id);
create index idx_people_team on people(team);