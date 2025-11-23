create table users (
    id        uuid         not null primary key,
    is_active boolean      default false,
    username  varchar(255) not null,
    team_id   uuid         references teams (
        id
    ) on update cascade on delete restrict,
    constraint unique_team_username unique (team_id, username)
);
