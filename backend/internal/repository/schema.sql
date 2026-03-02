-- When modifying this always modify structs in the backend/database as well

alter database reservations set timezone to 'UTC';
select pg_reload_conf();

create extension postgis;

create type price as (
    number                   numeric,
    currency                 char(3)
);

create type service_phase_type as ENUM ('active', 'wait');
create type subscription_tier as ENUM ('free', 'pro', 'enterprise');
create type booking_type as ENUM ('appointment', 'event', 'class');
create type booking_status as ENUM ('booked', 'confirmed', 'completed', 'cancelled', 'no-show');
create type employee_role as ENUM ('owner', 'admin', 'staff');
create type price_type as ENUM ('fixed', 'free', 'from');
create type auth_provider_type as ENUM ('facebook', 'google');
create type event_source as ENUM ('internal', 'google');
create type event_internal_type as ENUM ('booking', 'blocked_time');

create table if not exists "User" (
    ID                       uuid            primary key unique not null,
    first_name               varchar(30)     not null,
    last_name                varchar(30)     not null,
    email                    varchar(320)    not null,
    phone_number             varchar(30),
    password_hash            varchar(72),
    jwt_refresh_version      integer,
    preferred_lang           varchar(10),
    auth_provider            auth_provider_type,
    provider_id              text
);

create table if not exists "Merchant" (
    ID                       uuid            primary key unique not null,
    name                     varchar(30)     not null,
    url_name                 varchar(30)     unique not null,
    contact_email            varchar(320)    not null,
    introduction             varchar(150),
    announcement             varchar(200),
    about_us                 text,
    parking_info             text,
    payment_info             text,
    cancel_deadline          integer not null default 0,
    booking_window_min       integer not null default 0,
    booking_window_max       integer not null default 5, --  in months
    buffer_time              integer not null default 0,
    timezone                 text,
    currency_code            char(3)           not null,
    subscription_tier        subscription_tier not null
);

create table if not exists "Location" (
    ID                       serial           primary key unique not null,
    merchant_id              uuid             references "Merchant" (ID) on delete cascade not null,
    country                  varchar(50),
    city                     varchar(50),
    postal_code              varchar(10),
    address                  varchar(100),
    geo_point                geography(Point) not null,
    place_id                 text,
    formatted_location       text             not null,
    is_primary               boolean          not null,
    is_active                boolean          not null default true
);

create table if not exists "Employee" (
    ID                       serial          primary key unique not null,
    user_id                  uuid            references "User" (ID) on delete set null,
    merchant_id              uuid            references "Merchant" (ID) on delete cascade not null,
    role                     employee_role   not null default 'staff',
    first_name               varchar(30),
    last_name                varchar(30),
    email                    varchar(320),
    phone_number             varchar(30),
    is_active                boolean         not null default true,
    invited_on               timestamptz,
    accepted_on              timestamptz,

    constraint unique_merchant_user_employee unique (merchant_id, user_id)
);

create table if not exists "ServiceCategory" (
    ID                       serial          primary key unique not null,
    merchant_id              uuid            references "Merchant" (ID) on delete cascade not null,
    location_id              integer         references "Location" (ID) on delete cascade,
    name                     varchar(30)     not null,
    sequence                 integer         not null default 0
);

create table if not exists "Service" (
    ID                       serial          primary key unique not null,
    merchant_id              uuid            references "Merchant" (ID) on delete cascade not null,
    category_id              integer         references "ServiceCategory" (ID) on delete set null,
    booking_type             booking_type    not null,
    name                     varchar(30)     not null,
    description              varchar(200),
    color                    char(7)         not null,
    total_duration           integer         not null,
    price_per_person         price,
    cost_per_person          price,
    price_type               price_type      not null default 'fixed',
    is_active                boolean         not null,
    sequence                 integer         not null default 0,
    min_participants         integer         not null,
    max_participants         integer         not null,
    cancel_deadline          integer,
    booking_window_min       integer,
    booking_window_max       integer,
    buffer_time              integer,
    deleted_on               timestamptz
);

