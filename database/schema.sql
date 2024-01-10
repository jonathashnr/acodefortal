BEGIN;
CREATE TABLE usuario(
    id INTEGER PRIMARY KEY,
    nome TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    senha TEXT NOT NULL,
    validado INTEGER DEFAULT (0) NOT NULL
    criado INTEGER DEFAULT (unixepoch()) NOT NULL
);

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

CREATE TABLE categoria(
    id INTEGER PRIMARY KEY,
    nome TEXT UNIQUE NOT NULL
);

INSERT INTO categoria(nome) VALUES('Ajuda Humanitária');
INSERT INTO categoria(nome) VALUES('Animais');
INSERT INTO categoria(nome) VALUES('Cultura e Arte');
INSERT INTO categoria(nome) VALUES('Desenvolvimento Comunitário');
INSERT INTO categoria(nome) VALUES('Direitos Humanos');
INSERT INTO categoria(nome) VALUES('Educação');
INSERT INTO categoria(nome) VALUES('Meio Ambiente');
INSERT INTO categoria(nome) VALUES('Saúde');

CREATE TABLE org(
    id INTEGER PRIMARY KEY,
    nome TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    categoria_id INTEGER NOT NULL,
    missao TEXT NOT NULL,
    criado INTEGER DEFAULT (unixepoch()) NOT NULL,
    modificado INTEGER DEFAULT (unixepoch()) NOT NULL,
    FOREIGN KEY(categoria_id) REFERENCES categoria(id)
);

CREATE TABLE endereco(
    id INTEGER PRIMARY KEY,
    org_id INTEGER NOT NULL UNIQUE,
    logradouro TEXT NOT NULL,
    numero TEXT NOT NULL,
    bairro TEXT NOT NULL,
    cidade TEXT NOT NULL,
    uf TEXT NOT NULL,
    cep TEXT NOT NULL,
    complemento TEXT,
    FOREIGN KEY(org_id) REFERENCES org(id)
);

CREATE TABLE contato(
    id INTEGER PRIMARY KEY,
    org_id INTEGER NOT NULL UNIQUE,
    email TEXT,
    telefone TEXT,
    website TEXT,
    instagram TEXT,
    facebook TEXT,
    twitter TEXT,
    whatsapp TEXT,
    FOREIGN KEY(org_id) REFERENCES org(id)
);

CREATE TABLE org_imagem(
    id INTEGER PRIMARY KEY,
    org_id INTEGER NOT NULL,
    imagem BLOB NOT NULL,
    FOREIGN KEY(org_id) REFERENCES org(id)
);

COMMIT;