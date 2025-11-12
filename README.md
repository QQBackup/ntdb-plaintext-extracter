# ntdb-plaintext-extracter
提取QQ NT数据库 group_msg_table 中的纯文本

## 用法
> [!Tips]
> 如果提取消息失败，可尝试添加`-ver 数字`参数以适配不同版本数据库格式，默认值：`0`，目前可取：`0~2`。

1. 打印数据库中所有群号
    ```go
    go run main.go -pg nt_msg.body.db
    ```
2. 直接按时间顺序输出群号 1234567 的全部消息到屏幕
    ```go
    go run main.go -g 1234567 nt_msg.body.db
    ```
2. 按时间顺序写入群号 1234567 的全部消息到文本文件
    ```go
    go run main.go -g 1234567 nt_msg.body.db > chat.txt
    ```
