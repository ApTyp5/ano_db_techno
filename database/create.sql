CREATE OR REPLACE LANGUAGE plpgsql;
CREATE EXTENSION IF NOT EXISTS citext;



DROP TABLE IF EXISTS users CASCADE ;
CREATE TABLE users (
    about text NULL,
    email citext UNIQUE NOT NULL ,
    full_name text NOT NULL ,
    nick_name citext PRIMARY KEY
);



DROP TABLE IF EXISTS forums CASCADE ;
CREATE TABLE forums (
    slug citext PRIMARY KEY ,
    title text NOT NULL,
    responsible citext REFERENCES users(nick_name) NOT NULL ,
    post_num integer NOT NULL DEFAULT 0,
    thread_num integer NOT NULL DEFAULT 0
);



DROP TABLE IF EXISTS threads CASCADE ;
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

DROP INDEX IF EXISTS threads__slug__idx__not_null;
CREATE INDEX threads__slug__idx__not_null ON threads(slug) WHERE slug IS NOT NULL ;

DROP INDEX IF EXISTS threads__created__idx;
CREATE INDEX threads__created__idx ON threads(created);

DROP INDEX IF EXISTS threads__forum__hidx;
CREATE INDEX threads__forum__hidx ON threads USING hash(forum);



DROP TABLE IF EXISTS votes;
CREATE TABLE votes (
    author citext REFERENCES users(nick_name) NOT NULL ,
    thread integer REFERENCES threads(id) NOT NULL ,
    voice integer NOT NULL,
    PRIMARY KEY (author, thread),
    CHECK ( voice = 1 OR voice = -1)
);



DROP TABLE IF EXISTS posts CASCADE ;
CREATE TABLE posts (
    author citext REFERENCES users(nick_name) NOT NULL ,
    created timestamptz NOT NULL DEFAULT now(),
    id serial PRIMARY KEY ,
    is_edited bool DEFAULT FALSE NOT NULL,
    message text NOT NULL,
    parent integer REFERENCES posts(id) DEFAULT NULL,
    thread integer REFERENCES threads(id) NOT NULL ,
    forum citext REFERENCES forums(SLUG) NOT NULL ,
    path integer[]
);

DROP INDEX IF EXISTS posts__created__idx;
CREATE INDEX posts__created__idx ON posts(created);

DROP INDEX IF EXISTS posts__path__idx;
CREATE INDEX posts__path__idx ON posts(path);



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
    forum citext REFERENCES forums(slug) ,
    user_nick citext REFERENCES users(nick_name),
    primary key (forum, user_nick)
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
    return new;
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



DROP FUNCTION IF EXISTS select_threads_by_forum(forum citext, lmt integer, snc text, dsc bool);
CREATE OR REPLACE FUNCTION select_threads_by_forum(fslug citext, lmt integer, snc text, dsc bool)
    RETURNS SETOF threads AS $$
declare
    queryS text;
begin
    queryS := 'SELECT id, author, forum, ' ||
             'created, message, slug, ' ||
             'title, vote_num ' ||
             'FROM threads ' ||
             'WHERE forum = ' || quote_literal(fslug);

    if snc is not null then
        queryS = queryS || ' and created ';
        if dsc then
            queryS = queryS || ' <= ';
        else
            queryS = queryS || ' >= ';
        end if;
        queryS = queryS || quote_literal(snc);
    end if;

    queryS = queryS || ' order by created ';
    if dsc then
        queryS = queryS || ' desc ';
    end if;

    if lmt > 0 then
        queryS = queryS || ' limit ' || lmt;
    end if;

    return query execute queryS;
end
$$ LANGUAGE plpgsql;



