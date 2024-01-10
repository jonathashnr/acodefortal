package database

import (
	"time"

	"github.com/google/uuid"
)

func (m *Model) NewSession(userId int64) (token string, err error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	token = uuid.String()
	expires := time.Now().Add(900 * time.Second).Unix()
	_, err = m.db.Exec("INSERT INTO sessao(chave,usuario_id,expira) VALUES(?,?,?)", token, userId, expires)
	if err != nil {
		return "", err
	}
	return token, err
}

// func (m *model) IsSessionActive(token string) bool {
// 	var isIt bool
// 	_ = m.db.QueryRow("SELECT COUNT(1) FROM sessao WHERE chave = ? AND expira >= ?", token,time.Now().Unix()).Scan(&isIt)
// 	return isIt
// }

func (m *Model) ProlongSession(token string) error {
	expires := time.Now().Add(900 * time.Second).Unix()
	_, err := m.db.Exec("UPDATE sessao SET expira = ? WHERE chave = ?", expires, token)
	return err
}

func (m *Model) GetUserFromActiveSession(token string) (User, error) {
	var u User
	stm, err := m.db.Prepare("SELECT id, nome, email, senha, permissao, criado FROM usuario INNER JOIN sessao WHERE id = usuario_id AND chave = ? and expira >= ?")
	if err != nil {
		return u, err
	}
	defer stm.Close()
	err = stm.QueryRow(token, time.Now().Unix()).
			Scan(&u.Id,&u.Name,&u.Email,&u.Password,&u.Permission,&u.CreatedAt)

	return u, err
}