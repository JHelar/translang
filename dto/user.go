package dto

import "translang/db"

type User struct {
	ID int64 `db:"id"`
}

type PasswordUser struct {
	User
	Email        string `db:"email"`
	PasswordHash string `db:"password_hash"`
}

const SELECT_USER_BY_PASSWORD_PROVIDER_EMAIL = `
select user.id as id,password_provider.email as email,password_provider.password_hash as password_hash
	from user
	left join password_provider
		on password_provider.user_id=user.id
	where password_provider.email=$1

`

func GetPasswordUserByEmail(email string, db *db.DBClient) (PasswordUser, error) {
	user := PasswordUser{}
	if err := db.DB.Get(&user, SELECT_USER_BY_PASSWORD_PROVIDER_EMAIL, email); err != nil {
		return PasswordUser{}, err
	}
	return user, nil
}
