package queries

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type GetUsersQueryRow struct {
	ID                 int         `db:"id"`
	Username           string      `db:"username"`
	Email              string      `db:"email"`
	UserType           string      `db:"user_type"`
	Nickname           pgtype.Text `db:"nickname"`
	PermissionBitfield string      `db:"permission_bitfield"`
}

type UserDB struct {
	conn *pgx.Conn
}

func (db *UserDB) GetUsers(ctx context.Context) ([]GetUsersQueryRow, error) {
	// conn := GetConnection()
	// defer conn.Close(context.TODO())

	rows, err := db.conn.Query(ctx, `
		SELECT
			public.users.id, 
			public.users.username, 
			public.users.email,
			utype.type_key as "user_type",
			public.users.nickname,
			utype.permission_bitfield::text as "permission_bitfield"
		from 
			public.users 
		left join 
			user_types utype
		on
			utype.type_key = public.users.user_type
	`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []GetUsersQueryRow = []GetUsersQueryRow{}
	for rows.Next() {
		var user GetUsersQueryRow
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.UserType,
			&user.Nickname,
			&user.PermissionBitfield,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, err
}
