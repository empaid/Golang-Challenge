package queries

import (
	"context"
	"time"

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
	MessageCount       int         `db:"message_count"`
	Password           []byte      `db:"password"`
}

type GetUserMessagesQueryRow struct {
	ID        int       `db:"id"`
	UserId    int       `db:"user_id"`
	Message   string    `db:"message"`
	CreatedAt time.Time `db:"created_at"`
}

type UserDBStore interface {
	GetUsers(context.Context) ([]GetUsersQueryRow, error)
	CreateUser(context.Context, string, string, string, *string, []byte) (*GetUsersQueryRow, error)
	PatchUser(context.Context, int, *string, *string, *string, *string, bool) (*GetUsersQueryRow, error)
	CreateUserMessage(context.Context, int, string) (*GetUserMessagesQueryRow, error)
	FetchUserFromEmail(context.Context, string) (*GetUsersQueryRow, error)
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
			utype.permission_bitfield::text as "permission_bitfield",
			(SELECT COUNT(*) FROM public.user_messages WHERE public.user_messages.user_id = public.users.id) as "message_count"
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
			&user.MessageCount,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, err
}

func (db *UserDB) CreateUser(ctx context.Context, username, email, userType string, nickname *string, password []byte) (*GetUsersQueryRow, error) {
	var user GetUsersQueryRow
	err := db.conn.QueryRow(ctx, `
		INSERT INTO public.users (username, nickname, email, user_type, password) VALUES ($1, $2, $3, $4, $5) returning id, username, nickname, email, user_type`,
		username, nickname, email, userType, password).Scan(
		&user.ID,
		&user.Username,
		&user.Nickname,
		&user.Email,
		&user.UserType,
	)

	if err != nil {
		return nil, err
	}

	return &user, err
}

func (db *UserDB) PatchUser(ctx context.Context, userId int, username, email, userType, nickname *string, nicknameProvided bool) (*GetUsersQueryRow, error) {
	var user GetUsersQueryRow

	err := db.conn.QueryRow(ctx, `
		UPDATE public.users
      	SET
        username  = COALESCE($1, username),
        email     = COALESCE($2, email),
        user_type = COALESCE($3, user_type),
        nickname  = CASE WHEN $5 THEN $4 ELSE nickname END
      	WHERE id = $6
      	RETURNING id, username, nickname, email, user_type
		`, username, email, userType, nickname, nicknameProvided, userId).Scan(
		&user.ID,
		&user.Username,
		&user.Nickname,
		&user.Email,
		&user.UserType,
	)

	if err != nil {
		return nil, err
	}

	return &user, err
}

func (db *UserDB) CreateUserMessage(ctx context.Context, userId int, message string) (*GetUserMessagesQueryRow, error) {
	var userMessage GetUserMessagesQueryRow
	err := db.conn.QueryRow(ctx, `
		INSERT INTO public.user_messages (user_id, message) VALUES ($1, $2) returning id, user_id, message, created_at`,
		userId, message).Scan(
		&userMessage.ID,
		&userMessage.UserId,
		&userMessage.Message,
		&userMessage.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &userMessage, err
}

func (db *UserDB) FetchUserFromEmail(ctx context.Context, email string) (*GetUsersQueryRow, error) {
	var user GetUsersQueryRow
	err := db.conn.QueryRow(ctx, `
		SELECT id, username, nickname, email, user_type, password FROM public.users WHERE email=$1;`,
		email).Scan(
		&user.ID,
		&user.Username,
		&user.Nickname,
		&user.Email,
		&user.UserType,
		&user.Password,
	)

	if err != nil {
		return nil, err
	}

	return &user, err
}
