CREATE TABLE "users" (
    "id" bigserial PRIMARY KEY,
    "nickname" varchar,
    "tg_id" varchar,
    "role" varchar,
    "reserved_count" integer
);

CREATE TABLE "chats" (
    "id" bigserial PRIMARY KEY,
    "chat_id" integer,
    "is_noticing" boolean,
    "last_lost_id" integer,
    "last_get_id" integer
);

CREATE TABLE "chats_config" (
    "id" bigserial PRIMARY KEY,
    "chat_id" bigserial,
    "update_timeout" integer,
    "noticing_limit" integer,
    "reserve_time" timestamptz,
    "reserve_limit" integer,
    "is_internal_ennoble" boolean,
    "is_return_noticing" boolean
);

CREATE TABLE "villages" (
    "id" bigserial PRIMARY KEY,
    "info" TEXT,
    "is_reserved" boolean,
    "reserver_id" bigserial
);

ALTER TABLE "villages" ADD FOREIGN KEY ("reserver_id") REFERENCES "users" ("id");
ALTER TABLE "chats_config" ADD FOREIGN KEY ("chat_id") REFERENCES  "chats" ("id");