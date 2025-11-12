package ntdb

import (
	"time"
	"unsafe"

	sql "github.com/FloatTech/sqlite"
	"github.com/QQBackup/ntdb-plaintext-extracter/helper"
	"github.com/QQBackup/ntdb-plaintext-extracter/model"
	"github.com/pkg/errors"
)

type NTDatabase[
	G model.GroupMessageTables,
] sql.Sqlite

func NewNTDatabase[G model.GroupMessageTables](dbpath string, cachettl time.Duration) (ntdb NTDatabase[G], err error) {
	db := sql.New(dbpath)
	err = db.Open(cachettl)
	if err != nil {
		return
	}
	ntdb = NTDatabase[G](db)
	return
}

func (ntdb *NTDatabase[G]) Close() error {
	return (*sql.Sqlite)(ntdb).Close()
}

func (ntdb *NTDatabase[G]) GetGroupUserInfoByUserID(userID string) (*UserInfo, error) {
	ln := model.NewLargestGroupMessageTable()
	err := (*sql.Sqlite)(ntdb).Find(
		"group_msg_table", (*G)(unsafe.Pointer(&ln)),
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

func (ntdb *NTDatabase[G]) RangeMessages(gid int64, fn func(*G) error) error {
	var ln G
	questions := "WHERE ([40021]=? OR [40027]=?) AND [40011]=2 AND [40012]=1 ORDER BY [40050]"
	return (*sql.Sqlite)(ntdb).FindFor(
		"group_msg_table", &ln,
		questions,
		func() error { return fn(&ln) }, gid, gid,
	)
}

func (ntdb *NTDatabase[G]) GetAllGroupIDs() (x []int64, err error) {
	x = make([]int64, 0, 64)
	gid := struct {
		GIDs int64
	}{}
	err = (*sql.Sqlite)(ntdb).QueryFor(`SELECT CAST([40021] AS INTEGER) AS GIDs FROM group_msg_table
UNION
SELECT CAST([40027] AS INTEGER) AS GIDs FROM group_msg_table;`, &gid, func() error {
		x = append(x, gid.GIDs)
		return nil
	})
	return
}
