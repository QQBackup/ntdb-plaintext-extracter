// Package main 是一个解析消息为纯文本的示例, 可通过引用 package 的方式扩展更多处理逻辑.
package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/QQBackup/ntdb-plaintext-extracter/model"
	"github.com/QQBackup/ntdb-plaintext-extracter/ntdb"
)

func main() {
	pg := flag.Bool("pg", false, "print all available group IDs")
	g := flag.Int64("g", 0, "print messages in this group")
	ver := flag.Int("ver", 0, "the database version, try another if panic (available: 0-1)")
	flag.Parse()

	chats := flag.Args() // 聊天记录文件路径 (.db)

	switch *ver {
	case 0:
		domain[model.GroupMessageTableRow](*pg, *g, chats)
	case 1:
		domain[model.GroupMessageTableRowNew](*pg, *g, chats)
	default:
		panic("unsupported version: " + strconv.Itoa(*ver))
	}
}

func domain[G model.GroupMessageTables](pg bool, g int64, chats []string) {
	for _, name := range chats {
		if strings.HasSuffix(name, ".db") {
			fmt.Println("---------------", name, "---------------")
			db, err := ntdb.NewNTDatabase[G](name, time.Hour)
			if err != nil {
				panic(err)
			}
			if pg {
				gids, err := db.GetAllGroupIDs()
				if err != nil {
					panic(err)
				}
				for _, gid := range gids {
					fmt.Println(gid)
				}
				continue
			}
			err = db.RangeMessages(g, func(lng *G) error {
				ln := model.ToSmallestGroupMessageTable(lng)
				inf, err := db.GetGroupUserInfoByUserID(ln.UserID) // this is slow, use cache in production env.
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
				fmt.Println("[HINT] failed to parse group messages, maybe try a different -ver parameter?")
				fmt.Println()
				panic(err)
			}
			_ = db.Close()
			continue
		}
		panic("unsupported file " + name)
	}
}
