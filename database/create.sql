CREATE OR REPLACE LANGUAGE plpgsql;
CREATE EXTENSION IF NOT EXISTS citext;



DROP TABLE IF EXISTS users;
CREATE TABLE users (
   email citext UNIQUE NOT NULL ,
   nick_name citext PRIMARY KEY ,
   full_name text NOT NULL ,
   about text NULL
);
DROP INDEX IF EXISTS  users__nick_name__idx;
CREATE INDEX users__nick_name__idx on users (nick_name);



DROP TABLE IF EXISTS forums;
CREATE TABLE forums (
    slug citext PRIMARY KEY ,
    title text NOT NULL,
    responsible citext REFERENCES users(nick_name) NOT NULL ,
    post_num integer NOT NULL DEFAULT 0,
    thread_num integer NOT NULL DEFAULT 0
);
DROP INDEX IF EXISTS forums__slug__idx;
CREATE INDEX forums__slug__idx ON forums(slug);



DROP TABLE IF EXISTS threads;
CREATE TABLE threads (
    id serial PRIMARY KEY ,
    author citext REFERENCES users(nick_name) NOT NULL ,
    forum citext REFERENCES forums(slug) NOT NULL,
    created timestamptz NOT NULL DEFAULT now(),
    message text NOT NULL ,
    slug citext NULL ,
    title text NOT NULL ,
    vote_num integer default 0 NOT NULL
);
DROP INDEX IF EXISTS threads__id__idx;
CREATE INDEX threads__id__idx ON threads(id);

DROP INDEX IF EXISTS threads__slug__idx__not_null;
CREATE INDEX threads__slug__idx__not_null ON threads(slug) WHERE slug IS NOT NULL ;

DROP INDEX IF EXISTS threads__created__idx;
CREATE INDEX threads__created__idx ON threads(created);



DROP TABLE IF EXISTS votes;
CREATE TABLE votes (
    author citext REFERENCES users(nick_name) NOT NULL ,
    thread integer REFERENCES threads(id) NOT NULL ,
    voice integer NOT NULL,
    PRIMARY KEY (author, thread),
    CHECK ( voice = 1 OR voice = -1)
);



DROP TABLE IF EXISTS posts;
CREATE TABLE posts (
    id serial PRIMARY KEY ,
    path integer[],
    parent integer REFERENCES posts(id) DEFAULT NULL,
    forum citext REFERENCES forums(SLUG) NOT NULL ,
    author citext REFERENCES users(nick_name) NOT NULL ,
    thread integer REFERENCES threads(id) NOT NULL ,
    created timestamptz NOT NULL DEFAULT now(),
    is_edited bool DEFAULT FALSE NOT NULL,
    message text NOT NULL
);



DROP TABLE IF EXISTS status;
CREATE TABLE status (
    forum_num integer DEFAULT 0,
    thread_num integer DEFAULT 0,
    post_num integer DEFAULT 0,
    user_num integer DEFAULT 0
);
INSERT INTO status DEFAULT VALUES ;



DROP TABLE IF EXISTS forum_users;
CREATE TABLE forum_users (
    forum citext REFERENCES forums(slug),
    user_nick citext REFERENCES users(nick_name)
);


DROP FUNCTION IF EXISTS set_post_is_edited;
CREATE OR REPLACE FUNCTION set_post_is_edited() RETURNS TRIGGER AS $set_post_is_edited$
begin
    if (not old.is_edited) and (old.message != new.message) then
        new.is_edited := true;
    end if;
    return new;
end;
$set_post_is_edited$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS set_post_is_edited ON posts;
CREATE TRIGGER set_post_is_edited
    BEFORE UPDATE
    ON posts
    FOR EACH ROW
EXECUTE PROCEDURE  set_post_is_edited();



DROP FUNCTION IF EXISTS thread_num_inc;
CREATE OR REPLACE FUNCTION thread_num_inc() RETURNS TRIGGER AS $thread_num_inc$
begin
    update Forums set thread_num = thread_num + 1
    where slug = new.forum;
    update Status set thread_num = thread_num + 1;
    return new;
