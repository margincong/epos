


### gping

使用 go 实现的 ping 工具，主要用于监测应用的网络连通性情况。

```
用法: ping [-t true] [-n 10] [-l 32] [-w 1500] [-f 10] [-s 10] target_name
选项:
    -t                 Ping 指定的主机，直到停止。
                   	   若要停止，请键入 Ctrl+C。
<<<<<<< HEAD
    -n count           要发送的回显请求数。
    -l size            发送缓冲区大小。
    -w timeout         等待每次回复的超时时间(毫秒)。
    -f failtimes       最大能够接受的连续失败次数。
    -s sleepTime       每次发送请求的时间间隔（秒）
```
=======
   	-n count           要发送的回显请求数。
        -l size            发送缓冲区大小。
	-w timeout         等待每次回复的超时时间(毫秒)。
	-f failtimes       最大能够接受的连续失败次数。
	-s sleepTime       每次发送请求的时间间隔（秒）
```
>>>>>>> 8c0d4d66c04cf2d0cc7438d03c735e470bca27f4
