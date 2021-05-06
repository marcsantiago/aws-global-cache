package awsglobalcache

import (
	"log"
	"os"
)

func init() {
	WarningLogger = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
}

var (
	WarningLogger *log.Logger
)
