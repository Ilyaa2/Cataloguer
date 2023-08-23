CREATE table messages
(
    id         BIGSERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users (id),
    name       varchar,
    type       varchar   not null,
    time       timestamp not null,
    path       varchar   not null
);