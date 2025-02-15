package main

import (
	"flag"
	"fmt"
	"math/rand"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	sql "github.com/FloatTech/sqlite"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/source"
	"github.com/xitongsys/parquet-go/writer"

	"github.com/QQBackup/ntdb-plaintext-extracter/model"
)

const debug = false

type export struct {
	Conversations []conversasion `parquet:"name=conversations, type=LIST, convertedtype=LIST"`
}

type conversasion struct {
	From  string `parquet:"name=from, type=BYTE_ARRAY, convertedtype=UTF8"`
	Value string `parquet:"name=value, type=BYTE_ARRAY, convertedtype=UTF8"`
}

var (
	urlre  = regexp.MustCompile(`[a-zA-z]+://[^\s]*`)
	hashre = regexp.MustCompile(`^[a-fA-F0-9]{8,}$`)
	system = conversasion{From: "system"}
)

func main() {
	outpath := flag.String("o", "train-00000-of-00001.parquet", "输出文件位置, 包含一个格式化打印数字")
	systemprompt := flag.String("s", "你正在QQ群与用户交谈，以一句话做出回应。如果你愿意，也可以开启新话题。", "系统提示词")
	chatbatch := flag.Uint("b", 32, "几个聊天为一组")
	printonly := flag.Bool("p", false, "仅打印")
	flag.Parse()

	system.Value = *systemprompt
	chats := flag.Args() // 群友聊天记录文件路径 (.db)

	var (
		err  error
		fw   source.ParquetFile
		pw   *writer.ParquetWriter
		exps []export
	)
	if !*printonly {
		fw, err = local.NewLocalFileWriter(*outpath)
		if err != nil {
			panic(err)
		}
		pw, err = writer.NewParquetWriter(fw, &export{}, int64(runtime.NumCPU()))
		if err != nil {
			panic(err)
		}
		pw.RowGroupSize = 64 * 1024 * 1024 // 64M
		pw.PageSize = 4 * 1024             // 4K
		pw.CompressionType = parquet.CompressionCodec_SNAPPY
		exps = make([]export, 0, 65536)
	}
	for _, name := range chats {
		ln := model.Row{}
		exp := export{Conversations: []conversasion{system}}
		cnt := 0
		user := ""
		assistant := ""
		isuserturn := true
		iterfn := func(msg string) error {
			msb := strings.Builder{}
			msb.WriteString("【")
			msb.WriteString(ln.SenderName)
			msb.WriteString("】")
			msb.WriteString(msg)
			msb.WriteByte('\n')
			msb.WriteByte('\n')
			test := msb.String()
			if *printonly {
				fmt.Print(test)
				return nil
			}
			if !utf8.ValidString(msg) ||
				strings.Contains(test, "[CQ:") ||
				strings.Contains(test, "magnet:?xt=") ||
				strings.Contains(test, "[bot]") ||
				strings.Contains(test, "[BOT]") ||
				strings.Contains(test, "[Bot]") {
				return nil
			}
			cv := conversasion{Value: msg}
			switch {
			case len(exp.Conversations) == 1:
				cv.From = "human"
				user = ln.SenderName
				isuserturn = false
			case ln.SenderName == user:
				cv.From = "human"
				if !isuserturn {
					old := exp.Conversations[len(exp.Conversations)-1]
					old.Value += "\n" + msg
					return nil
				}
				isuserturn = false
			case assistant == "":
				cv.From = "gpt"
				assistant = ln.SenderName
				isuserturn = true
			case assistant == ln.SenderName:
				cv.From = "gpt"
				if isuserturn {
					old := exp.Conversations[len(exp.Conversations)-1]
					old.Value += "\n" + msg
					return nil
				}
				isuserturn = true
			default: // new person
				if !isuserturn { // make sure gpt end at last
					cv.From = "gpt"
					exp.Conversations = append(exp.Conversations, cv)
				}
				newexp := export{Conversations: make([]conversasion, len(exp.Conversations))}
				copy(newexp.Conversations, exp.Conversations)
				exps = append(exps, newexp)
				exp.Conversations = exp.Conversations[:2]
				cv.From = "human"
				user = ln.SenderName
				assistant = ""
				exp.Conversations[1] = cv
				isuserturn = false
				cnt = 1
				return nil
			}
			exp.Conversations = append(exp.Conversations, cv)
			cnt++
			if cnt%int(*chatbatch) == 0 {
				newexp := export{Conversations: make([]conversasion, len(exp.Conversations))}
				copy(newexp.Conversations, exp.Conversations)
				exps = append(exps, newexp)
				exp.Conversations = exp.Conversations[:1]
				user = ""
				assistant = ""
			}
			return nil
		}
		itertail := func() {
			if len(exp.Conversations) > 0 {
				newexp := export{Conversations: make([]conversasion, len(exp.Conversations))}
				copy(newexp.Conversations, exp.Conversations)
				exps = append(exps, newexp)
				exp.Conversations = exp.Conversations[:1]
				user = ""
				assistant = ""
			}
		}
		if strings.HasSuffix(name, ".db") {
			db := sql.New(name)
			err := db.Open(time.Hour)
			if err != nil {
				panic(err)
			}
			err = db.FindFor("group_msg_table", &ln, "WHERE [40011]=2 AND [40012]=1 ORDER BY [40050]", func() error {
				if ln.SenderName == "" {
					ln.SenderName = ln.SenderName2
				}
				if ln.SenderName == "" {
					ln2 := model.Row{}
					err = db.Find("group_msg_table", &ln2, "WHERE [40020]=? AND ([40090]<>'' OR [40093]<>'')", ln.UserID)
					if err != nil {
						fmt.Println("get user name of", ln.UserID, "err:", err)
					}
					ln.SenderName = ln2.SenderName
					if ln.SenderName == "" {
						ln.SenderName = ln2.SenderName2
					}
				}
				if *printonly {
					return iterfn(ln.Msg.String())
				}
				msg := urlre.ReplaceAllString(ln.Msg.String(), "")
				msg = hashre.ReplaceAllString(msg, "")
				if debug {
					fmt.Println(msg)
				}
				return iterfn(msg)
			})
			if err != nil {
				panic(err)
			}
			if !*printonly {
				itertail()
			}
			_ = db.Close()
			continue
		}
		panic("unsupported file " + name)
	}

	if !*printonly {
		rand.Shuffle(len(exps), func(i, j int) {
			exps[i], exps[j] = exps[j], exps[i]
		})
		for _, exp := range exps {
			err := pw.Write(&exp)
			if err != nil {
				panic(err)
			}
		}
		if err = pw.WriteStop(); err != nil {
			panic(err)
		}
		err = fw.Close()
		if err != nil {
			panic(err)
		}
	}
}
