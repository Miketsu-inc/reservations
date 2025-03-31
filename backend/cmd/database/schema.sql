-- When modifying this always modify structs in the backend/database as well

create table if not exists "User" (
    ID                       uuid            primary key unique not null,
    first_name               varchar(30)     not null,
    last_name                varchar(30)     not null,
    email                    varchar(320),
    phone_number             varchar(30),
    password_hash            varchar(72),
    jwt_refresh_version      integer,
    subscription             integer,
    is_dummy                 boolean         not null,
    added_by                 uuid
);

create table if not exists "Merchant" (
    ID                       uuid            primary key unique not null,
    name                     varchar(30)     not null,
    url_name                 varchar(30)     unique not null,
    owner_id                 uuid            references "User" (ID) not null,
    contact_email            varchar(320)    not null,
    introduction             varchar(150),
    announcement             varchar(200),
    about_us                 text,
    parking_info             text,
    payment_info             text
);

create table if not exists "Service" (
    ID                       serial          primary key unique not null,
    merchant_id              uuid            references "Merchant" (ID) not null,
    name                     varchar(30)     not null,
    description              varchar(200),
    color                    char(7)         not null,
    duration                 integer         not null,
    price                    bigint          not null,
    cost                     bigint          not null,
    deleted_on               timestamptz
);

create table if not exists "Location" (
    ID                       serial          primary key unique not null,
    merchant_id              uuid            references "Merchant" (ID) not null,
    country                  varchar(50)     not null,
    city                     varchar(50)     not null,
    postal_code              varchar(10)     not null,
    address                  varchar(100)    not null
);

create table if not exists "Appointment" (
    ID                       serial          primary key unique not null,
    user_id                  uuid            references "User" (ID) not null,
    merchant_id              uuid            references "Merchant" (ID) not null,
    service_id               integer         references "Service" (ID) not null,
    location_id              integer         references "Location" (ID) not null,
    from_date                timestamptz     not null,
    to_date                  timestamptz     not null,
    user_note                text,
    merchant_note            text,
    price_then               bigint          not null,
    cost_then                bigint,
    cancelled_by_user_on     timestamptz,
    cancelled_by_merchant_on timestamptz,
    cancellation_reason      text,
    transferred_to           uuid
);

create table if not exists "Preferences" (
    ID                       serial           primary key unique not null,
    merchant_id              uuid             references "Merchant" (ID) not null,
    first_day_of_week        varchar(10)      default 'Monday' check (first_day_of_week in ('Monday', 'Sunday')) not null,
    time_format              varchar(10)      default '24-hour' check (time_format in ('12-hour', '24-hour')) not null,
    calendar_view            varchar(10)      default 'week' check (calendar_view in ('month', 'week', 'day', 'list')) not null,
    calendar_view_mobile     varchar(10)      default 'day' check (calendar_view_mobile in ('month', 'week', 'day', 'list')) not null,
    start_hour               time             default '08:00:00' not null,
    end_hour                 time             default '17:00:00' not null,
    time_frequency           time             default '00:15:00' not null
);

create table if not exists "Blacklist" (
    ID                       serial           primary key unique not null,
    merchant_id              uuid             references "Merchant" (ID) not null,
    user_id                  uuid             references "User" (ID) not null
);

create table if not exists "Product" (
    ID                       serial          primary key unique not null,
    merchant_id              uuid            references "Merchant" (ID) not null,
    name                     varchar(50)     not null,
    description              text,
    price                    bigint          not null,
    stock_quantity           integer         not null,
    usage_per_unit           integer         not null,
    deleted_on               timestamptz
);

create table if not exists "ServiceProduct" (
    service_id               integer references "Service" (ID) not null,
    product_id               integer references "Product" (ID) not null,
    primary key (service_id, product_id)
);

create table if not exists "BusinessHours" (
    ID                       serial                primary key unique not null,
    merchant_id              uuid                  references "Merchant" (ID) not null,
    day_of_week              smallint              check (day_of_week BETWEEN 1 AND 7) not null,
    start_time               time                  not null,
    end_time                 time                  not null,

    constraint unique_business_hours unique (merchant_id, day_of_week, start_time, end_time)
);