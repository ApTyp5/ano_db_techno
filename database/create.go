package database

import (
	"database/sql"
)

func CreateTables(db *sql.DB) {
	_, err := db.Exec(`

CREATE OR REPLACE LANGUAGE plpgsql;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
	email citext UNIQUE NOT NULL ,
	nick_name citext PRIMARY KEY ,
	full_name text NOT NULL ,
	about text NULL
);
    
CREATE TABLE forums (
	slug citext PRIMARY KEY ,
	title text NOT NULL,
	responsible citext REFERENCES users(nick_name) NOT NULL ,
	post_num integer NOT NULL DEFAULT 0,
	thread_num integer NOT NULL DEFAULT 0
-- 	check (Slug ~ $$^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$$)
);

create index on forums using hash(slug);

CREATE TABLE threads (
	id serial PRIMARY KEY ,
	author citext REFERENCES users(nick_name) NOT NULL ,
	forum citext REFERENCES forums(slug) NOT NULL,
	created timestamptz NOT NULL DEFAULT now(),
	message text NOT NULL ,
	slug citext NULL ,
	title text NOT NULL ,
	vote_num integer default 0 NOT NULL 
-- 	check (Slug ~ $$^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$$)
);

create index on threads using hash(id);
create index on	threads using hash(slug) where slug != '';

CREATE TABLE votes (
	author citext REFERENCES users(nick_name) NOT NULL ,
	thread integer REFERENCES threads(id) NOT NULL ,
	voice integer NOT NULL,
	PRIMARY KEY (author, thread),
	CHECK ( voice = 1 OR voice = -1)
);

CREATE TABLE posts (
    id serial PRIMARY KEY ,
    parent integer REFERENCES posts(id) DEFAULT NULL,
	author citext REFERENCES users(nick_name) NOT NULL ,
	thread integer REFERENCES threads(id) NOT NULL ,
	created timestamptz NOT NULL DEFAULT now(),
	is_edited bool DEFAULT FALSE NOT NULL,
	message text NOT NULL
);

create index on posts using hash(id);

CREATE TABLE status (
    forum_num integer DEFAULT 0,
    thread_num integer DEFAULT 0,
    post_num integer DEFAULT 0,
    user_num integer DEFAULT 0
);
INSERT INTO status DEFAULT VALUES ;

CREATE OR REPLACE FUNCTION set_post_is_edited() RETURNS TRIGGER AS $setPostIsEdited$
	begin
		if (not old.is_edited) and (old.message != new.message) then
			new.is_edited := true;
		end if;
		return new;
	end;
$setPostIsEdited$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION post_num_inc() RETURNS TRIGGER AS $postNumInc$
	begin
		update Forums set post_num = post_num + 1
			where slug = (
				select forum
				from Threads t 
				where new.thread = t.id
			);	
		update Status set post_num = post_num + 1;
		return new;
	end;
$postNumInc$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION thread_num_inc() RETURNS TRIGGER AS $threadNumInc$ 
	begin
	    update Forums set thread_num = thread_num + 1
	    	where slug = new.forum;
	    update Status set thread_num = thread_num + 1;
	    return new;
	end;
$threadNumInc$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION forum_num_inc() RETURNS TRIGGER AS $forumNumInc$
	begin 
   		update Status set forum_num = forum_num + 1;
   		return new;
	end;
$forumNumInc$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION thread_rating_count() RETURNS TRIGGER AS $threadRatingCount$
	begin
		update Threads set vote_num = vote_num + new.voice
	    	where id = new.thread;
		return new;
	end;
$threadRatingCount$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION thread_rating_recount() RETURNS TRIGGER AS $threadRatingCount$
	begin
	    if new.voice = old.voice then
			return new;
		end if;
	    
		update Threads set vote_num = vote_num + new.voice - old.voice
	    	where id = new.thread;
		return new;
	end;
$threadRatingCount$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION user_num_inc() RETURNS TRIGGER AS $userNumInc$
	begin
		update Status set user_num = user_num + 1;
		return new;
	end;
$userNumInc$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION post_set_id_check_parent() RETURNS TRIGGER AS $postsSetId$
	begin
	    if new.parent is not null then
	        if new.thread != (select t.id from Threads t join Posts P on t.id = P.thread where P.id = new.parent) then
				raise EXCEPTION 'Parent post was created in another thread';
			end if;
		end if;
	    
		return new;
	end;
$postsSetId$ LANGUAGE plpgsql;

CREATE TRIGGER posts_check_par BEFORE INSERT ON posts FOR EACH ROW EXECUTE PROCEDURE post_set_id_check_parent();
CREATE TRIGGER post_num_inc AFTER INSERT ON postS FOR EACH ROW EXECUTE PROCEDURE  post_num_inc();
CREATE TRIGGER thread_num_inc AFTER INSERT ON threads FOR EACH ROW EXECUTE PROCEDURE  thread_num_inc();
CREATE TRIGGER thread_rating_count AFTER INSERT ON votes FOR EACH ROW EXECUTE PROCEDURE  thread_rating_count();
CREATE TRIGGER thread_rating_recount AFTER UPDATE ON votes FOR EACH ROW EXECUTE PROCEDURE  thread_rating_recount();
CREATE TRIGGER set_post_is_edited BEFORE UPDATE ON posts FOR EACH ROW EXECUTE PROCEDURE  set_post_is_edited();
CREATE TRIGGER forum_num_inc AFTER INSERT ON forums FOR EACH ROW EXECUTE PROCEDURE  forum_num_inc();
CREATE TRIGGER user_num_inc AFTER INSERT ON users FOR EACH ROW EXECUTE PROCEDURE user_num_inc();
`)
	if err != nil {
		panic(err)
	}
}
