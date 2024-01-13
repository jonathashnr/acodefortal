package models

import (
	"time"

	"github.com/google/uuid"
)

func (s *Store) NewSession(userId int64) (token string, err error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	token = uuid.String()
	expires := time.Now().Add(900 * time.Second).Unix()
	_, err = s.db.Exec("INSERT INTO sessao(chave,usuario_id,expira) VALUES(?,?,?)", token, userId, expires)
	if err != nil {
		return "", err
	}
	return token, err
}

// func (m *Store) IsSessionActive(token string) bool {
// 	var isIt bool
// 	_ = m.db.QueryRow("SELECT COUNT(1) FROM sessao WHERE chave = ? AND expira >= ?", token,time.Now().Unix()).Scan(&isIt)
// 	return isIt
// }

func (s *Store) ProlongSession(token string) error {
	expires := time.Now().Add(900 * time.Second).Unix()
	_, err := s.db.Exec("UPDATE sessao SET expira = ? WHERE chave = ?", expires, token)
	return err
}

func (s *Store) GetUserFromActiveSession(token string) (User, error) {
	var u User
	stm, err := s.db.Prepare("SELECT id, nome, email, senha, permissao, criado FROM usuario INNER JOIN sessao WHERE id = usuario_id AND chave = ? and expira >= ?")
	if err != nil {
		return u, err
	}
	defer stm.Close()
	err = stm.QueryRow(token, time.Now().Unix()).
			Scan(&u.Id,&u.Name,&u.Email,&u.Password,&u.Permission,&u.CreatedAt)

	return u, err
}

func (m *Store) RemoveSession(token string) error {
	_, err := m.db.Exec("DELETE FROM sessao WHERE chave = ?", token)
	return err
}