begin;

alter table "Location"
    add column if not exists name text,
    add column if not exists sequence integer,
    add column if not exists country_code char(2),
    add column if not exists region text,
    add column if not exists geocoding_provider text,
    add column if not exists provider_place_id text,
    add column if not exists contact_email text,
    add column if not exists phone_number text,
    add column if not exists parking_info text,
    add column if not exists arrival_info text;

with ranked_locations as (
    select id, row_number() over (
        partition by merchant_id
        order by is_primary desc, id
    )::integer - 1 as sequence
    from "Location"
)
update "Location" l
set name = coalesce(l.name, 'Main location'),
    sequence = coalesce(l.sequence, ranked_locations.sequence),
    geocoding_provider = case
        when l.place_id is not null then 'mapbox'
        else l.geocoding_provider
    end,
    provider_place_id = coalesce(l.provider_place_id, l.place_id),
    parking_info = coalesce(l.parking_info, m.parking_info, ''),
    arrival_info = coalesce(l.arrival_info, '')
from "Merchant" m, ranked_locations
where m.id = l.merchant_id
  and ranked_locations.id = l.id;

alter table "Location"
    alter column name set default 'Main location',
    alter column name set not null,
    alter column sequence set default 0,
    alter column sequence set not null,
    alter column country type text,
    alter column city type text,
    alter column postal_code type text,
    alter column address type text,
    alter column geo_point type geography(Point, 4326)
        using geo_point::geography(Point, 4326),
    alter column parking_info set default '',
    alter column parking_info set not null,
    alter column arrival_info set default '',
    alter column arrival_info set not null;

alter table "Location"
    add constraint location_geocoding_provider_complete check (
        (geocoding_provider is null) = (provider_place_id is null)
    );

drop index if exists location_one_primary_per_merchant;

alter table "Location"
    drop column if exists is_primary,
    drop column if exists place_id;

create index if not exists location_merchant_sequence
on "Location" (merchant_id, is_active, sequence, ID);

alter table "BusinessHours"
    add column if not exists location_id integer;

update "BusinessHours" bh
set location_id = (
    select l.id
    from "Location" l
    where l.merchant_id = bh.merchant_id
    order by l.is_active desc, l.sequence, l.id
    limit 1
)
where bh.location_id is null;

-- Merchants that abandoned signup before choosing an address have default
-- business hours but no physical location. A location creation will recreate
-- these defaults if they resume onboarding.
delete from "BusinessHours"
where location_id is null;

alter table "BusinessHours"
    alter column location_id set not null,
    add constraint business_hours_location_id_fkey
        foreign key (location_id) references "Location" (ID) on delete cascade;

alter table "BusinessHours"
    drop constraint if exists unique_business_hours;

alter table "BusinessHours"
    add constraint unique_business_hours
        unique (location_id, day_of_week, start_time, end_time),
    drop column if exists merchant_id;

alter table "Merchant"
    drop column if exists parking_info;

commit;
