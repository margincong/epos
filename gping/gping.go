package main

import (
	"bytes"
	"encoding/binary"
	"flag" // command-line flag parsing
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"time"
)

// ICMP : Internet Control Message Protocol's Structure
type ICMP struct {
	Type        uint8
	Code        uint8
	Checksum    uint16
	Identifier  uint16
	SequenceNum uint16
}

var (
	icmp      ICMP
	laddr     = net.IPAddr{IP: net.ParseIP("ip")}
	num       int
	timeout   int64
	size      int
	stop      bool
	maxfails  int // 最大连续失败次数
	sleepTime int // 每次发送请求的时间间隔，秒数
)

var (
	greenBg      = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	whiteBg      = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellowBg     = string([]byte{27, 91, 57, 48, 59, 52, 51, 109})
	redBg        = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blueBg       = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magentaBg    = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyanBg       = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	green        = string([]byte{27, 91, 51, 50, 109})
	white        = string([]byte{27, 91, 51, 55, 109})
	yellow       = string([]byte{27, 91, 51, 51, 109})
	red          = string([]byte{27, 91, 51, 49, 109})
	blue         = string([]byte{27, 91, 51, 52, 109})
	magenta      = string([]byte{27, 91, 51, 53, 109})
	cyan         = string([]byte{27, 91, 51, 54, 109})
	reset        = string([]byte{27, 91, 48, 109})
	disableColor = false
)

func main() {
	ParseArgs()
	args := os.Args

	if len(args) < 2 {
		Usage()
	}
	desIP := args[len(args)-1]

	// Dial connects to the address on the named network.
	// Known networks are "tcp", "tcp4" (IPv4-only), "tcp6" (IPv6-only), "udp",
	// 		"udp4" (IPv4-only), "udp6" (IPv6-only), "ip", "ip4" (IPv4-only),
	// 		"ip6" (IPv6-only), "unix", "unixgram" and "unixpacket".
	// DialTimeout acts like Dial but takes a timeout.
	// func DialTimeout(network, address string, timeout time.Duration) (Conn, error)
	conn, err := net.DialTimeout("ip:icmp", desIP, time.Duration(timeout)*time.Millisecond)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	//icmp头部填充
	icmp.Type = 8
	icmp.Code = 0
	icmp.Checksum = 0
	icmp.Identifier = 1
	icmp.SequenceNum = 1

	// desIP、size 两个数据是通过命令行传入的参数进行按地址填充的数据
	fmt.Printf("\n正在 ping %s 具有 %d 字节的数据:\n", desIP, size)

	var buffer bytes.Buffer
	// A ByteOrder specifies how to convert byte sequences into 16-, 32-, or 64-bit unsigned integers.
	// BigEndian is the big-endian implementation of ByteOrder.
	// In computing, endianness is the ordering or sequencing of bytes of a word of digital data in computer memory storage
	//		or during transmission. Endianness is primarily expressed as big-endian (BE) or little-endian (LE).
	// A big-endian system stores the most significant byte of a word at the smallest memory address
	//		and the least significant byte at the largest.
	binary.Write(&buffer, binary.BigEndian, icmp)
	data := make([]byte, size)
	buffer.Write(data)
	data = buffer.Bytes()

	var SuccessTimes int      // 成功次数
	var FailTimes int         // 失败次数
	var FailContinueTimes int // 连续失败次数
	var minTime int = int(math.MaxInt32)
	var maxTime int
	var totalTime int

	for i := 0; i < num; i++ {
		icmp.SequenceNum = uint16(1)
		// 检验和设为0
		data[2] = byte(0)
		data[3] = byte(0)

		data[6] = byte(icmp.SequenceNum >> 8)
		data[7] = byte(icmp.SequenceNum)
		icmp.Checksum = CheckSum(data)
		data[2] = byte(icmp.Checksum >> 8)
		data[3] = byte(icmp.Checksum)

		// 开始时间
		t1 := time.Now()
		conn.SetDeadline(t1.Add(time.Duration(time.Duration(timeout) * time.Millisecond)))
		n, err := conn.Write(data)
		if err != nil {
			log.Fatal(err)
		}
		buf := make([]byte, 65535)
		n, err = conn.Read(buf)
		if err != nil {
			fmt.Println("请求超时。")
			FailTimes++
			FailContinueTimes++

			if FailContinueTimes >= maxfails {
				fmt.Println(redBg)
				fmt.Printf("当前已经连续失败 %v 次，请检查网络！", FailContinueTimes)
				break
			}

			continue
		}
		et := int(time.Since(t1) / 1000000)
		if minTime > et {
			minTime = et
		}
		if maxTime < et {
			maxTime = et
		}
		totalTime += et

		fmt.Print(green)
		fmt.Printf("%v 来自 %s 的回复: 字节=%d 时间=%dms TTL=%d\n", time.Now().Format("2006-01-02 15:04:05"), desIP, len(buf[28:n]), et, buf[8])
		SuccessTimes++
		FailContinueTimes = 0
		time.Sleep(time.Duration(sleepTime) * time.Second)

		// 实现 -t 参数
		if stop {
			i = 0
		}
	}

	fmt.Printf("\n%s 的 Ping 统计信息:\n", desIP)
	fmt.Printf("    数据包: 已发送 = %d，已接收 = %d，丢失 = %d (%.2f%% 丢失)，\n", SuccessTimes+FailTimes, SuccessTimes, FailTimes, float64(FailTimes*100)/float64(SuccessTimes+FailTimes))
	if maxTime != 0 && minTime != int(math.MaxInt32) {
		fmt.Printf("往返行程的估计时间(以毫秒为单位):\n")
		fmt.Printf("    最短 = %dms，最长 = %dms，平均 = %dms\n", minTime, maxTime, totalTime/SuccessTimes)
	}
}

