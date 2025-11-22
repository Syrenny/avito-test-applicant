create table teams (
    id        uuid         default uuidv4() primary key,
    team_name varchar(255) not null         unique
);
