package repositories

import (
	"github.com/ApTyp5/new_db_techno/internals/models"
	"github.com/jackc/pgx"
)

type UserRepo interface {
	SelectByForum(users *[]models.User, forum *models.Forum, limit int, since string, desc bool) error // forum.GetUsers
	Insert(user *models.User) error                                                                    // Create
	SelectByNickname(user *models.User) error                                                          // Get
	UpdateByNickname(user *models.User) error                                                          // Update
	SelectByNickNameOrEmail(users *[]models.User, user *models.User) error
}

type PSQLUserRepo struct {
	db                  *pgx.ConnPool
	selectByNickOrEmail *pgx.PreparedStatement
	insert              *pgx.PreparedStatement
	selectByNick        *pgx.PreparedStatement
	updateByNick        *pgx.PreparedStatement
}

func CreatePSQLUserRepo(db *pgx.ConnPool) UserRepo {
	var err error
	prefix := "user_"
	repo := PSQLUserRepo{db: db}

	repo.updateByNick, err = db.Prepare(prefix+"updateByNick", `
		update Users 
			set 
			    About = coalesce(nullif($1, ''), About), 
			    Email = coalesce(nullif($2, ''), Email), 
			    full_name = coalesce(nullif($3, ''), full_name)
		where nick_name = $4
		returning About, Email, full_name, nick_name;
	`)
	panicIfErr(err)

	repo.selectByNickOrEmail, err = db.Prepare(prefix+"selectByNickOrEmail", `
	SELECT about, email, full_name, nick_name
		FROM users
		WHERE email = $1 or nick_name = $2;
	`)
	panicIfErr(err)

	repo.insert, err = db.Prepare(prefix+"insertStat", `
		INSERT INTO users (about, email, full_name, nick_name)
		VALUES ($1, $2, $3, $4);
	`)
	panicIfErr(err)

	repo.selectByNick, err = db.Prepare(prefix+"selectByNick", `
		select About, Email, full_name, nick_name
		from Users
		where nick_name = $1;
	`)
	panicIfErr(err)

	return repo
}

func (userRepo PSQLUserRepo) SelectByForum(users *[]models.User, forum *models.Forum, limit int, since string, desc bool) error {
	rows, err := userRepo.db.Query("SELECT about, email, full_name, nick_name "+
		"from select_users_by_forum($1, $2, $3, $4)", forum.Slug, desc, limit, since)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		i := len(*users)
		*users = append(*users, models.User{})
		if err := rows.Scan(&(*users)[i].About, &(*users)[i].Email, &(*users)[i].FullName, &(*users)[i].NickName); err != nil {
			return err
		}
	}

	return nil
}

func (userRepo PSQLUserRepo) Insert(user *models.User) error {
	_, err := userRepo.db.Exec(userRepo.insert.Name, user.About,
		user.Email, user.FullName, user.NickName)
	return err
}

func (userRepo PSQLUserRepo) SelectByNickname(user *models.User) error {
	row := userRepo.db.QueryRow(userRepo.selectByNick.Name, user.NickName)
	return row.Scan(&user.About, &user.Email, &user.FullName, &user.NickName)
}

func (userRepo PSQLUserRepo) UpdateByNickname(user *models.User) error {
	row := userRepo.db.QueryRow(userRepo.updateByNick.Name, user.About, user.Email, user.FullName, user.NickName)
	return row.Scan(&user.About, &user.Email, &user.FullName, &user.NickName)
}

func (userRepo PSQLUserRepo) SelectByNickNameOrEmail(users *[]models.User, user *models.User) error {
	rows, err := userRepo.db.Query(userRepo.selectByNickOrEmail.Name, user.Email, user.NickName)

	if err != nil {
		return err
	}

	for rows.Next() {
		i := len(*users)
		*users = append(*users, models.User{})
		if err := rows.Scan(&(*users)[i].About, &(*users)[i].Email, &(*users)[i].FullName, &(*users)[i].NickName); err != nil {
			return err
		}
	}

	return nil
}
