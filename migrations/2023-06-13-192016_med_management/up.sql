-- Your SQL goes here
create table medications (
    uuid text primary key not null,
    name text not null,
    dosage integer not null,
    dosage_unit text not null,
    period_hours integer not null,
    flags text not null,
    options text not null,

    created datetime not null,
    updated datetime not null
);

create table medication_logs (
    uuid text primary key not null,
    med_uuid text not null,
    dosage integer not null,
    time_actual datetime not null,
    time_expected datetime not null,
    dose_offset real not null,

    created datetime not null,
    updated datetime not null,

    FOREIGN KEY(med_uuid) REFERENCES medications(uuid)
)