# General description
WeightWatcher is a very simple utility to keep track of a human weight. It should be fully operated from terminal using subcommands and flags. Some of the flags, such as data file path, should be stored in configuration file so that a user can skip it in everyday usage.

Utility should enable to add, edit and remove the weight measurement. It should calculate current weight based but making average from a few last measurements (the period should be a user parameter).

The data stored should be available for use in other applications, so it should be stored using a standard and commonly used file format.

# Technical requirement
- Language: golang
- License: GNU General Public License
- Data file: sqlite3 database flat file
- Configuration file format: json

