package database

import (
	"database/sql"
	"github.com/bwmarrin/lit"
	"strconv"
	"strings"
)

// ExecQuery executes a simple query given a DB
func ExecQuery(db *sql.DB, query ...string) {
	for _, q := range query {
		_, err := db.Exec(q)
		if err != nil {
			lit.Error("Error executing query, %s", err)
		}
	}
}

// EncodeSegments returns an encoded string of segments
func EncodeSegments(segments map[int]bool) string {
	if segments == nil {
		return ""
	}

	var out string

	for k := range segments {
		out += strconv.Itoa(k) + ","
	}

	return strings.TrimSuffix(out, ",")
}

// DecodeSegments decodes segments into a map
func DecodeSegments(segments string) map[int]bool {
	if segments == "" {
		return nil
	}

	splitted := strings.Split(segments, ",")
	mapSegments := make(map[int]bool, len(splitted))

	for _, s := range splitted {
		frame, err := strconv.Atoi(s)
		if err == nil {
			mapSegments[frame] = true
		}
	}

	return mapSegments
}
