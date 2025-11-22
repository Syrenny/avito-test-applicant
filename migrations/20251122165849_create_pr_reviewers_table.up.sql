create table pr_reviewers (
    pr_id   UUID not null references pull_requests (
        id
    ) on delete cascade,
    user_id UUID not null references users (id) on delete restrict,
    primary key (pr_id, user_id)
);

create index idx_pr_reviewers_user_id on pr_reviewers (user_id);
create index idx_pr_reviewers_pr_id on pr_reviewers (pr_id);
