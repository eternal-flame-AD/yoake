-- Your SQL goes here
create table jpn_wordbook (
    uuid text primary key not null,
    ja text not null,
    altn text not null,
    jm text not null,
    fu text not null,
    en text not null,
    ex text not null,

    src text not null,

    created datetime not null,
    updated datetime not null
)