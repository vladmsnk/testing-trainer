-- +goose Up
-- +goose StatementBegin

create table if not exists users (
   id serial primary key,
   username varchar(255) not null unique,
   email varchar(255) not null unique,
   password_hash varchar(255) not null,
   created_at timestamp default current_timestamp,
   updated_at timestamp default current_timestamp
);

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
    frequency_type varchar(255) not null, -- тип частоты выполнения привычки
    times_per_frequency int not null, -- количество раз в частоте
    total_tracking_days int, -- общее количество периодов отслеживания
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp,
    is_active bool default false
);

create table if not exists goal_stats (
    id serial primary key,
    goal_id int not null unique,
    total_completed_periods int not null,
    total_completed_times int not null,
    total_skipped_periods int default 0,
    most_longest_streak int not null,
    current_streak int not null
);

create table if not exists goal_logs (
    id serial primary key,
    goal_id int not null,
    record_created_at timestamp default current_timestamp-- время создания записи в этой таблице
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists habits;
drop table if exists goals;
drop table if exists goal_logs;
drop table if exists users;
drop table if exists goal_stats;
-- +goose StatementEnd