DROP FUNCTION IF EXISTS select_posts_by_thread(threadId integer, lmt integer, snc integer, dsc bool, mode text);
CREATE OR REPLACE FUNCTION select_posts_by_thread(threadId integer, lmt integer, snc integer, dsc bool, mode text)
    RETURNS SETOF posts AS $$
    declare
        mainPart text;
        wherePart text;
        orderPart text;
    begin
        -- mode = 1, flag; 2, tree; 3, par_tree
        mainPart := 'SELECT author, created, id, ' ||
                    'is_edited, message, coalesce(parent, 0),' ||
                    'thread, forum, path ' ||
                    'FROM posts ';
        wherePart := 'WHERE ';
        orderPart := 'ORDER BY ';

        if mode = '' or mode = 'flat' then
            wherePart = wherePart || ' thread = ' || threadId;

            if snc > 0 then
                wherePart = wherePart || ' and id ';
                if dsc then
                    wherePart = wherePart || ' < ';
                else
                    wherePart = wherePart || ' > ';
                end if;
                wherePart = wherePart || quote_literal(snc);
            end if;

            orderPart = orderPart || ' created ';
            if dsc then
                orderPart = orderPart || ' desc ';
            end if;

            orderPart = orderPart || ', id ';
            if dsc then
                orderPart = orderPart || ' desc ';
            end if;

            if lmt > 0 then
                orderPart = orderPart || ' limit ' || lmt;
            end if;
        end if;

        if mode = 'tree' then
            wherePart = wherePart || ' thread = ' || threadId;

            if snc > 0 then
                wherePart = wherePart || ' and path ';
                if dsc then
                    wherePart = wherePart || ' < ';
                else
                    wherePart = wherePart || ' > ';
                end if;

                wherePart = wherePart || '(select path from posts ' ||
                            'where id = ' || snc || ')';
            end if;

            orderPart = orderPart || ' path ';
            if dsc then
                orderPart = orderPart || ' desc ';
            end if;

            if lmt > 0 then
                orderPart = orderPart || ' limit ' || lmt;
            end if;
        end if;

        if mode = 'parent_tree' then
            wherePart = wherePart || 'path[1] in ' ||
                        '(select path[1] ' ||
                        'from posts ' ||
                        'where thread = ' ||
                        threadId ||
                        ' and parent is null ';
            if snc > 0 then
                if dsc then
                    wherePart = wherePart || ' and id < ';
                else
                    wherePart = wherePart || ' and id > ';
                end if;
                wherePart = wherePart ||
                    '(select path[1] from posts where id = ' ||
                            snc || ')';
                end if;

                wherePart = wherePart || ' order by id ';
                if dsc then
                    wherePart = wherePart || ' desc ';
                end if;

                if lmt > 0 then
                    wherePart = wherePart || ' limit ' || lmt;
                end if;

                wherePart = wherePart || ')';

                if dsc then
                    orderPart = orderPart || ' path[1] desc, path[2:] ';
                else
                    orderPart = orderPart || ' path ';
                end if;
            end if;

        mainPart = mainPart || wherePart || orderPart;
        return query execute mainPart;
    end
$$ LANGUAGE plpgsql;



DROP FUNCTION IF EXISTS select_users_by_forum(forum citext, dsc bool, lim integer, sinc citext);
CREATE OR REPLACE FUNCTION select_users_by_forum(forum citext, dsc bool, lim integer, sinc citext)
    RETURNS SETOF users AS $$
declare
    queryS text;
begin
    queryS := 'SELECT u.about, u.email, u.full_name, u.nick_name ' ||
             'from forum_users fu ' ||
             'join users u on u.nick_name = fu.user_nick ' ||
             'where fu.forum = ' || quote_literal(forum);

    if sinc <> '' then
        queryS = queryS || 'and u.nick_name ';
        if dsc then
            queryS = queryS || ' < ' ;
        else
            queryS = queryS || ' > ';
        end if;
        queryS = queryS || quote_literal(sinc);
    end if;

    queryS = queryS || ' order by u.nick_name ';
    if dsc then
        queryS = queryS || ' desc ';
    end if;

    if lim > 0 then
        queryS = queryS || ' limit ' || lim;
    end if;

    return query execute queryS;
end
$$ LANGUAGE plpgsql;