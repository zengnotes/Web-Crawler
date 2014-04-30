package commonstripper

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	. "github.com/PuerkitoBio/goquery"
)


type Stripper interface {
	StripPage( *Document,  *sql.DB, chan string)
	IgnoreUrl(*Document)
}

