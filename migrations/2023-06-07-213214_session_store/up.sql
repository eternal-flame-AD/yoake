-- Your SQL goes here
create table sessions (
    uuid text primary key not null,
    expiry datetime not null,
    content text not null
);
