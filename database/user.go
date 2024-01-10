package database

type User struct {
	Id int64
	Name string
	Email string
	Password string
	CreatedAt int64
	IsValidated bool
}

func (m *Model) NewUser(name string, email string, password string) (userId int64, err error) {
	result, err := m.db.Exec("INSERT INTO usuario(nome,email,senha) VALUES(?,?,?)",name,email,password)
	if err != nil {
		return 0, err
	}
	userId, err = result.LastInsertId()
	return userId, err
}
func (m *Model) GetUserById(id int) (user User, err error) {
	var u User
	err = m.db.QueryRow("SELECT id, nome, email, senha, validado, criado FROM usuario WHERE id = ?",id).
			Scan(&u.Id,&u.Name,&u.Email,&u.Password,&u.IsValidated,&u.CreatedAt)

	return u, err
}
func (m *Model) GetUserByEmail(email string) (user User, err error) {
	var u User
	err = m.db.QueryRow("SELECT id, nome, email, senha, validado, criado FROM usuario WHERE email = ?",email).
			Scan(&u.Id,&u.Name,&u.Email,&u.Password,&u.IsValidated,&u.CreatedAt)

	return u, err
}
func (m *Model) IsUserEmailTaken(email string) bool {
	var isIt bool
	_ = m.db.QueryRow("SELECT COUNT(1) FROM usuario WHERE email = ?", email).Scan(&isIt)
	return isIt
}