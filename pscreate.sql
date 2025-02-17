CREATE TABLE "users" (
                         "id" bigserial PRIMARY KEY,
                         "name" string,
                         "tg_id" string
);

CREATE TABLE "chats" (
                         "id" bigserial PRIMARY KEY,
                         "chat_id" integer,
                         "is_noticing" boolean,
                         "last_noticing_id" integer
);

CREATE TABLE "villages" (
                            "id" bigserial PRIMARY KEY,
                            "info" string,
                            "is_reserved" boolean,
                            "reserver_id" bigserial
);

ALTER TABLE "villages" ADD FOREIGN KEY ("reserver_id") REFERENCES "users" ("id");
