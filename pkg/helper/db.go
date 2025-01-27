package helper

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type TLS struct {
	ClientCert string
	ClientKey  string
	ServerCA   string
	ServerName string
}

type Config struct {
	Username string
	Password string
	Hostname string
	Port     string
	Database string
	TLS      TLS
}

func Open(cfg Config) (*gorm.DB, error) {
	connectString := ConnectString(cfg)

	db, err := gorm.Open("postgres", connectString)
	if err != nil {
		return nil, err
	}
	db.LogMode(false)
	return db, nil
}

func ConnectString(cfg Config) string {
	var str string

	str = fmt.Sprintf(`host=%v port=%v user=%v dbname=%v password=%v sslmode=require`,
		cfg.Hostname,
		cfg.Port,
		cfg.Username,
		cfg.Database,
		cfg.Password,
	)

	return str
}
