package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"operators.kloudlite.io/lib/errors"
)

type Client struct {
	db   *sql.DB
	ctx  context.Context
	conn *sql.Conn
}

func (c *Client) Connect(ctx context.Context) error {
	if err := c.db.PingContext(ctx); err != nil {
		return errors.NewEf(err, "could not ping with mysql connection")
	}
	conn, err := c.db.Conn(ctx)
	if err != nil {
		return errors.NewEf(err, "could not create db connection")
	}
	c.ctx = ctx
	c.conn = conn
	return nil
}

func (c *Client) sanitizeDbName(dbname string) string {
	return strings.ReplaceAll(dbname, "-", "_")
}

func (c *Client) Close() error {
	return c.db.Close()
}

func (c *Client) UpsertUser(dbName, username, password string) error {
	if c.conn == nil {
		return errors.Newf("please connect to mysql prior to calling UpsertUser")
	}

	if err := c.Connect(context.Background()); err != nil {
		return err
	}

	_, err := c.conn.ExecContext(context.Background(), fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName))
	if err != nil {
		return errors.NewEf(err, "creating database")
	}

	_, err = c.conn.ExecContext(
		c.ctx, fmt.Sprintf("CREATE USER IF NOT EXISTS '%s'@'%%' IDENTIFIED BY '%s'", username, password),
	)
	if err != nil {
		return errors.NewEf(err, "creating user")
	}

	_, err = c.conn.ExecContext(
		c.ctx, fmt.Sprintf(
			"GRANT ALL PRIVILEGES ON %s.* TO '%s'@'%%' ", c.sanitizeDbName(dbName),
			username,
		),
	)
	if err != nil {
		return errors.NewEf(err, "granting privileges")
	}
	_, err = c.conn.ExecContext(c.ctx, fmt.Sprintf("FLUSH PRIVILEGES"))
	if err != nil {
		return errors.NewEf(err, "flusing privileges")
	}

	return nil
}

func (c *Client) DropUser(username string) error {
	_, err := c.conn.ExecContext(c.ctx, fmt.Sprintf("DROP USER '%s'@'%%'", username))
	if err != nil {
		return errors.NewEf(err, "dropping user")
	}
	return nil
}

func (c *Client) DropDatabase(dbName string) error {
	_, err := c.conn.ExecContext(c.ctx, fmt.Sprintf("DROP DATABASE %s", c.sanitizeDbName(dbName)))
	if err != nil {
		return errors.NewEf(err, "dropping database")
	}
	return nil
}

func (c *Client) UserExists(username string) (bool, error) {
	if c.conn == nil {
		return false, errors.Newf("please connect to mysql prior to calling UserExists")
	}

	rows, err := c.conn.QueryContext(
		c.ctx,
		fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM mysql.user WHERE user = '%s')", username),
	)
	if err != nil {
		return false, err
	}
	count := 0
	if rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return false, errors.NewEf(err, "could not scan into *int")
		}
	}
	return count == 1, nil
}

func NewClient(hosts, dbName, username, password string) (*Client, error) {
	connector, err := mysql.NewConnector(
		&mysql.Config{
			User:                 username,
			Passwd:               password,
			Addr:                 hosts,
			DBName:               dbName,
			Collation:            "utf8mb4_general_ci",
			AllowNativePasswords: true,
		},
	)
	if err != nil {
		return nil, err
	}
	db := sql.OpenDB(connector)
	// See "Important settings" section.
	db.SetConnMaxLifetime(20 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return &Client{db: db}, nil
}
