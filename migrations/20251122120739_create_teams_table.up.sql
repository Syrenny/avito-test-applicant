create table teams (
    id        uuid         not null primary key,
    team_name varchar(255) not null unique
);
