BEGIN;
CREATE TABLE usuario(
    id INTEGER PRIMARY KEY,
    nome TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    senha TEXT NOT NULL,
    criado INTEGER DEFAULT (unixepoch()) NOT NULL
);
-- falta tabela de perfil
CREATE TABLE admin(
    id INTEGER PRIMARY KEY,
    usuario_id INTEGER NOT NULL UNIQUE,
    nivel INTEGER NOT NULL,
    FOREIGN KEY(usuario_id) REFERENCES usuario(id) 
);

CREATE TABLE sessao(
    chave TEXT PRIMARY KEY,
    usuario_id INTEGER NOT NULL,
    expira INTEGER NOT NULL,
    FOREIGN KEY(usuario_id) REFERENCES usuario(id)
);
COMMIT;