package mysql

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"operators.kloudlite.io/lib/errors"
	"strings"
)

type Client struct {
	db   *sql.DB
	ctx  context.Context
	conn *sql.Conn
}

func (c *Client) Connect(ctx context.Context) error {
	if err := c.db.PingContext(ctx); err != nil {
		return err
	}
	conn, err := c.db.Conn(ctx)
	if err != nil {
		return err
	}
	c.ctx = ctx
	c.conn = conn
	return nil
}

func (c *Client) sanitizeDbName(dbname string) string {
	return strings.ReplaceAll(dbname, "-", "_")
}

func (c *Client) Disconnect() error {
	return c.db.Close()
}

func (c *Client) UpsertUser(dbName, username, password string) error {
	defer func() {
		err := recover()
		fmt.Println("ERR:", err)
	}()
	// result, err := c.conn.ExecContext(c.ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName))
	result, err := c.conn.ExecContext(
		context.Background(), fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", c.sanitizeDbName(dbName)),
	)
	fmt.Println("result:", result)
	if err != nil {
		return errors.NewEf(err, "creating database")
	}

	_, err = c.conn.ExecContext(
		c.ctx, fmt.Sprintf(
			"CREATE USER IF NOT EXISTS '%s'@'%%' IDENTIFIED BY '%s'", username,
			password,
		),
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
	rows, err := c.conn.QueryContext(
		c.ctx,
		fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM mysql.user WHERE user = '%s')", username),
	)
	if err != nil {
		return false, errors.NewEf(err, "dropping database")
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
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", username, password, hosts, dbName))
	if err != nil {
		return nil, err
	}
	// See "Important settings" section.
	// db.SetConnMaxLifetime(time.Minute * 3)
	// db.SetMaxOpenConns(10)
	// db.SetMaxIdleConns(10)
	return &Client{db: db}, nil
}
