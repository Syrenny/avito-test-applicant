create table pull_requests (
    id         uuid         default uuidv7() primary key,
    created_at timestamptz  not null         default now(),
    merged_at  timestamptz,
    pr_name    varchar(255) not null,
    pr_status  smallint     not null,
    author_id  uuid         not null         references users (
        id
    ) on delete restrict
);
