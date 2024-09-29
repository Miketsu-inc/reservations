create table "User" (
    ID              uuid            primary key unique not null,
    first_name      varchar(30)     not null,
    last_name       varchar(30)     not null,
    email           varchar(320)    not null,
    -- TODO
    phone_number    varchar(1)      not null,
    password_hash   varchar(72)     not null,
    -- TODO
    subscription    varchar(1),
    settings        json
);

create table Merchant (
    ID              uuid            primary key unique not null,
    name            varchar(30)     not null,
    -- TODO
    owner           varchar(1)      not null,
    contact_email   varchar(320)    not null,
    settings        json
)

create table Appointment (
    ID              serial          primary key unique not null,
    -- TODO
    client          varchar(1)      not null,
    -- TODO
    merchant        varchar(1)      not null,
    -- TODO
    location        varchar(1)      not null,
    time_range      tstzrange       not null,
)

create table Appointment_type (
    ID              serial          primary key unique not null,
    -- TODO
    merchant        varchar(1)      not null,
    name            varchar(30)     not null,
    duration        integer         not null,
    price           bigint          not null,
    blocking        boolean         not null,
)

create table Employee (
    -- TODO
    user_id         varchar(1)      not null,
    -- TODO
    merchant_id     varchar(1)      not null,
    -- TODO
    location_id     varchar(1)      not null,
)

create table Location (
    ID              serial          primary key unique not null,
    -- TODO
    merchant        varchar(1)      not null,
    -- TODO
    address         varchar(1)      not null,
    -- TODO
    employees       varchar(1)      not null,
)

create table Subscription (
    ID              serial          primary key unique not null,
    start_date      timestamptz     not null,
    tier            integer         not null,
    price           bigint          not null,
)