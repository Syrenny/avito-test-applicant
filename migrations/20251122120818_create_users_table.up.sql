create table users (
    id        uuid         default uuidv7() primary key,
    is_active boolean      not null,
    username  varchar(255) not null,
    team_id   uuid         references teams (
        id
    ) on update cascade on delete restrict
);
