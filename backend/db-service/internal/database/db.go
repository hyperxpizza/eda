package database

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/lib/pq"

	"github.com/hyperxpizza/eda/backend/db-service/internal/config"
	"github.com/sirupsen/logrus"
)

type PostgresClient struct {
	db    *sql.DB
	mutex sync.Mutex
	lgr   logrus.FieldLogger
	cfg   *config.PostgresConfig
}

func NewPostgresClient(cfg *config.PostgresConfig, lgr logrus.FieldLogger) (*PostgresClient, error) {
	psqlInfo := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	lgr.Info("connected to postgres")

	return &PostgresClient{
		db:    db,
		mutex: sync.Mutex{},
		lgr:   lgr.WithField("module", "postgres-client"),
		cfg:   cfg,
	}, nil
}

func (c *PostgresClient) InsertMessage(msg string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, err := c.db.Exec(`insert into messages(content) values ($1)`, msg)
	if err != nil {
		c.lgr.Errorf("cou")
		return err
	}

	return nil
}
