// Copyright 2009 Marcin 'Zbroju' Zbroinski. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
package main

import (
	"database/sql"
	"fmt"
	"github.com/codegangsta/cli"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zbroju/gprops"
	"os"
	"path"
	"strconv"
	"time"
)

// Config settings
const (
	CONF_DATAFILE = "DATA_FILE"
	CONF_VERBOSE  = "VERBOSE"
)

// Database properties
const (
	DB_PROP_APPNAME_KEY   = "applicationName"
	DB_PROP_APPNAME_VALUE = "weightWatcher"

	DB_PROP_VERSION_KEY   = "databaseVersion"
	DB_PROP_VERSION_VALUE = "1.0"
)

func main() {
	dataFile := ""
	verbose := false

	// Loading properties from config file if exists
	configSettings := gprops.NewProps()
	configFile, err := os.Open(path.Join(os.Getenv("HOME"), ".wwrc"))
	if err == nil {
		err = configSettings.Load(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "weightWatcher: syntax error in %s. Exit.\n", configFile.Name())
			return
		}
	}
	configFile.Close()
	if configSettings.ContainsKey(CONF_DATAFILE) {
		dataFile = configSettings.Get(CONF_DATAFILE)
	}
	if configSettings.ContainsKey(CONF_VERBOSE) {
		verbose, err = strconv.ParseBool(configSettings.Get(CONF_VERBOSE))
		if err != nil {
			verbose = false
		}
	}

	// Commandline arguments
	app := cli.NewApp()
	app.Name = "weightWatcher"
	app.Usage = "keeps track of your weight"
	app.Version = "0.1"
	app.Authors = []cli.Author{
		cli.Author{"Marcin 'Zbroju' Zbroinski", "marcin@zbroinski.net"},
	}

	// Flags definitions
	flagDate := cli.StringFlag{
		Name:  "date, d",
		Value: today(),
		Usage: "date of measurement (format: YYYY-MM-DD)",
	}
	flagVerbose := cli.BoolFlag{
		Name:        "verbose, b",
		Usage:       "show more output",
		Destination: &verbose,
	}
	flagWeight := cli.Float64Flag{
		Name:  "weight, w",
		Value: 0,
		Usage: "measured weight",
	}
	flagFile := cli.StringFlag{
		Name:  "file, f",
		Value: dataFile,
		Usage: "data file",
	}

	// Commands
	app.Commands = []cli.Command{
		{
			Name:    "init",
			Aliases: []string{"I"},
			Flags:   []cli.Flag{flagVerbose, flagFile},
			Usage:   "init a new data file specified by the user",
			Action:  cmdInit,
		},
		{
			Name:    "add",
			Aliases: []string{"A"},
			Flags:   []cli.Flag{flagVerbose, flagDate, flagWeight, flagFile},
			Usage:   "add a new measurement",
			Action:  cmdAdd,
		},
		{
			Name:    "edit",
			Aliases: []string{"E"},
			Flags:   []cli.Flag{flagVerbose, flagDate, flagWeight, flagFile},
			Usage:   "edit a measurement",
			Action:  cmdEdit,
		},
		{
			Name:    "remove",
			Aliases: []string{"R"},
			Flags:   []cli.Flag{flagVerbose, flagDate, flagFile},
			Usage:   "remove a measurement",
			Action:  cmdRemove,
		},
		{
			Name:    "show",
			Aliases: []string{"S"},
			Usage:   "show report",
			// Reports
			Subcommands: []cli.Command{
				{
					Name:   "summary",
					Usage:  "current weight (average of last few days)",
					Action: reportSummary,
				},
				{
					Name:   "history",
					Usage:  "historical data with moving average (<x> periods)",
					Action: reportHistory,
				},
			},
		},
	}
	app.Run(os.Args)
}

// today returns string with actual date
func today() string {
	year, month, day := time.Now().Date()
	return dateString(year, int(month), day)
}

// dateString returns string with given year, month and day in the format: YYYY-MM-DD
func dateString(year, month, day int) string {
	yearString := strconv.Itoa(year)
	monthString := strconv.Itoa(month)
	dayString := strconv.Itoa(day)
	if len(dayString) < 2 {
		dayString = "0" + dayString
	}

	return yearString + "-" + monthString + "-" + dayString
}

func cmdInit(c *cli.Context) {
	// Open file
	db, err := sql.Open("sqlite3", c.String("file"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "weightWatcher: %s\n", err)
		return
	}
	defer db.Close()

	// Create tables
	sqlStmt := `
	CREATE TABLE measurements (day DATE, measurement REAL);
	CREATE TABLE properties (key TEXT, value TEXT);
	`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "weightWatcher:  %s\n", err)
		return
	}

	// Insert properties values
	tx, err := db.Begin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "weightWatcher: %s\n", err)
		return
	}
	stmt, err := tx.Prepare("INSERT INTO properties VALUES (?,?);")
	if err != nil {
		fmt.Fprint(os.Stderr, "weightWatcher: %s", err)
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(DB_PROP_APPNAME_KEY, DB_PROP_APPNAME_VALUE)
	_, err = stmt.Exec(DB_PROP_VERSION_KEY, DB_PROP_VERSION_VALUE)
	if err != nil {
		fmt.Fprintf(os.Stderr, "weightWatcher: %s", err)
		tx.Rollback()
		return
	}
	tx.Commit()
}

func cmdAdd(c *cli.Context) {
	//TODO: write command 'add measurement'
}

func cmdEdit(c *cli.Context) {
	//TODO: write command 'edit measurement'
}

func cmdRemove(c *cli.Context) {
	//TODO: write command 'remove measurement'
}

func reportSummary(c *cli.Context) {
	//TODO: write report 'show summary'
}

func reportHistory(c *cli.Context) {
	//TODO: write report 'show history'
}
