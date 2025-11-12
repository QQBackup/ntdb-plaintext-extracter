package model

import (
	"fmt"
	"unsafe"
)

type GroupMessageTables interface {
	GroupMessageTableRow | GroupMessageTableRowNew
}

func ToSmallestGroupMessageTable[G GroupMessageTables](x *G) *GroupMessageTableRow {
	return (*GroupMessageTableRow)(unsafe.Pointer(x))
}

func NewLargestGroupMessageTable() (x GroupMessageTableRowNew) {
	return
}

type GroupMessageTableRow struct {
	ID          int64   `db:"40001"`
	Unk0        int64   `db:"40002"`
	MsgID       int64   `db:"40003"`
	Unk1        int     `db:"40010"`
	MsgType1    int     `db:"40011"` // 2 is normal
	MsgType2    int     `db:"40012"` // 1 is normal
	SenderType  int     `db:"40013"`
	UserID      string  `db:"40020"`
	Unk2        int     `db:"40026"`
	GroupID     string  `db:"40021"`
	GroupID2    int64   `db:"40027"`
	IsSelf      int     `db:"40040"`
	Unk3        int     `db:"40041"`
	Timestamp   int64   `db:"40050"`
	Unk4        int     `db:"40052"`
	SenderName  string  `db:"40090"`
	SenderName2 string  `db:"40093"`
	Msg         Message `db:"40800"`
	Unk5        []byte  `db:"40900"`
	IsSelf2     int     `db:"40105"`
	IsSelf3     int     `db:"40005"`
	Timestamp2  int64   `db:"40058"`
	UserID2     int64   `db:"40006"`
	Unk6        int     `db:"40100"`
	Unk7        []byte  `db:"40600"`
	Unk8        int     `db:"40060"`
	Unk9        int     `db:"40850"`
	Unk10       int     `db:"40851"`
	Unk11       []byte  `db:"40601"`
	Unk12       []byte  `db:"40801"`
	Unk13       []byte  `db:"40605"`
	GroupID3    int64   `db:"40030"`
	Uin         int64   `db:"40033"` // Uin 一般为QQ号
}

type GroupMessageTableRowNew struct {
	ID          int64   `db:"40001"`
	Unk0        int64   `db:"40002"`
	MsgID       int64   `db:"40003"`
	Unk1        int     `db:"40010"`
	MsgType1    int     `db:"40011"` // 2 is normal
	MsgType2    int     `db:"40012"` // 1 is normal
	SenderType  int     `db:"40013"`
	UserID      string  `db:"40020"`
	Unk2        int     `db:"40026"`
	GroupID     string  `db:"40021"`
	GroupID2    int64   `db:"40027"`
	IsSelf      int     `db:"40040"`
	Unk3        int     `db:"40041"`
	Timestamp   int64   `db:"40050"`
	Unk4        int     `db:"40052"`
	SenderName  string  `db:"40090"`
	SenderName2 string  `db:"40093"`
	Msg         Message `db:"40800"`
	Unk5        []byte  `db:"40900"`
	IsSelf2     int     `db:"40105"`
	IsSelf3     int     `db:"40005"`
	Timestamp2  int64   `db:"40058"`
	UserID2     int64   `db:"40006"`
	Unk6        int     `db:"40100"`
	Unk7        []byte  `db:"40600"`
	Unk8        int     `db:"40060"`
	Unk9        int     `db:"40850"`
	Unk10       int     `db:"40851"`
	Unk11       []byte  `db:"40601"`
	Unk12       []byte  `db:"40801"`
	Unk13       []byte  `db:"40605"`
	GroupID3    int64   `db:"40030"`
	Uin         int64   `db:"40033"` // Uin 一般为QQ号
	Unk14       []byte  `db:"40062"`
	Unk15       int     `db:"40083"`
	Unk16       int     `db:"40884"`
	Unk17       int     `db:"40808"`
	Unk18       int     `db:"40009"`
}

type Message []byte

func (m Message) String() string {
	tagstatus := 0
	remainlen := 0
	text := []byte{}
	if debug {
		defer fmt.Print("\n")
	}
	for _, ch := range m {
		if debug {
			fmt.Printf("%02x ", ch)
		}
		if tagstatus == 0 {
			if ch == 0x82 {
				tagstatus = 1
				if debug {
					fmt.Print("tag start, ")
				}
				continue
			}
			if debug {
				fmt.Print("tag ignore, ")
			}
			continue
		}
		if tagstatus == 1 {
			if ch == 0x16 { // pure text
				tagstatus = 2
				if debug {
					fmt.Print("tag text, ")
				}
				continue
			}
			if debug {
				fmt.Print("tag others, ")
			}
			tagstatus = -2
			continue
		}
		if tagstatus == 3 || tagstatus == -3 { // going
			if tagstatus > 0 {
				if ch == 0 {
					text = append(text, '\n')
				} else {
					text = append(text, ch)
					if ch == 0xfd {
						panic("unexpected")
					}
				}
			}
			remainlen--
			if remainlen == 0 {
				tagstatus = 0
			}
			continue
		}
		// tagstatus == -2 or 2
		remainlen = int(ch)
		if debug {
			fmt.Print("tag data len ", remainlen, ", ")
		}
		if tagstatus > 0 {
			tagstatus++
		} else {
			tagstatus--
			remainlen--
		}
		if remainlen <= 0 {
			remainlen = 0
			tagstatus = 0
			continue
		}
	}
	return string(text)
}
