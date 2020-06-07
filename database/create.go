package database

import (
	"database/sql"
	_ "github.com/lib/pq"
)

func CreateTables(db *sql.DB) {
	_, err := db.Exec(`

create or replace language plpgsql;
create extension if not exists citext;

create table Users (
	Id serial primary key,
	NickName citext unique not null,
	FullName text not null,
	Email citext unique not null,
	About text null
);

create index on Users (id);
create index on Users (NickName);
    
create table Forums (
	Id serial primary key,
	Slug citext unique not null,
	Title text not null,
	Responsible integer references Users(Id) not null,
	PostNum integer not null default 0,
	ThreadNum integer not null default 0,
	check (Slug ~ $$^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$$)
);

create index on Forums (Id);
create index on Forums (Slug);

create table Threads (
	Id serial primary key,
	Author integer references Users(Id) not null,
	Forum citext references Forums(Slug) not null,
	Created timestamptz not null default now(),
	Message text not null,
	Slug citext null,
	Title text not null,
	VoteNum integer default 0 not null
-- 	check (Slug ~ $$^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$$)
);

create index on Threads (ID);
create index on Threads (Slug);

create table Votes (
	Author integer references Users(Id) not null,
	Thread integer references Threads(Id) not null,
	Voice integer not null,
	primary key (Author, Thread),
	check ( voice = 1 or voice = -1)
);

create table Posts (
    Id serial primary key ,
    Parent integer references Posts(Id) default null check ( Id != Parent ),
	Author integer references Users(Id) not null,
	Thread integer references Threads(Id) not null,
	Created timestamptz not null default now(),
	IsEdited bool default false not null,
	Message text not null
);


create table Status (
    ForumNum integer,
    ThreadNum integer,
    PostNum integer,
    UserNum integer
);
insert into Status values (0, 0, 0, 0);

create or replace function PostId(p Posts) returns integer as $Id$
begin 
	   return p.Id;
end;
$Id$ language plpgsql;

create or replace function PostPar(p Posts) returns integer as $Parent$
begin 
    return p.Parent;
end;
$Parent$ language plpgsql;

create or replace function setPostIsEdited() returns trigger as $setPostIsEdited$
	begin
		if (not old.IsEdited) and (old.Message != new.Message) then
			new.IsEdited := true;
		end if;
		return new;
	end;
$setPostIsEdited$ language plpgsql;

create or replace function postNumInc() returns trigger as $postNumInc$ 
	begin
		update Forums set PostNum = PostNum + 1
			where Id = (
				select F.Id
				from Threads t 
					join Forums F on t.Forum = F.Slug
				where new.Thread = t.Id
			);	
		update Status set PostNum = PostNum + 1;
		return new;
	end;
$postNumInc$ language plpgsql;

create or replace function threadNumInc() returns trigger as $threadNumInc$ 
	begin
	    update Forums set ThreadNum = ThreadNum + 1
	    	where Slug = new.forum;
	    update Status set ThreadNum = ThreadNum + 1;
	    return new;
	end;
$threadNumInc$ language plpgsql;

create or replace function forumNumInc() returns trigger as $forumNumInc$
	begin 
   		update Status set ForumNum = ForumNum + 1;
   		return new;
	end;
$forumNumInc$ language plpgsql;

create or replace function threadRatingCount() returns trigger as $threadRatingCount$
	begin
		update Threads set VoteNum = VoteNum + new.Voice
	    	where Id = new.Thread;
		return new;
	end;
$threadRatingCount$ language plpgsql;

create or replace function threadRatingRecount() returns trigger as $threadRatingCount$
	begin
	    if new.Voice = old.Voice then
			return new;
		end if;
	    
		update Threads set VoteNum = VoteNum + new.Voice - old.Voice
	    	where Id = new.Thread;
		return new;
	end;
$threadRatingCount$ language plpgsql;

create or replace function userNumInc() returns trigger as $userNumInc$
	begin
		update Status set UserNum = UserNum + 1;
		return new;
	end;
$userNumInc$ language plpgsql;

create or replace function postsSetIdCheckForum() returns trigger as $postsSetId$
	begin
	    if new.Parent is not null then
	        if new.Thread != (select t.Id from Threads t join Posts P on t.Id = P.Thread where P.Id = new.Parent) then
				raise EXCEPTION 'Parent post was created in another thread';
			end if;
		end if;
	    
		return new;
	end;
$postsSetId$ language plpgsql;

create trigger postsSetId before insert on Posts for each row execute procedure postsSetIdCheckForum();
create trigger postNumInc after insert on Posts for each row execute procedure postNumInc();
create trigger threadNumInc after insert on	Threads for each row execute procedure threadNumInc();
create trigger threadRatingCount after insert on Votes for each row execute procedure threadRatingCount();
create trigger threadRatingRecount after update on Votes for each row execute procedure threadRatingRecount();
create trigger setPostIsEdited before update on Posts for each row execute procedure setPostIsEdited();
create trigger forumNumInc after insert on Forums for each row execute procedure forumNumInc();
create trigger userNumInc after insert on Users for each row execute procedure userNumInc();
`)
	if err != nil {
		panic(err)
	}
}
