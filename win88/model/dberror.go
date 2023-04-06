package model

import (
	"errors"
	"fmt"
)

var (
	NOT_OPEN          = "not open"
	ErrRPClientNoConn = errors.New("RPClient no connect")
)

func NewDBError(dbName, collectionName, errMsg string) error {
	err := &DBError{
		DBName:         dbName,
		CollectionName: collectionName,
		Msg:            errMsg,
	}
	return err
}

type DBError struct {
	DBName         string
	CollectionName string
	Msg            string
}

func (dbe *DBError) Error() string {
	return fmt.Sprintf("[%v].[%v]: %v", dbe.DBName, dbe.CollectionName, dbe.Msg)
}