create table if not exists "ServicePhase" (
    ID                       serial                 primary key unique not null,
    service_id               integer                references "Service" (ID) on delete cascade not null,
    name                     varchar(30)            not null,
    sequence                 integer                not null,
    duration                 integer                not null,
    phase_type               service_phase_type     not null,
    deleted_on               timestamptz,

    constraint unique_service_phase_sequence unique (service_id, sequence)
);

-- constraint is neccessary for the on conflict
create table if not exists "Customer" (
    ID                      uuid            primary key unique not null,
    merchant_id             uuid            references "Merchant" (ID) on delete cascade not null,
    user_id                 uuid            references "User" (ID),
    first_name              varchar(30),
    last_name               varchar(30),
    email                   varchar(320),
    phone_number            varchar(30),
    birthday                date,
    note                    text,
    is_blacklisted          boolean default false not null,
    blacklist_reason        text,

    constraint unique_merchant_user unique (merchant_id, user_id)
);

create table if not exists "BookingSeries" (
    ID                       serial          primary key unique not null,
    booking_type             booking_type    not null,
    merchant_id              uuid            references "Merchant" (ID) on delete cascade not null,
    employee_id              integer         references "Employee" (ID) on delete set null,
    service_id               integer         references "Service" (ID) not null,
    location_id              integer         references "Location" (ID) not null,
    rrule                    text            not null,
    dstart                   timestamptz     not null,
    timezone                 text            not null,
    is_active                boolean         not null default true
);

create table if not exists "BookingSeriesDetails" (
    ID                       serial          primary key unique not null,
    booking_series_id        integer         references "BookingSeries" (ID) on delete cascade not null,
    price_per_person         price           not null,
    cost_per_person          price           not null,
    total_price              price           not null,
    total_cost               price           not null,
    min_participants         integer         not null,
    max_participants         integer         not null,
    current_participants     integer         not null
);

create table if not exists "BookingSeriesParticipant" (
    ID                       serial          primary key unique not null,
    booking_series_id        integer         references "BookingSeries" (ID) on delete cascade not null,
    customer_id              uuid            references "Customer" (ID) on delete cascade,
    is_active                boolean         not null default true,
    dropped_out_on           timestamptz,

    constraint unique_booking_series_participant unique (booking_series_id, customer_id)
);

create table if not exists "Booking" (
    ID                       serial          primary key unique not null,
    status                   booking_status  not null default 'booked',
    booking_type             booking_type    not null,
    is_recurring             boolean         not null default false,
    merchant_id              uuid            references "Merchant" (ID) on delete cascade not null,
    employee_id              integer         references "Employee" (ID) on delete set null,
    service_id               integer         references "Service" (ID) not null,
    location_id              integer         references "Location" (ID) not null,
    booking_series_id        integer         references "BookingSeries" (ID) on delete set null,
    series_original_date     timestamptz,
    from_date                timestamptz     not null,
    to_date                  timestamptz     not null
);

create table if not exists "BookingPhase" (
    ID                       serial           primary key unique not null,
    booking_id               integer          references "Booking" (ID) on delete cascade not null,
    service_phase_id         integer          references "ServicePhase" (ID) not null,
    from_date                timestamptz      not null,
    to_date                  timestamptz      not null,

    constraint unique_booking_phase unique (booking_id, service_phase_id)
);

create table if not exists "BookingDetails" (
    ID                       serial           primary key unique not null,
    booking_id               integer          unique references "Booking" (ID) on delete cascade not null,
    price_per_person         price            not null,
    cost_per_person          price            not null,
    total_price              price            not null,
    total_cost               price            not null,
    merchant_note            text,
    min_participants         integer          not null,
    max_participants         integer          not null,
    current_participants     integer          not null,
    cancelled_by_merchant_on timestamptz,
    cancellation_reason      text
);

