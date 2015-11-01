// Copyright 2009 Marcin 'Zbroju' Zbroinski. All rights reserved.
// Use of this source code is governed by a GNU General Public License
// that can be found in the LICENSE file.
package main

import (
	"github.com/codegangsta/cli"
	"os"
	"strconv"
	"time"
)

func main() {

	//TODO: read config data (JSON format) - if the file doesn't exists - create a new one

	// Commandline arguments
	app := cli.NewApp()
	app.Name = "weightWatcher"
	app.Usage = "keeps track of your weight"
	app.Version = "0.1"
	app.Authors = []cli.Author{
		cli.Author{"Marcin 'Zbroju' Zbroinski", "marcin@zbroinski.net"},
	}

	// Global flags
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, b",
			Usage: "show more output",
		},
		cli.StringFlag{
			Name:  "date, d",
			Value: today(),
			Usage: "date of measurement",
		},
		cli.Float64Flag{
			Name:  "weight, w",
			Value: 0,
			Usage: "measured weight",
		},
		cli.StringFlag{
			Name:  "file, f",
			Value: "",
			Usage: "data file",
			//TODO: assign default value to file taken from config file
		},
	}

	// Commands
	app.Commands = []cli.Command{
		{
			Name:    "init",
			Aliases: []string{"I"},
			Usage:   "init a new data file specified by the user",
			Action:  cmdInit,
		},
		{
			Name:    "add",
			Aliases: []string{"A"},
			Usage:   "add a new measurement",
			Action:  cmdAdd,
		},
		{
			Name:    "edit",
			Aliases: []string{"E"},
			Usage:   "edit a measurement",
			Action:  cmdEdit,
		},
		{
			Name:    "remove",
			Aliases: []string{"R"},
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

func today() string {
	year, month, day := time.Now().Date()
	return dateString(year, int(month), day)
}

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
	//TODO: write command 'init new data file'
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