// CheckSum : ICMP 报文校验
// 1）将ICMP头部内容中的校验内容(Checksum)的值设为0
// 2）将拼接好(Type+Code+Checksum+Id+Seq+传输Data)的ICMP包按Type开始每两个字节一组（其中Checksum的两个字节都看成0），进行加和处理，如果字节个数为奇数个，则直接加上这个字节内容。说明：这个加和过程的内容放在一个4字节上，如果溢出4字节，则将溢出的直接抛弃
// 3）将高16位与低16位内容加和，直到高16为0
// 4）将步骤三得出的结果取反，得到的结果就是ICMP校验和的值
func CheckSum(data []byte) uint16 {
	var sum uint32
	var length = len(data)
	var index int

	for length > 1 { // 溢出部分直接去除
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
		length -= 2
	}
	if length == 1 {
		sum += uint32(data[index])
	}
	// CheckSum的值是16位，计算是将高16位加低16位，得到的结果进行重复以该方式进行计算，直到高16位为0
	/*
	   sum的最大情况是：ffffffff
	   第一次高16位+低16位：ffff + ffff = 1fffe
	   第二次高16位+低16位：0001 + fffe = ffff
	   即推出一个结论，只要第一次高16位+低16位的结果，再进行之前的计算结果用到高16位+低16位，即可处理溢出情况
	*/
	sum = uint32(sum>>16) + uint32(sum)
	sum = uint32(sum>>16) + uint32(sum)
	return uint16(^sum)
}

// ParseArgs : 解析命令行传入的参数
// e.g.  gping -i=100 www.baidu.com
func ParseArgs() {
	flag.Int64Var(&timeout, "w", 1500, "等待每次回复的超时时间(毫秒)")
	flag.IntVar(&num, "n", 4, "要发送的请求数")
	flag.IntVar(&size, "l", 32, "要发送缓冲区大小")
	flag.BoolVar(&stop, "t", false, "Ping 指定的主机，直到停止")
	flag.IntVar(&maxfails, "f", 10, "最多连续失败次数")
	flag.IntVar(&sleepTime, "s", 1, "每次发送请求的时间间隔")

	flag.Parse()
}

// Usage : 命令行工具的使用说明
func Usage() {
	argNum := len(os.Args)
	if argNum < 2 {
		fmt.Print(
			`
用法: ping [-t true] [-n 10] [-l 32] [-w 1500] [-f 10] [-s 10] target_name
选项:
    -t             Ping 指定的主机，直到停止。
                   若要停止，请键入 Ctrl+C。
    -n count       要发送的回显请求数。
    -l size        发送缓冲区大小。
	-w timeout     等待每次回复的超时时间(毫秒)。
	-f failtimes   最大能够接受的连续失败次数。
	-s sleepTime   每次发送请求的时间间隔
`)
	}
}
