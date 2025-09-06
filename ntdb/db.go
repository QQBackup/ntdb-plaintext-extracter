package ntdb

import (
	"strings"
	"time"

	sql "github.com/FloatTech/sqlite"
	"github.com/QQBackup/ntdb-plaintext-extracter/helper"
	"github.com/QQBackup/ntdb-plaintext-extracter/model"
	"github.com/pkg/errors"
)

type NTDatabase sql.Sqlite

func NewNTDatabase(dbpath string, cachettl time.Duration) (ntdb NTDatabase, err error) {
	db := sql.New(dbpath)
	err = db.Open(cachettl)
	if err != nil {
		return
	}
	ntdb = NTDatabase(db)
	return
}

func (ntdb *NTDatabase) Close() error {
	return (*sql.Sqlite)(ntdb).Close()
}

func (ntdb *NTDatabase) GetUserInfoByUserID(userID string) (*UserInfo, error) {
	ln := model.Row{}
	err := (*sql.Sqlite)(ntdb).Find(
		"group_msg_table", &ln,
		"WHERE [40020]=? AND ([40090]<>'' OR [40093]<>'')",
		ln.UserID,
	)
	if err != nil {
		return nil, errors.Wrap(err, helper.ThisFuncName()+" "+userID)
	}
	if ln.SenderName == "" {
		ln.SenderName = ln.SenderName2
	}
	return &UserInfo{
		Uin:      ln.Uin,
		UserID:   ln.UserID,
		Nickname: ln.SenderName,
	}, nil
}

func (ntdb *NTDatabase) RangeMessages(fn func(*model.Row) error, q ...string) error {
	ln := model.Row{}
	questions := "WHERE [40011]=2 AND [40012]=1 ORDER BY [40050]"
	if len(q) > 0 {
		questions = strings.Join(q, " ")
	}
	return (*sql.Sqlite)(ntdb).FindFor(
		"group_msg_table", &ln,
		questions,
		func() error { return fn(&ln) },
	)
}
