begin;

-- Blocked times are disposable in development. Clearing them lets the new
-- mutually-exclusive timed/all-day shape be installed without backfilling
-- date-only values from timezone-dependent timestamps.
delete from "ExternalCalendarEvent"
where internal_type = 'blocked_time';

delete from "BlockedTime";

alter table "BlockedTime"
    rename column all_day to is_all_day;

alter table "BlockedTime"
    alter column from_date drop not null,
    alter column to_date drop not null,
    add column blocked_day date;

alter table "BlockedTime"
    add constraint blocked_time_value_shape check (
        (is_all_day and blocked_day is not null and from_date is null and to_date is null)
        or
        (not is_all_day and blocked_day is null and from_date is not null and to_date is not null and from_date < to_date)
    );

commit;
