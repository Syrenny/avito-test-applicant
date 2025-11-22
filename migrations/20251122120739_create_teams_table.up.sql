create table teams (
    id        uuid         default uuidv7() primary key,
    team_name varchar(255) not null         unique
);
