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
	// Version
	Version string = "0.0.6"
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
	Channel int    `toml:"channel"`
}

// Run ...
func (c *CLI) Run(args []string) int {
	var configPath string

	app := cli.NewApp()
	app.Name = "epgrec-program-finder"
	app.Version = Version
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
	Channel   int       `db:"channel"`
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
	sql := "SELECT rp.title, rp.channel, rp.starttime FROM Recorder_programTbl AS rp "
	sql += "LEFT OUTER JOIN Recorder_reserveTbl AS rr ON rr.program_id = rp.id "
	sql += "WHERE (rp.category_id = :category_id OR rp.genre2 = :category_id OR rp.genre3 = :category_id) "
	sql += "AND rp.starttime > :starttime "
	sql += "AND rr.id IS NULL "
	for k, e := range cnf.Excludes {
		condition := ""
		if len(e.Keyword) > 0 {
			keyword := fmt.Sprintf("title%d", k)
			args[keyword] = e.Keyword
			condition += fmt.Sprintf("rp.title LIKE :%s ", keyword)
		}
		if len(condition) > 0 && e.Channel > 0 {
			channel := fmt.Sprintf("channel%d", k)
			args[channel] = e.Channel
			condition += fmt.Sprintf("AND rp.channel = :%s ", channel)
		}
		if len(condition) > 0 {
			sql += fmt.Sprintf("AND NOT (%s) ", condition)
		}
	}
	nstmt, err := db.PrepareNamed(sql)
	err = nstmt.Select(&pgs, args)
	if err != nil {
		log.Fatalln(err)
	}

	if len(pgs) > 0 {
		for _, pg := range pgs {
			fmt.Printf("%s %02dch %s\n", pg.StartTime.Format("2006-01-02 15:04:05"), pg.Channel, pg.Title)
		}
	} else {
		fmt.Println("Not found  programs.")
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
