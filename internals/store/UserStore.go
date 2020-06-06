package store

import (
	"database/sql"
	"github.com/ApTyp5/new_db_techno/internals/models"
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

	sne := ""
	if since != "" {
		if desc {
			sne = " and u.NickName < $2"
		} else {
			sne = " and u.NickName > $2"
		}
	}

	lmt := ""
	if limit > 0 {
		lmt = " limit " + strconv.FormatInt(int64(limit), 10)
	}

	query := `
		select distinct on (NickName) u.About, u.Email, u.FullName, u.NickName
		from Forums f 
			left join Threads t on f.Slug = t.Forum
			left join Posts p on t.Id = p.Thread
			join Users u on (u.Id = p.Author or u.Id = t.Author)
		where f.Slug = $1 ` + sne + `
		order by u.NickName ` + dsc + lmt + ";"

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
		insert into Users (About, Email, FullName, NickName)
		values ($1, $2, $3, $4);
`, user.About, user.Email, user.FullName, user.NickName)

	return errors.Wrap(err, "PSQLUserStore insert err")
}

func (P PSQLUserStore) SelectByNickname(user *models.User) error {
	row := P.db.QueryRow(`
		select About, Email, FullName, NickName
		from Users
		where NickName = $1;
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
			    FullName = coalesce(nullif($3, ''), FullName)
		where NickName = $4
		returning About, Email, FullName;
`, user.About, user.Email, user.FullName, user.NickName)

	return errors.Wrap(row.Scan(&user.About, &user.Email, &user.FullName), "PSQLUserStore updateByNickName")
}

func (P PSQLUserStore) SelectByNickNameOrEmail(users *[]*models.User) error {
	rows, err := P.db.Query(`
		select About, Email, FullName, NickName
		from Users
		where NickName = $1 or Email = $2
`, (*users)[0].NickName, (*users)[0].Email)

	if err != nil {
		return errors.Wrap(err, "PSQLUserStore SelectByNickNameOrEmail query")
	}

	rows.Next()
	user := &models.User{}
	if err := rows.Scan(&user.About, &user.Email, &user.FullName, &user.NickName); err != nil {
		return errors.Wrap(err, "PSQLUserStore SelectByNickNameOrEmail scan1")
	}
	(*users)[0] = user

	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.About, &user.Email, &user.FullName, &user.NickName); err != nil {
			return errors.Wrap(err, "PSQLUserStore SelectByNickNameOrEmail scan2")
		}
		*users = append(*users, user)
	}

	return rows.Close()
}
