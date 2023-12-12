package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"timetracker/constants"
	"timetracker/internal/models"

	"github.com/golang-module/carbon/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

type Database struct {
	Filename string
	Conn     *sql.DB
	Context  context.Context
}

func New(filename string) *Database {
	conn, err := sql.Open("sqlite3", filename+"?_loc=UTC")
	if err != nil {
		log.Fatalf(err.Error())
		os.Exit(1)
	}

	db := Database{}
	db.Filename = filename
	db.Conn = conn
	db.Context = context.Background()

	// Ping the database to ensure we are connected.
	err = db.Conn.Ping()
	if err != nil {
		log.Fatalf(err.Error())
		os.Exit(1)
	}

	return &db
}

func (db *Database) Close() {
	db.Conn.Close()
}

func (db *Database) Create() {

	// Create the entry table.
	query := "CREATE TABLE entry (uid INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, project TEXT(128) NOT NULL, note TEXT(128), entry_datetime TEXT NOT NULL);"
	_, err := db.Conn.Exec(query)
	if err != nil {
		log.Fatalf(err.Error())
		os.Exit(1)
	}

	// Create the property table.
	query = "CREATE TABLE property (entry_uid INTEGER NOT NULL, name TEXT(128) NOT NULL, value TEXT(128) NOT NULL, CONSTRAINT property_FK FOREIGN KEY (entry_uid) REFERENCES entry(uid) ON DELETE CASCADE);"
	_, err = db.Conn.Exec(query)
	if err != nil {
		log.Fatalf(err.Error())
		os.Exit(1)
	}
}

