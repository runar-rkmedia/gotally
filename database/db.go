package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/runar-rkmedia/go-common/logger"
)

func NewDatabase(l logger.AppLogger, dsn string) (DB, error) {
	if dsn == "" {
		dsn = os.Getenv("DSN")
	}
	if dsn == "" {
		dsn = "root:secret@tcp(localhost)/tallyboard"
	}
	debug := l.HasDebug()

	hostAndSuch := strings.Split(dsn, "@")
	if len(hostAndSuch) >= 2 {
		if debug {
			l.Debug().Str("partialConnectionString", hostAndSuch[len(hostAndSuch)-1]).Msg("connecting to database")
		}

	} else {
		if debug {
			l.Debug().Str("partialConnectionString", "hidden").Msg("connecting to database")
		}
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return DB{}, err
	}
	d := DB{db, l}
	go func() {
		if err := d.Deploy(); err != nil {
			l.Fatal().Err(err).Msg("failed to deploy database")
		}
		err := db.Ping()
		if err != nil {
			l.Fatal().Err(err).Msg("failed to ping database")
		}
		if debug {
			l.Debug().Msg("Successfully connected to database")
		}
	}()
	return d, err
}

type DB struct {
	db *sql.DB
	l  logger.AppLogger
}

func (db DB) Deploy() error {
	definition := `
	CREATE TABLE IF NOT EXISTS board_vote (
		id                  VARCHAR(255) NOT NULL,
	  user                VARCHAR(255) NOT NULL,
	  userName            VARCHAR(255) NOT NULL,
		createdAt           DATETIME NOT NULL,
		updatedAt           DATETIME,
		funVote             INT DEFAULT 0,
	  primary key(id, user)
	);
	`

	if db.l.HasDebug() {
		db.l.Debug().
			Str("definition", definition).
			Msg("ensuring table is created")

	}
	_, err := db.db.Exec(strings.TrimSpace(definition))
	if err != nil {
		db.l.Error().
			Err(err).
			Str("definition", definition).
			Msg("failed to write definition to file ")
	}
	return err
}

func (db DB) VoteForBoard(id, user, userName string, funVote int) (*VoteOld, error) {
	vote := VoteOld{
		ID:        id,
		User:      user,
		UserName:  userName,
		CreatedAt: time.Now(),
		FunVote:   funVote,
	}
	if id == "" {
		return nil, fmt.Errorf("Empty id")
	}
	if user == "" {
		return nil, fmt.Errorf("Empty user")
	}
	if userName == "" {
		return nil, fmt.Errorf("Empty user")
	}
	_, err := db.db.Exec("INSERT INTO board_vote (id, user, userName, createdAt, funVote) VALUES(?, ?, ?, NOW(), ?) ON DUPLICATE KEY UPDATE funVote = ?, updatedAT = NOW()",
		vote.ID,
		vote.User,
		vote.UserName,
		vote.FunVote,
		vote.FunVote,
	)
	return &vote, err

}

type DateType time.Time

func (t DateType) String() string {
	return time.Time(t).String()
}
func (db DB) GetAllVotes() (map[string]VoteOld, error) {
	if db.l.HasDebug() {
		db.l.Debug().Msg("getting all votes")
	}
	result, err := db.db.Query(
		"SELECT id, user, createdAt, updatedAt, funVote FROM board_vote;")

	if err != nil {
		db.l.Error().
			Err(err).
			Msg("failed to retrieve votes from database")
		return nil, err
	}

	votes := map[string]VoteOld{}

	for result.Next() {
		vote := VoteOld{}
		var createdAt string
		var updatedAt sql.NullString
		err := result.Scan(&vote.ID, &vote.User, &createdAt, &updatedAt, &vote.FunVote)
		if err != nil {
			log.Println("failed to scan result", err)
			continue
		}
		if err := parseTime(&vote.CreatedAt, createdAt); err != nil {
			log.Println("failed to parse time-format", err)
		}
		if err := parseTime(vote.UpdatedAt, updatedAt.String); err != nil {
			log.Println("failed to parse time-format", err)
		}
		votes[vote.ID] = vote
	}
	if db.l.HasDebug() {
		db.l.Debug().Int("voteCount", len(votes)).Msg("Got votes")
	}

	return votes, nil

}
func (db DB) GetVotesForBoardByUserName(userName string) (map[string]VoteOld, error) {
	result, err := db.db.Query(
		"SELECT id, user, createdAt, updatedAt, funVote FROM board_vote WHERE userName = ?", userName)
	if err != nil {
		return nil, err
	}
	votes := map[string]VoteOld{}
	for result.Next() {
		vote := VoteOld{}
		var createdAt string
		var updatedAt sql.NullString
		err := result.Scan(&vote.ID, &vote.User, &createdAt, &updatedAt, &vote.FunVote)
		if err != nil {
			log.Println("failed to scan result", err)
			continue
		}
		var upd time.Time
		if err := parseTime(&vote.CreatedAt, createdAt); err != nil {
			log.Println("failed to parse time-format", err)
		}
		if err := parseTime(&upd, updatedAt.String); err == nil {
			vote.UpdatedAt = &upd
		}
		votes[vote.ID] = vote
	}
	return votes, err
}
func parseTime(datePointer *time.Time, value string) error {
	if value == "" {
		return nil
	}
	c, err := time.Parse("2006-01-02 15:04:05", value)
	if err != nil {
		return err
	}
	(*datePointer) = c
	return nil
}

type VoteOld struct {
	ID, User, UserName string
	CreatedAt          time.Time
	UpdatedAt          *time.Time
	FunVote            int
}
