-- When modifying this always modify structs in the backend/database as well

create table "User" (
    ID                      uuid            primary key unique not null,
    first_name              varchar(30)     not null,
    last_name               varchar(30)     not null,
    email                   varchar(320)    not null,
    phone_number            varchar(30)     not null,
    password_hash           varchar(72)     not null,
    jwt_refresh_version     integer         not null,
    -- TODO
    -- trial_ended     boolean         not null,
    subscription            integer         not null
);

create table "Merchant" (
    ID                      uuid            primary key unique not null,
    name                    varchar(30)     not null,
    url_name                varchar(30)     unique not null,
    owner_id                uuid            references "User" (ID) not null,
    contact_email           varchar(320)    not null,
    settings                jsonb
);

create table "Appointment" (
    ID                      serial          primary key unique not null,
    client_id               uuid            references "User" (ID) not null,
    merchant_id             uuid            references "Merchant" (ID) not null,
    service_id              integer         references "Service" (ID) not null,
    location_id             integer         references "Location" (ID) not null,
    from_date               timestamptz     not null,
    to_date                 timestamptz     not null
    -- TODO Possible future alternative
    -- time_range      tstzrange       not null
);

create table "Service" (
    ID                      serial          primary key unique not null,
    merchant_id             uuid            references "Merchant" (ID) not null,
    name                    varchar(30)     not null,
    duration                integer         not null,
    price                   bigint          not null,
    blocking                boolean         not null
);

-- TODO
-- create table "Employee" (
--     user_id         uuid            references "User" (ID) not null,
--     merchant_id     uuid            references "Merchant" (ID) not null,
--     location_id     integer         references "Location" (ID) not null
-- );

create table "Location" (
    ID                      serial          primary key unique not null,
    merchant_id             uuid            references "Merchant" (ID) not null,
    country                 varchar(50)     not null,
    city                    varchar(50)     not null,
    postal_code             varchar(10)     not null,
    address                 varchar(100)    not null
    -- TODO
    -- employees       varchar(1)      not null
);

-- create table "Subscription" (
--     ID                      serial          primary key unique not null,
--     name                    varchar(30)     not null,
--     purchase_date           timestamptz     not null,
--     start_date              timestamptz     not null,
--     end_date                timestamptz     not null,
--     price_per_month         bigint          not null
-- );