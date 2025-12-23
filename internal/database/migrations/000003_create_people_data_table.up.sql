create table if not exists people_data (
  id serial primary key,
  user_id integer references users(id) on delete cascade,
  created_at timestamp default current_timestamp
);

alter table people
  add column if not exists people_data_id 
  integer references people_data(id) on delete cascade;

create index idx_people_data_person_id on people_data(person_id);
create index idx_people_data_user_id on people_data(user_id);
create index idx_people_user_id on people(user_id);
create index idx_people_team on people(team);