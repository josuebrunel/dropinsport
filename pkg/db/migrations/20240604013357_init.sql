-- +goose Up
-- +goose StatementBegin
create table if not exists groups (
    uuid uuid primary key,
    name varchar(64) not null,
    sport varchar(64) not null,
    description text,
    street varchar(256),
    city varchar(64),
    country varchar(64),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
) ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists groups;
-- +goose StatementEnd