create table if not exists "BookingParticipant" (
    ID                       serial           primary key unique not null,
    status                   booking_status   not null default 'booked',
    booking_id               integer          references "Booking" (ID) on delete cascade not null,
    customer_id              uuid             references "Customer" (ID) on delete cascade,
    customer_note            text,
    cancelled_on             timestamptz,
    cancellation_reason      text,
    transferred_to           uuid,
    email_id                 uuid,

    constraint unique_booking_participant unique (booking_id, customer_id)
);

create table if not exists "Preferences" (
    ID                       serial           primary key unique not null,
    merchant_id              uuid             references "Merchant" (ID) on delete cascade not null,
    first_day_of_week        varchar(10)      default 'Monday' check (first_day_of_week in ('Monday', 'Sunday')) not null,
    time_format              varchar(10)      default '24-hour' check (time_format in ('12-hour', '24-hour')) not null,
    calendar_view            varchar(10)      default 'week' check (calendar_view in ('month', 'week', 'day', 'list')) not null,
    calendar_view_mobile     varchar(10)      default 'day' check (calendar_view_mobile in ('month', 'week', 'day', 'list')) not null,
    start_hour               time(0)          default '08:00:00' not null,
    end_hour                 time(0)          default '17:00:00' not null,
    time_frequency           time(0)          default '00:15:00' not null
);

create table if not exists "Product" (
    ID                       serial          primary key unique not null,
    merchant_id              uuid            references "Merchant" (ID) on delete cascade not null,
    name                     varchar(50)     not null,
    description              text,
    price                    price,
    unit                     varchar(10)     check (unit in ('ml', 'g', 'pcs')) not null,
    max_amount               bigint          not null,
    current_amount           bigint          not null,
    deleted_on               timestamptz
);

create table if not exists "ServiceProduct" (
    service_id               integer references "Service" (ID) on delete cascade not null,
    product_id               integer references "Product" (ID) not null,
    amount_used              bigint  not null,
    primary key (service_id, product_id)
);

create table if not exists "BusinessHours" (
    ID                       serial          primary key unique not null,
    merchant_id              uuid            references "Merchant" (ID) on delete cascade not null,
    day_of_week              smallint        check (day_of_week BETWEEN 0 AND 6) not null,
    start_time               time(0)         not null,
    end_time                 time(0)         not null,

    constraint unique_business_hours unique (merchant_id, day_of_week, start_time, end_time)
);

create table if not exists "BlockedTime" (
    ID                       serial          primary key unique not null,
    merchant_id              uuid            references "Merchant" (ID) on delete cascade not null,
    employee_id              integer         references "Employee" (ID) on delete cascade not null,
    blocked_type_id          integer         references "BlockedTimeType" (ID) on delete set null,
    name                     varchar(50)     not null,
    from_date                timestamptz     not null,
    to_date                  timestamptz     not null,
    all_day                  boolean         not null,
    source                   event_source
);

create table if not exists "BlockedTimeType" (
    ID                       serial          primary key unique not null,
    merchant_id              uuid            references "Merchant" (ID) on delete cascade,
    name                     varchar(50)     not null,
    duration                 integer         not null,
    icon                     varchar(10)
);

create table if not exists "ExternalCalendar" (
    ID                       serial           primary key unique not null,
    employee_id              integer          references "Employee" (ID) on delete cascade not null,
    calendar_id              text             not null,
    access_token             text             not null,
    refresh_token            text             not null,
    token_expiry             timestamptz      not null,
    sync_token               text,
    channel_id               text,
    resource_id              text,
    channel_expiry           timestamptz,
    timezone                 text             not null
);

create table if not exists "ExternalCalendarEvent" (
    ID                       serial              primary key unique not null,
    external_calendar_id     integer             references "ExternalCalendar" (ID) on delete cascade not null,
    external_event_id        text                not null,
    etag                     text                not null,
    status                   text                not null,
    title                    text                not null,
    description              text                not null,
    from_date                timestamptz         not null,
    to_date                  timestamptz         not null,
    is_all_day               boolean             not null,
    internal_id              integer,
    internal_type            event_internal_type,
    is_blocking              boolean             not null,
    source                   event_source        not null,
    last_synced_at           timestamptz         not null default now(),

    unique (external_calendar_id, external_event_id)
);