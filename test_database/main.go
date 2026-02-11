package main

import (
	"fmt"

	"github.com/AliTaghipour1/Talk-to_DB/internal/modules/db"
)

func main() {
	database, err := db.NewDatabase(db.DatabaseConfig{
		Host:     "localhost",
		Port:     "55007",
		User:     "cockroach",
		Password: "cockroach",
		Database: "cockroach",
	}, db.Cockroach)
	if err != nil {
		return
	}
	initializeCockroach(database)

}

func initializeCockroach(database db.Database) {
	_, err := database.Query(`CREATE SEQUENCE IF NOT EXISTS payments_id_seq START 1 MAXVALUE 4294967296;`)
	if err != nil {
		panic(fmt.Errorf("payments_id_seq failed [%v]", err))
	}

	_, err = database.Query(`CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    first_name TEXT NOT NULL DEFAULT '',
    last_name TEXT NOT NULL DEFAULT '',
    nickname TEXT,
    genre SMALLINT NOT NULL DEFAULT '0',
	created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);`)
	if err != nil {
		panic(fmt.Errorf("create table users failed [%v]", err))
	}

	_, err = database.Query(`CREATE TABLE IF NOT EXISTS user_account (
    id INTEGER PRIMARY KEY,
    owner_user_id INTEGER NOT NULL,
    account_number TEXT NOT NULL DEFAULT '',
    account_name TEXT,
    balance BIGINT NOT NULL DEFAULT 0,
	created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (owner_user_id, account_number)
);`)
	if err != nil {
		panic(fmt.Errorf("create table user_account failed [%v]", err))
	}

	_, err = database.Query(`CREATE TABLE IF NOT EXISTS user_payments (
    payment_id INTEGER PRIMARY KEY DEFAULT nextval('payments_id_seq'),
    amount BIGINT NOT NULL DEFAULT 0,
    fee BIGINT NOT NULL DEFAULT 0,
    account_id INTEGER NOT NULL,
	created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);`)
	if err != nil {
		panic(fmt.Errorf("create table user_payments failed [%v]", err))
	}

	_, err = database.Query(`INSERT INTO users (id,first_name,last_name,genre,nickname,created_at) VALUES 
(1,'Ali','Taghipour',1,'atp1380',now()-interval '35d'),
(2,'Mahdieh','Shariat',2,'mshariat',now()-interval '10d'),
(3,'Hesam','Soleymani',1,'bellinghesam',now()-interval '3d')`)
	if err != nil {
		panic(fmt.Errorf("insert data to table users with nick failed [%v]", err))
	}

	_, err = database.Query(`INSERT INTO users (id,first_name,last_name,genre,created_at) VALUES 
(4,'Nargess','Dehghani',2,now()-interval '1d'),
(5,'Hamid','Beigy',1,now()-interval '2h')`)
	if err != nil {
		panic(fmt.Errorf("insert data to table users without nick failed [%v]", err))
	}

	_, err = database.Query(`INSERT INTO user_account (id,owner_user_id,account_number,account_name,balance,created_at) VALUES 
(1,1,'6037998210432286','daei jan',32000000,now()-interval '23d'),
(2,1,'5892101410421075','tashakor',12001000,now()-interval '16d'),
(3,2,'6037998217437216','katooni',820010000,now()-interval '9d'),
(4,3,'6221061226421075','salam',213456000,now()-interval '3d'),
(5,4,'6037998297437468','coffee',1820010000,now()-interval '1d'),
(6,5,'1111998297437468','project',2000010000,now()-interval '1h')`)
	if err != nil {
		panic(fmt.Errorf("insert data to table users with account name failed [%v]", err))
	}

	_, err = database.Query(`INSERT INTO user_payments (amount,fee,account_id,created_at) VALUES 
(10000,100,1,now()-interval '16d'),
(68000,500,1,now()-interval '22d'),
(1000000,10000,2,now()-interval '5d'),
(520000,10000,3,now()-interval '2d'),
(7800000,78000,3,now()-interval '3d'),
(12300000,123000,3,now()-interval '1d'),
(100987,1098,4,now()-interval '18h'),
(2000000,20000,5,now()-interval '35m'),
(1000000,10000,6,now()-interval '10m')`)
	if err != nil {
		panic(fmt.Errorf("insert data to table users without account name failed [%v]", err))
	}
}
