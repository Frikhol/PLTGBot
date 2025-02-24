CREATE TABLE "users" (
    "id" bigserial PRIMARY KEY,
    "nickname" varchar,
    "tg_id" bigint,
    "reserved_count" bigint
);

CREATE TABLE "chats" (
    "id" bigserial PRIMARY KEY,
    "chat_id" bigint,
    "is_noticing" boolean,
    "last_lost_id" bigint,
    "last_get_id" bigint
);

CREATE TABLE "chats_config" (
    "id" bigserial PRIMARY KEY,
    "chat_id" bigserial,
    "update_timeout" timestamptz,
    "noticing_limit" bigint,
    "reserve_time" timestamptz,
    "reserve_limit" bigint,
    "is_internal_ennoble" boolean,
    "is_return_noticing" boolean
);

CREATE TABLE "villages" (
    "id" bigserial PRIMARY KEY,
    "cords" varchar,
    "is_reserved" boolean,
    "reserver_id" bigserial
);

ALTER TABLE "villages" ADD FOREIGN KEY ("reserver_id") REFERENCES "users" ("id");
ALTER TABLE "chats_config" ADD FOREIGN KEY ("chat_id") REFERENCES  "chats" ("id");