func (db *Database) InsertNewEntry(entry models.Entry) {
	tx, err := db.Conn.BeginTx(db.Context, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		log.Fatalf(err.Error())
		os.Exit(1)
	}

	result, err := tx.ExecContext(db.Context, "INSERT INTO entry (uid, project, note, entry_datetime) VALUES (?, ?, ?, ?);", nil, entry.Project, entry.Note, entry.EntryDatetime)
	if err != nil {
		rollBackError := tx.Rollback()
		if rollBackError != nil {
			log.Fatalf(rollBackError.Error())
			os.Exit(1)
		}

		log.Fatalf(err.Error())
		os.Exit(1)
	}

	// Now that the record was inserted, get the last inserted id... in our case it it the UID.
	uid, err := result.LastInsertId()
	if err != nil {
		log.Fatalf(err.Error())
		os.Exit(1)
	}

	// Now insert each of the properties for this entry.
	for _, v := range entry.Properties {
		_, err := tx.ExecContext(db.Context, "INSERT INTO property (entry_uid, name, value) VALUES (?, ?, ?);", uid, v.Name, v.Value)
		if err != nil {
			rollBackError := tx.Rollback()
			if rollBackError != nil {
				log.Fatalf(rollBackError.Error())
				os.Exit(1)
			}

			log.Fatalf(err.Error())
			os.Exit(1)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Fatalf(err.Error())
		os.Exit(1)
	}
}

//	select DISTINCT
//	e.uid, entry_datetime
//	from entry e
//	where e.entry_datetime BETWEEN '2023-11-19 00:00:00 -0500 EST' AND '2023-12-01 00:00:00 -0500 EST'
//	order by entry_datetime;
//
//	select e.uid, e.project, e.note, e.entry_datetime, p.name, p.value
//	from entry e
//	left outer join property p on p.entry_uid = e.uid
//	where e.uid in (1, 2, 3, 4)
//	order by entry_datetime;
//
//	select e.uid, e.project, e.note, e.entry_datetime, p.name, p.value
//	from entry e
//	left outer join property p on p.entry_uid = e.uid
//	where e.uid in (1, 2, 3, 4) and p.name = "task" and p.value = "task1"
//	order by e.entry_datetime, p.name, p.value
//
//	select e.uid, e.project, e.note, e.entry_datetime, p.name, p.value
//	from entry e
//	left outer join property p on p.entry_uid = e.uid
//	where e.uid in (1, 2, 3, 4) and p.name = "task" and p.value = "task2"
//	order by e.entry_datetime, p.name, p.value
//
//	select e.uid, e.project, e.note, e.entry_datetime, p.name, p.value
//	from entry e
//	left outer join property p on p.entry_uid = e.uid
//	where e.uid in (1, 2, 3, 4) and p.name = "task" and p.value = "task3"
//	order by e.entry_datetime, p.name, p.value

func (db *Database) GetDistinctUIDs(start carbon.Carbon, end carbon.Carbon) []DistinctUID {
	results, err := db.Conn.Query(`
		SELECT DISTINCT
			e.uid, e.project, e.entry_datetime
		FROM entry e
		WHERE e.entry_datetime BETWEEN ? AND ?
		ORDER BY entry_datetime;
		`, start.ToIso8601String(), end.ToIso8601String(),
	)

	if err != nil {
		log.Fatalf("Fatal error trying to retrieve distinct uids. %s.", err.Error())
		os.Exit(1)
	}

	records := []DistinctUID{}
	for results.Next() {
		var distinctUID DistinctUID
		err = results.Scan(&distinctUID.Uid, &distinctUID.Project, &distinctUID.EntryDatetime)
		if err != nil {
			log.Fatalf("Fatal error trying to scan results into DistinctUID data structure. %s\n", err.Error())
			os.Exit(1)
		}

		records = append(records, distinctUID)
	}

	return records
}

//func (db *Database) GetDistinctEntries(start time.Time, end time.Time) []DistinctEntry {
//	results, err := db.Conn.Query(`
//		SELECT DISTINCT
//			e.uid, e.project, e.entry_datetime
//		FROM entry e
//		WHERE e.entry_datetime BETWEEN ? AND ?
//		ORDER BY entry_datetime;
//		`, start, end,
//	)
//
//	if err != nil {
//		log.Fatalf("Fatal Error trying to retrieve Report records. %s.", err.Error())
//		os.Exit(1)
//	}
//
//	records := []DistinctEntry{}
//	for results.Next() {
//		var distinctEntry DistinctEntry
//		err = results.Scan(&distinctEntry.Uid, &distinctEntry.Project, &distinctEntry.EntryDatetime)
//		if err != nil {
//			log.Fatalf("Fatal error trying to Scan DistinctEntries results into data structure. %s\n", err.Error())
//			os.Exit(1)
//		}
//
//		records = append(records, distinctEntry)
//	}
//
//	return records
//}

func (db *Database) GetEntries(in string) []Entry {
	var s string = fmt.Sprintf("SELECT e.uid, e.project, e.note, e.entry_datetime, p.name, p.value FROM entry e LEFT OUTER JOIN property p on p.entry_uid = e.uid WHERE e.uid IN (%s) ORDER BY entry_datetime;", in)

	results, err := db.Conn.Query(s)
	if err != nil {
		log.Fatalf("Fatal Error trying to retrieve Entry records. %s.", err.Error())
		os.Exit(1)
	}

	records := []Entry{}
	for results.Next() {
		var entry Entry
		err = results.Scan(&entry.Uid, &entry.Project, &entry.Note, &entry.EntryDatetime, &entry.Name, &entry.Value)
		if err != nil {
			log.Fatalf("Fatal error trying to Scan Entries results into data structure. %s\n", err.Error())
			os.Exit(1)
		}

		records = append(records, entry)
	}

	return records
}

func (db *Database) GetLastEntry() models.Entry {
	result, err := db.Conn.QueryContext(db.Context, "SELECT e.uid FROM entry e ORDER BY entry_datetime DESC LIMIT 1;")
	if err != nil {
		log.Fatalf("Fatal Error trying to retrieve last Uid. %s.", err.Error())
		os.Exit(1)
	}

	var lastUid int64
	result.Next()
	err = result.Scan(&lastUid)
	if err != nil {
		log.Fatalf("Fatal error trying to Scan last Uid into data structure. %s\n", err.Error())
		os.Exit(1)
	}

	result.Close()

	var s string = fmt.Sprintf("SELECT e.uid, e.project, e.note, e.entry_datetime, p.name, p.value FROM entry e LEFT OUTER JOIN property p on p.entry_uid = e.uid WHERE e.uid = %d ORDER BY entry_datetime;", lastUid)
	results, err := db.Conn.QueryContext(db.Context, s)
	if err != nil {
		log.Fatalf("Fatal Error trying to retrieve last Uid's Entry records. %s.", err.Error())
		os.Exit(1)
	}

	records := []Entry{}
	for results.Next() {
		var entry Entry
		err = results.Scan(&entry.Uid, &entry.Project, &entry.Note, &entry.EntryDatetime, &entry.Name, &entry.Value)
		if err != nil {
			log.Fatalf("Fatal error trying to Scan last Uid's Entries results into data structure. %s\n", err.Error())
			os.Exit(1)
		}

		records = append(records, entry)
	}

	results.Close()

	var entry models.Entry
	for i, e := range records {
		if i == 0 {
			entry = models.NewEntry(e.Uid, e.Project, e.Note.String, e.EntryDatetime)

			if strings.EqualFold(e.Project, constants.HELLO) {
				break
			}
		} else {
			entry.AddProperty(e.Name.String, e.Value.String)
		}
	}

	return entry
}

func (db *Database) UpdateEntry(entry Entry) {
	var previous bool = false
	var query strings.Builder
	query.WriteString("UPDATE entry")
	query.WriteString(" SET")

	if entry.Project != constants.EMPTY {
		query.WriteString(fmt.Sprintf(" project = '%s'", entry.Project))
		previous = true
	}

	if entry.Note.Valid {
		if previous {
			query.WriteString(", ")
		}
		query.WriteString(fmt.Sprintf(" note = '%s'", entry.Note.String))
		previous = true
	}

	if entry.EntryDatetime != constants.EMPTY {
		if previous {
			query.WriteString(", ")
		}
		query.WriteString(fmt.Sprintf(" entry_datetime = '%s'", entry.EntryDatetime))
	}

	query.WriteString(fmt.Sprintf(" WHERE uid = %d;", entry.Uid))

	if viper.GetBool("debug") {
		log.Printf("Query[%s]\n", query.String())
	}

	// Execute the update.
	_, err := db.Conn.ExecContext(db.Context, query.String())
	if err != nil {
		log.Fatalf(err.Error())
		os.Exit(1)
	}
}
