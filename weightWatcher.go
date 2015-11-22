// Copyright 2009 Marcin 'Zbroju' Zbroinski. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
package main

import (
	"database/sql"
	"errors"
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
var DB_PROPERTIES = map[string]string{
	"applicationName": "weightWatcher",
	"databaseVersion": "1.0",
}

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
		Name: "weight, w",
		//		Value: 0,
		Usage: "measured weight",
	}
	flagFile := cli.StringFlag{
		Name:  "file, f",
		Value: dataFile,
		Usage: "data file",
	}
	flagId := cli.IntFlag{
		Name:  "id, i",
		Value: -1,
		Usage: "id of edited or removed object",
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
			Action:  cmdAddmeasurement,
		},
		{
			Name:    "edit",
			Aliases: []string{"E"},
			Flags:   []cli.Flag{flagVerbose, flagDate, flagWeight, flagFile, flagId},
			Usage:   "edit a measurement",
			Action:  cmdEditMeasurement,
		},
		{
			Name:    "remove",
			Aliases: []string{"R"},
			Flags:   []cli.Flag{flagVerbose, flagDate, flagFile},
			Usage:   "remove a measurement",
			Action:  cmdRemoveMeasurement,
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

// cmdInit creates a new data file and add basic information about the file to properties table.
func cmdInit(c *cli.Context) {
	// Check the obligatory parameters and exit if missing
	if c.String("file") == "" {
		fmt.Fprint(os.Stderr, "weightWatcher: missing information about data file. Specify it with --file or -f flag.\n")
		return
	}

	// Check if file exist and if so - exit
	if _, err := os.Stat(c.String("file")); !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "weightWatcher: file %s already exists.\n", c.String("file"))
		return
	}

	// Open file
	db, err := sql.Open("sqlite3", c.String("file"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "weightWatcher: %s\n", err)
		return
	}
	defer db.Close()

	// Create tables
	sqlStmt := `
	BEGIN TRANSACTION;
	CREATE TABLE measurements (measurement_id INTEGER PRIMARY KEY, day DATE, measurement REAL);
	CREATE TABLE properties (key TEXT, value TEXT);
	COMMIT;
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
	for key, value := range DB_PROPERTIES {
		_, err = stmt.Exec(key, value)
		if err != nil {
			fmt.Fprintf(os.Stderr, "weightWatcher: %s", err)
			tx.Rollback()
			return
		}
	}
	tx.Commit()

	// Show summary if verbose
	if c.Bool("verbose") == true {
		fmt.Fprintf(os.Stdout, "weightWatcher: created file %s.\n", c.String("file"))
	}
}

// cmdAddMeasurement adds measurement to data file
func cmdAddmeasurement(c *cli.Context) {

	// Check obligatory flags (file, date, measurement)
	if c.String("file") == "" {
		fmt.Fprintf(os.Stderr, "weightWatcher: missing file parameter. Specify it with --file or -f flag.\n")
		return
	}
	if c.String("date") == "" {
		fmt.Fprintf(os.Stderr, "weightWatcher: missing date parameter. Specify it with --date or -d flag.\n")
		return
	}
	if c.Float64("weight") == 0 {
		fmt.Fprintf(os.Stderr, "weightWatcher: missing weight parameter. Specify it with --weight or -w flag.\n")
		return
	}

	// Open data file
	db, err := getDataFile(c.String("file"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	defer db.Close()

	// Add data to file
	sqlStmt := fmt.Sprintf("INSERT INTO measurements VALUES (NULL, '%s', %f);", c.String("date"), c.Float64("weight"))

	_, err = db.Exec(sqlStmt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "weightWatcher: %s\n", err)
		return
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		fmt.Fprintf(os.Stdout, "weightWatcher: add measurement %3.2f to file %s with date %s.\n",
			c.Float64("weight"),
			c.String("file"),
			c.String("date"))
	}
}

// cmdEditMeasurement edit value or date for a measurement with given ID.
func cmdEditMeasurement(c *cli.Context) {
	// Check obligatory flags (id, file)
	if c.Int("id") < 0 {
		fmt.Fprintf(os.Stderr, "weightWatcher: missing ID parameter. Specify it with --id or -i flag.\n")
		return
	}
	if c.String("file") == "" {
		fmt.Fprintf(os.Stderr, "weightWatcher: missing file parameter. Specify it with --file or -f flag.\n")
		return
	}

	// Open data file
	db, err := getDataFile(c.String("file"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	defer db.Close()

	// Check if measurement with given ID exists
	if !measurementExist(c.Int("id"), db) {
		fmt.Fprintf(os.Stderr, "weightWatcher: measurement with id=%d does not exist.\n", c.Int("id"))
		return
	}

	// Edit data
	sqlStmt := "BEGIN TRANSACTION;"
	if c.String("date") != "" {
		sqlStmt += fmt.Sprintf("UPDATE measurements SET day='%s' WHERE measurement_id=%d;", c.String("date"), c.Int("id"))
	}
	if c.Float64("weight") != 0 {
		sqlStmt += fmt.Sprintf("UPDATE measurements SET measurement=%f WHERE measurement_id=%d;", c.Float64("weight"), c.Int("id"))
	}
	sqlStmt += "COMMIT;"
	_, err = db.Exec(sqlStmt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "weightWatcher: %s\n", err)
		return
	}

	// Show summary if verbose
	if c.Bool("verbose") == true {
		fmt.Fprintf(os.Stdout, "weightWatcher: edited measurement %3.2f to file %s with date %s.\n",
			c.Float64("weight"),
			c.String("file"),
			c.String("date"))
	}
}

func cmdRemoveMeasurement(c *cli.Context) {
	//TODO: write command 'remove measurement'
}

func reportSummary(c *cli.Context) {
	//TODO: write report 'show summary'
}

func reportHistory(c *cli.Context) {
	//TODO: write report 'show history'
}

// getDataFile checks if file exists and is a correct weightWatcher data file.
// If so, it returns pointer to the sql.DB, or otherwise nil and error.
func getDataFile(filePath string) (*sql.DB, error) {
	errorMessage := "weightWatcher: file " + filePath + " is not a correct weightWatcher data file."

	// Check if file exist and if not - exit
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, errors.New("weightWatcher: file " + filePath + " does not exist.")
	}

	// Open file
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, errors.New(errorMessage)
	}

	// Check if the file is weightWatcher database
	rows, err := db.Query("SELECT key, value FROM properties;")
	if err != nil {
		return nil, errors.New(errorMessage)
	}
	if rows.Next() == false {
		return nil, errors.New(errorMessage)
	} else {
		for rows.Next() {
			var key, value string
			err = rows.Scan(&key, &value)
			if err != nil {
				return nil, errors.New(errorMessage)
			}
			if DB_PROPERTIES[key] != "" && DB_PROPERTIES[key] != value {
				return nil, errors.New(errorMessage)
			}
		}
	}
	rows.Close()

	return db, nil
}

// measurementExists returns true if a measurement with given id exists, or false otherwise.
func measurementExist(id int, db *sql.DB) bool {
	sqlStmt := fmt.Sprintf("SELECT measurement_id FROM measurements WHERE measurement_id=%d;", id)
	rows, err := db.Query(sqlStmt)
	defer rows.Close()
	if err == nil && rows.Next() {
		return true
	} else {
		return false
	}
}