end;
$thread_num_inc$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS thread_num_inc ON threads;
CREATE TRIGGER thread_num_inc
    AFTER INSERT ON threads
    FOR EACH ROW
EXECUTE PROCEDURE  thread_num_inc();



DROP FUNCTION IF EXISTS forum_num_inc;
CREATE OR REPLACE FUNCTION forum_num_inc() RETURNS TRIGGER AS $forum_num_inc$
begin
    update Status set forum_num = forum_num + 1;
    return new;
end;
$forum_num_inc$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS forum_num_inc ON forums;
CREATE TRIGGER forum_num_inc
    AFTER INSERT ON forums
    FOR EACH ROW
EXECUTE PROCEDURE  forum_num_inc();



DROP FUNCTION IF EXISTS thread_rating_count;
CREATE OR REPLACE FUNCTION thread_rating_count() RETURNS TRIGGER AS $thread_rating_count$
begin
    update Threads set vote_num = vote_num + new.voice
    where id = new.thread;
    return new;
end;
$thread_rating_count$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS thread_rating_count ON votes;
CREATE TRIGGER thread_rating_count
    AFTER INSERT
    ON votes
    FOR EACH ROW
EXECUTE PROCEDURE  thread_rating_count();



DROP FUNCTION IF EXISTS thread_rating_recount;
CREATE OR REPLACE FUNCTION thread_rating_recount() RETURNS TRIGGER AS $thread_rating_recount$
begin
    if new.voice = old.voice then
        return new;
    end if;

    update Threads set vote_num = vote_num + new.voice - old.voice
    where id = new.thread;
    return new;
end;
$thread_rating_recount$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS thread_rating_recount ON votes;
CREATE TRIGGER thread_rating_recount
    AFTER UPDATE
    ON votes
    FOR EACH ROW
EXECUTE PROCEDURE  thread_rating_recount();



DROP FUNCTION IF EXISTS user_num_inc;
CREATE OR REPLACE FUNCTION user_num_inc() RETURNS TRIGGER AS $user_num_inc$
begin
    update Status set user_num = user_num + 1;
    return new;
end;
$user_num_inc$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS user_num_inc ON users;
CREATE TRIGGER user_num_inc
    AFTER INSERT
    ON users
    FOR EACH ROW
EXECUTE PROCEDURE user_num_inc();



DROP FUNCTION IF EXISTS post_check_parent;
CREATE OR REPLACE FUNCTION post_check_parent() RETURNS TRIGGER AS $post_check_parent$
begin
    if new.parent is not null then
        if new.thread != (select P.thread from Posts P where P.id = new.parent) then
            raise EXCEPTION 'Parent post was created in another thread';
        end if;
    end if;

    return new;
end;
$post_check_parent$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS posts_check_parent ON posts;
CREATE TRIGGER posts_check_parent
    BEFORE INSERT
    ON posts
    FOR EACH ROW
EXECUTE PROCEDURE post_check_parent();



DROP FUNCTION IF EXISTS post_set_path;
CREATE OR REPLACE FUNCTION post_set_path() RETURNS TRIGGER AS $post_set_path$
begin
    if new.parent is null then
        new.path = array [new.id];
        return new;
    end if;

    new.path = (select path from posts where id = new.parent) || array [new.id];
    return new;
end;
$post_set_path$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS post_set_path ON posts;
CREATE TRIGGER post_set_path
    BEFORE INSERT
    ON posts
    FOR EACH ROW
EXECUTE PROCEDURE post_set_path();



CREATE OR REPLACE FUNCTION add_forum_user() RETURNS TRIGGER AS $add_forum_user$
begin
    INSERT INTO forum_users (FORUM, USER_NICK)
    VALUES (new.forum, new.author)
    ON CONFLICT DO NOTHING ;
end;
$add_forum_user$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS add_forum_user on posts;
CREATE TRIGGER add_forum_user
    AFTER INSERT
    ON posts
    FOR EACH ROW
EXECUTE PROCEDURE add_forum_user();

DROP TRIGGER IF EXISTS add_forum_user on threads;
CREATE TRIGGER add_forum_user
    AFTER INSERT
    ON threads
    FOR EACH ROW
EXECUTE PROCEDURE add_forum_user();