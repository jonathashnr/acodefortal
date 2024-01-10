package database

import (
	"time"

	"github.com/google/uuid"
)

func (m *Model) NewSession(userId int) (token string, err error) {
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

func (m *Model) GetUserIdFromActiveSession(token string) (int, error) {
	var userId int
	err := m.db.QueryRow("SELECT usuario_id FROM sessao WHERE chave = ? AND expira >= ?", token,time.Now().Unix()).Scan(&userId)
	return userId, err
}