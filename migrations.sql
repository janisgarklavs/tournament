create table if not exists player (
    id varchar(64) not null primary key,
    balance integer not null default 0 check (balance >= 0)
);

create table if not exists tournament (
    id varchar(64) not null primary key,
    deposit integer not null,
    finished boolean not null default false
);

create table if not exists tournament_entries (
    id serial not null primary key,
    tournament_id varchar(64) not null references tournament (id),
    user_id varchar(64) not null references player (id),
    backing_id varchar(64) references player (id)
);




create table playlist (
	id serial not null primary key,
	name varchar(255) default ''::character varying not null,
	created_at timestamp default now() not null,
	updated_at timestamp default now() not null,
	selected boolean default true not null,
	repeat integer,
	org_id uuid references organization (id)
);