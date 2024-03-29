package models

type User struct {
	Id int64
	Name string
	Email string
	Password string
	CreatedAt int64
	Permission int
}

func (s *Store) NewUser(name string, email string, password string) (userId int64, err error) {
	result, err := s.db.Exec("INSERT INTO usuario(nome,email,senha) VALUES(?,?,?)",name,email,password)
	if err != nil {
		return 0, err
	}
	userId, err = result.LastInsertId()
	return userId, err
}

func (s *Store) GetUserById(id int) (user User, err error) {
	var u User
	err = s.db.QueryRow("SELECT id, nome, email, senha, permissao, criado FROM usuario WHERE id = ?",id).
			Scan(&u.Id,&u.Name,&u.Email,&u.Password,&u.Permission,&u.CreatedAt)

	return u, err
}

func (s *Store) GetUserByEmail(email string) (user User, err error) {
	var u User
	err = s.db.QueryRow("SELECT id, nome, email, senha, permissao, criado FROM usuario WHERE email = ?",email).
			Scan(&u.Id,&u.Name,&u.Email,&u.Password,&u.Permission,&u.CreatedAt)

	return u, err
}

func (s *Store) IsUserEmailTaken(email string) bool {
	var isIt bool
	_ = s.db.QueryRow("SELECT COUNT(1) FROM usuario WHERE email = ?", email).Scan(&isIt)
	return isIt
}