/*
    This file runs as part of the docker build process. If you're making 
    changes to this file, you'll have to rebuild the docker image.
*/


/* * * * * * * * * * * * * * * * * * * * * *
 *
 *          SCHEMA
 *
 * * * * * * * * * * * * * * * * * * * * * */

CREATE SCHEMA IF NOT EXISTS public
;

CREATE TABLE public.users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    nickname VARCHAR(50),
    user_type VARCHAR(50) NOT NULL DEFAULT 'UTYPE_USER',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    password   BYTEA NOT NULL
)
;

CREATE TABLE public.user_types (
    id SERIAL PRIMARY KEY,
    type_key VARCHAR(50) NOT NULL,
    permission_bitfield BIT(8) NOT NULL DEFAULT B'00000000'
)
;

ALTER TABLE public.user_types 
    ADD CONSTRAINT unique_type_key UNIQUE(type_key);


CREATE TABLE public.user_messages (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    message TEXT NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE
);

/* * * * * * * * * * * * * * * * * * * * * *
 *
 *          DATA
 *
 * * * * * * * * * * * * * * * * * * * * * */

INSERT INTO public.users 
    (username, nickname, email, user_type, password)
 VALUES 
    ('Liam', 'L dawg', 'liam@email.com', 'UTYPE_ADMIN', 'xxx'),
    ('Jon', NULL, 'jon@email.com', 'UTYPE_USER', 'xxx'),
    ('Myles', 'Big M', 'myles@email.com', 'UTYPE_USER', 'xxx')
;

INSERT INTO public.user_types 
    (type_key, permission_bitfield)
 VALUES 
    ('UTYPE_USER',      B'00000000'),
    ('UTYPE_ADMIN',     B'10000000'),
    ('UTYPE_MODERATOR', B'01000000')
;
