create table users (
    id        uuid         default uuidv4() primary key,
    is_active boolean      default false,
    username  varchar(255) not null,
    team_id   uuid         references teams (
        id
    ) on update cascade on delete restrict
);
