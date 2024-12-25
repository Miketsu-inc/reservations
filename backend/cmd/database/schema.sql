-- When modifying this always modify structs in the backend/database as well

create table if not exists "User" (
    ID                      uuid            primary key unique not null,
    first_name              varchar(30)     not null,
    last_name               varchar(30)     not null,
    email                   varchar(320)    not null,
    phone_number            varchar(30)     not null,
    password_hash           varchar(72)     not null,
    jwt_refresh_version     integer         not null,
    subscription            integer         not null
);

create table if not exists "Merchant" (
    ID                      uuid            primary key unique not null,
    name                    varchar(30)     not null,
    url_name                varchar(30)     unique not null,
    owner_id                uuid            references "User" (ID) not null,
    contact_email           varchar(320)    not null,
    introduction            varchar(150),
    announcement            varchar(200),
    about_us                text,
    parking_info            text,
    settings                jsonb
);

create table if not exists "Service" (
    ID                      serial          primary key unique not null,
    merchant_id             uuid            references "Merchant" (ID) not null,
    name                    varchar(30)     not null,
    duration                integer         not null,
    price                   bigint          not null,
    blocking                boolean         not null
);

create table if not exists "Location" (
    ID                      serial          primary key unique not null,
    merchant_id             uuid            references "Merchant" (ID) not null,
    country                 varchar(50)     not null,
    city                    varchar(50)     not null,
    postal_code             varchar(10)     not null,
    address                 varchar(100)    not null
);

create table if not exists "Appointment" (
    ID                      serial          primary key unique not null,
    client_id               uuid            references "User" (ID) not null,
    merchant_id             uuid            references "Merchant" (ID) not null,
    service_id              integer         references "Service" (ID) not null,
    location_id             integer         references "Location" (ID) not null,
    from_date               timestamptz     not null,
    to_date                 timestamptz     not null,
    user_comment            text,
    merchant_comment        text
);