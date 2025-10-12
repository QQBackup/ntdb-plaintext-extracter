// Package main 是一个解析消息为纯文本的示例, 可通过引用 package 的方式扩展更多处理逻辑.
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/QQBackup/ntdb-plaintext-extracter/model"
	"github.com/QQBackup/ntdb-plaintext-extracter/ntdb"
)

func main() {
	chats := os.Args[1:] // 聊天记录文件路径 (.db)

	for _, name := range chats {
		if strings.HasSuffix(name, ".db") {
			db, err := ntdb.NewNTDatabase(name, time.Hour)
			if err != nil {
				panic(err)
			}
			err = db.RangeMessages(func(ln *model.Row) error {
				inf, err := db.GetUserInfoByUserID(ln.UserID) // this is slow, use cache in production env.
				if err != nil {
					return err
				}
				msb := strings.Builder{}
				msb.WriteString("【")
				msb.WriteString(inf.Nickname)
				msb.WriteString("(")
				msb.WriteString(strconv.FormatInt(inf.Uin, 10))
				msb.WriteString(")】")
				msb.WriteString(ln.Msg.String())
				msb.WriteByte('\n')
				msb.WriteByte('\n')
				fmt.Print(msb.String())
				return nil
			})
			if err != nil {
				panic(err)
			}
			_ = db.Close()
			continue
		}
		panic("unsupported file " + name)
	}
}
