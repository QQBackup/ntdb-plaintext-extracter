# ntdb-plaintext-extracter
提取QQ NT数据库 group_msg_table 中的纯文本

## 用法
1. 直接输出到屏幕
    ```go
    go run main.go nt_msg.body.db
    ```
2. 写入文本文件
    ```go
    go run main.go nt_msg.body.db > chat.txt
    ```
