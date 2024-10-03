-- +goose Up
-- +goose StatementBegin
create table if not exists habits (
    id serial primary key,
    username varchar(255) not null,
    name varchar(255) not null, -- имя привычки
    description text, -- описание привычки
    is_archived bool default false, -- архивирована ли привычка
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp
);

create table if not exists goals (
    id serial primary key,
    habit_id int not null,
    frequency int not null, -- сколько раз нужно выполнить привычку за указанный период
    duration bigint not null, -- период, за который нужно выполнить привычку frequency раз
    num_of_periods int not null, -- количество таких периодов, за которые нужно выполнить привычку для достижеия цели
    start_tracking_at timestamp default current_timestamp, -- начало отслеживания привычки
    end_tracking_at timestamp, -- конец отслеживания привычки, start_tracking_at + duration * num_of_periods
    created_at timestamp default current_timestamp,
    is_active bool default false -- активна ли цель
);

create table if not exists habit_logs (
    id serial primary key,
    goal_id int not null,
    record_created_at timestamp default current_timestamp, -- время создания записи в этой таблице
    execution_time timestamp not null -- время выполнения привычки
);

create table users (
    id serial primary key,
    username varchar(255) not null unique,
    email varchar(255) not null unique,
    password_hash varchar(255) not null,
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists habits;
drop table if exists goals;
drop table if exists habit_logs;
-- +goose StatementEnd
