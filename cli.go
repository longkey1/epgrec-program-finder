package main

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/BurntSushi/toml"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/urfave/cli"
)

const (
	// ExitCodeOK ...
	ExitCodeOK int = 0
	// ExitCodeError ..
	ExitCodeError int = 1
)

// CLI ...
type CLI struct {
	outStream io.Writer
	errStream io.Writer
}

// Config ...
type Config struct {
	Database Database  `toml:"database"`
	Excludes []Exclude `toml:"excludes"`
}

// Database ...
type Database struct {
	Driver   string `toml:"driver"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	Host     string `toml:"host"`
	Port     string `toml:"port"`
	DBName   string `toml:"dbname"`
	Option   string `toml:"option"`
}

// DSN ...
func (d *Database) DSN() string {
	return fmt.Sprintf("%s:%s@(%s:%s)/%s?%s", d.Username, d.Password, d.Host, d.Port, d.DBName, d.Option)
}

// Exclude ...
type Exclude struct {
	Keyword string `toml:"keyword"`
}

// Run ...
func (c *CLI) Run(args []string) int {
	var configPath string

	app := cli.NewApp()
	app.Name = "epgrec-program-finder"
	app.Version = "0.0.1"
	app.Usage = "epgrec program finder"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config, c",
			Usage:       "Load configration from `FILE`",
			Destination: &configPath,
		},
		cli.StringFlag{
			Name:  "since, s",
			Value: time.Now().Format("2006-01-02 15:04:05"),
			Usage: "Since start time",
		},
	}
	app.Action = func(c *cli.Context) error {
		conf, err := loadConfig(configPath)
		if err != nil {
			return err
		}

		return find(c, conf)
	}

	err := app.Run(args)
	if err != nil {
		fmt.Fprintln(c.errStream, err)
		return ExitCodeError
	}

	return ExitCodeOK
}

// Program ...
type Program struct {
	Title     string    `db:"title"`
	StartTime time.Time `db:"starttime"`
}

func find(ctx *cli.Context, cnf *Config) error {
	db, err := sqlx.Connect(cnf.Database.Driver, cnf.Database.DSN())
	if err != nil {
		return err
	}

	pgs := []Program{}
	args := map[string]interface{}{
		"category_id": 8,
		"starttime":   ctx.String("since"),
	}
	sql := "SELECT rp.title, rp.starttime FROM Recorder_programTbl AS rp "
	sql += "LEFT OUTER JOIN Recorder_reserveTbl AS rr ON rr.program_id = rp.id "
	sql += "WHERE (rp.category_id = :category_id OR rp.genre2 = :category_id OR rp.genre3 = :category_id) "
	sql += "AND rp.starttime > :starttime "
	sql += "AND rr.id IS NULL "
	for k, e := range cnf.Excludes {
		key := fmt.Sprintf("title%d", k)
		args[key] = e.Keyword
		sql += fmt.Sprintf("AND (rp.title NOT LIKE :%s) ", key)
	}
	nstmt, err := db.PrepareNamed(sql)
	err = nstmt.Select(&pgs, args)
	if err != nil {
		log.Fatalln(err)
	}

	for _, pg := range pgs {
		fmt.Printf("%s %s\n", pg.StartTime.Format("2006-01-02 15:04:05"), pg.Title)
	}

	return nil
}

func loadConfig(path string) (*Config, error) {
	c := &Config{}
	if _, err := toml.DecodeFile(path, c); err != nil {
		return nil, err
	}

	return c, nil
}
