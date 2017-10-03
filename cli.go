package main

import (
	"fmt"
	"io"
	"log"
	"time"

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

// Run ...
func (c *CLI) Run(args []string) int {
	app := cli.NewApp()
	app.Name = "epgrec-program-finder"
	app.Version = "0.0.1"
	app.Usage = "epgrec program finder"
	app.Action = find

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

func find(ctx *cli.Context) error {
	db, err := sqlx.Connect("mysql", "moto:motomoto@(192.168.0.40:3306)/epgrec?parseTime=true")
	if err != nil {
		log.Fatalln(err)
	}

	programs := []Program{}
	args := map[string]interface{}{
		"category_id": 8,
	}
	nstmt, err := db.PrepareNamed("SELECT rp.title, rp.starttime FROM Recorder_programTbl AS rp LEFT OUTER JOIN Recorder_reserveTbl AS rr ON rr.program_id = rp.id WHERE rp.category_id = :category_id")
	err = nstmt.Select(&programs, args)
	if err != nil {
		log.Fatalln(err)
	}

	for _, v := range programs {
		fmt.Printf("%v\n", v)
	}

	return nil
}
