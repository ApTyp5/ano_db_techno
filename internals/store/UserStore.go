package store

import (
	"database/sql"
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/ApTyp5/new_db_techno/logs"
	"github.com/pkg/errors"
	"strconv"
)

type UserStore interface {
	SelectByForum(users *[]*models.User, forum *models.Forum, limit int, since string, desc bool) error // forum.GetUsers
	Insert(user *models.User) error                                                                     // Create
	SelectByNickname(user *models.User) error                                                           // Get
	UpdateByNickname(user *models.User) error                                                           // Update
	SelectByNickNameOrEmail(users *[]*models.User) error
}

type PSQLUserStore struct {
	db *sql.DB
}

func CreatePSQLUserStore(db *sql.DB) UserStore {
	return PSQLUserStore{db: db}
}

func (P PSQLUserStore) SelectByForum(users *[]*models.User, forum *models.Forum, limit int, since string, desc bool) error {
	dsc := "ASC"
	if desc {
		dsc = "DESC"
	}

	sinc := ""
	if since != "" {
		if desc {
			sinc = " and u.nick_name < $2"
		} else {
			sinc = " and u.nick_name > $2"
		}
	}

	lmt := ""
	if limit > 0 {
		lmt = " limit " + strconv.FormatInt(int64(limit), 10)
	}

	query := `
		select distinct on (nick_name) u.About, u.Email, u.full_name, u.nick_name
		from Forums f 
			left join Threads t on f.Slug = t.Forum
			left join Posts p on t.Id = p.Thread
			join Users u on (u.nick_name = p.Author or u.nick_name = t.Author)
		where f.Slug = $1 ` + sinc + `
		order by u.nick_name ` + dsc + lmt + ";"

	var (
		rows *sql.Rows
		err  error
	)

	if since == "" {
		rows, err = P.db.Query(query, forum.Slug)
	} else {
		rows, err = P.db.Query(query, forum.Slug, since)
	}

	if err != nil {
		return errors.Wrap(err, "PSQLUserStore SelectByForum query")
	}

	defer rows.Close()

	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.About, &user.Email, &user.FullName, &user.NickName); err != nil {
			return errors.Wrap(err, "PSQLUserStore SelectByForum query scan")
		}
		*users = append(*users, user)
	}

	return nil
}

func (P PSQLUserStore) Insert(user *models.User) error {
	_, err := P.db.Exec(`
		insert into Users (About, Email, full_name, nick_name)
		values ($1, $2, $3, $4);
`, user.About, user.Email, user.FullName, user.NickName)

	return errors.Wrap(err, "PSQLUserStore insert err")
}

func (P PSQLUserStore) SelectByNickname(user *models.User) error {
	row := P.db.QueryRow(`
		select About, Email, full_name, nick_name
		from Users
		where nick_name = $1;
`, user.NickName)

	return errors.Wrap(row.Scan(&user.About, &user.Email, &user.FullName, &user.NickName),
		"PSQLUserStore selectByNickName")
}

func (P PSQLUserStore) UpdateByNickname(user *models.User) error {
	row := P.db.QueryRow(`
		update Users 
			set 
			    About = coalesce(nullif($1, ''), About), 
			    Email = coalesce(nullif($2, ''), Email), 
			    full_name = coalesce(nullif($3, ''), full_name)
		where nick_name = $4
		returning About, Email, full_name, nick_name;
`, user.About, user.Email, user.FullName, user.NickName)

	return errors.Wrap(row.Scan(&user.About, &user.Email, &user.FullName, &user.NickName), "PSQLUserStore updateByNickName")
}

func (P PSQLUserStore) SelectByNickNameOrEmail(users *[]*models.User) error {
	logs.Info("ENTRY: ", "EMAIL: ", (*users)[0].Email)
	logs.Info("ENTRY: ", "Nick: ", (*users)[0].NickName)
	rows, err := P.db.Query(`
		select About, Email, full_name, nick_name
		from Users
		where email = $1 or nick_name = $2;
`, (*users)[0].Email, (*users)[0].NickName)

	if err != nil {
		return errors.Wrap(err, "PSQLUserStore SelectByNickNameOrEmail query")
	}

	rows.Next()
	user := &models.User{}
	if err := rows.Scan(&user.About, &user.Email, &user.FullName, &user.NickName); err != nil {
		return errors.Wrap(err, "PSQLUserStore SelectByNickNameOrEmail scan1")
	}
	(*users)[0] = user

	logs.Info("HORAY")
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.About, &user.Email, &user.FullName, &user.NickName); err != nil {
			return errors.Wrap(err, "PSQLUserStore SelectByNickNameOrEmail scan2")
		}
		*users = append(*users, user)
	}
	logs.Info("HORAY")

	return nil
}
