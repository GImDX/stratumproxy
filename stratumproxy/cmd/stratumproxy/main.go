package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

// 命令行参数
var (
	serverPemFile     string
	serverKeyFile     string
	listenAddr        string
	serverAddr        string
	replacedUser      string
	replacedPassword  string
)

//debug printf
const debug = false

func main() {
	// 解析命令行参数
	flag.StringVar(&serverPemFile, "server-pem", "./server.pem", "Path to server PEM file")
	flag.StringVar(&serverKeyFile, "server-key", "./server.key", "Path to server private key file")
	flag.StringVar(&listenAddr, "listen-addr", ":9999", "Address to listen on")
	flag.StringVar(&serverAddr, "server-addr", ":1177", "Address of the server to connect to")
	flag.StringVar(&replacedUser, "replaced-user", "pyrin:qq0240xcnlk52jt4t007gwe97hnr33g5knx9kkgarmm0p9ghm9sg68qrakyf2.pyi114514", "Replaced user string")
	flag.StringVar(&replacedPassword, "replaced-password", "pyi114514", "Replaced password string")

	flag.Parse()

	// 加载TLS证书和私钥
	cert, err := tls.LoadX509KeyPair(serverPemFile, serverKeyFile)
	if err != nil {
		log.Fatalf("failed to load key pair: %s", err)
	}

	// 创建TLS配置
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// 打印配置信息
	log.Printf("Listening on: %s", listenAddr)
	log.Printf("Server: %s", serverAddr)
	log.Printf("Params: [%s, %s]", replacedUser, replacedPassword)

	listener, err := tls.Listen("tcp", listenAddr, tlsConfig)
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}

	for {
		// 接受客户端连接
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept Client connection: %s", err)
			continue
		} else {
			fmt.Print("\n")
			log.Printf("Succeed to accept Client connection: %s", clientConn.RemoteAddr())
		}

		go handleClient(clientConn, tlsConfig, serverAddr)
	}
}

// handleClient 处理从客户端到服务器的连接
func handleClient(clientConn net.Conn, tlsConfig *tls.Config, serverAddr string) {
    defer clientConn.Close()

    insecureTlsConfig := &tls.Config{InsecureSkipVerify: true}
    serverConn, err := tls.Dial("tcp", serverAddr, insecureTlsConfig)
    if err != nil {
        log.Printf("failed to connect to server: %s", err)
        return
    } else {
		log.Printf("Succeed to establish Server connection: %s", serverConn.RemoteAddr())
	}
    defer serverConn.Close()

    done := make(chan bool)
    var once sync.Once
    closeOnce := func() { once.Do(func() { close(done) }) }

    go func() {
        defer closeOnce()
        processAndForwardClientData(clientConn, serverConn)
    }()

    go func() {
        defer closeOnce()
		logServerToClientData(serverConn, clientConn)
    }()

    <-done
}


func processAndForwardClientData(src, dst net.Conn) {
    scanner := bufio.NewScanner(src)

    for scanner.Scan() {
        line := scanner.Text()

        // 解析 JSON 结构
        var msg map[string]interface{}
        if err := json.Unmarshal([]byte(line), &msg); err != nil {
            log.Printf("Error unmarshalling JSON: %s", err)
            fmt.Fprintf(dst, line+"\n") // 直接转发无法解析的行
            continue
        }

        method, ok := msg["method"].(string)
        if !ok {
            fmt.Fprintf(dst, line+"\n") // 如果 method 字段不存在，直接转发
			if debug {log.Printf("Client:%s -> Server Raw Data: %s", src.RemoteAddr().String(), string(line))}
            continue
        }else {
            if debug {log.Printf("Client:%s -> Server Raw Data: %s", src.RemoteAddr().String(), string(line))}
        }

        // 处理 mining.authorize 和 mining.submit
        if method == "mining.authorize" || method == "mining.submit" || method == "mining.subscribe"{
            params, ok := msg["params"].([]interface{})
			original := params[0] // 记录修改前的 params[0]
            if ok && len(params) >= 1 {
                modifyParams(params) // 调用修改方法
                msg["params"] = params
            }
			if method != "mining.submit"{
				log.Printf("Modifying %v: %v -> %s",method , original, params[0])
			}
        }

        // 重新编码 JSON 并转发
        modifiedLine, err := json.Marshal(msg)
        if err != nil {
            log.Printf("Error marshalling JSON: %s", err)
            fmt.Fprintf(dst, line+"\n")
			if debug {log.Printf("Client:%s -> Server Raw Data: %s", src.RemoteAddr().String(), string(line))}
        } else {
            fmt.Fprintf(dst, string(modifiedLine)+"\n")
			if debug {log.Printf("Client:%s -> Server Raw Data: %s", src.RemoteAddr().String(), string(modifiedLine))}
        }
    }

    // 检测到 EOF 或其他读取错误
    if err := scanner.Err(); err != nil {
        log.Printf("Error reading from client: %v", err)
    } else {
        log.Printf("Client closed connection: %s", src.RemoteAddr())
    }
}

// 只替换 params[0] "." 之前的部分
func modifyParams(params []interface{}) {
    if len(params) >= 1 {
        original, ok := params[0].(string)
        if ok {
            parts := strings.SplitN(original, ".", 2) // 只分割一次
            if len(parts) == 2 {
                params[0] = replacedUser + "." + parts[1] // 拼接新的用户名+原后缀
            } else {
                params[0] = replacedUser // 保险起见，如果没有"."，直接替换
            }
        }
    }
}



func logServerToClientData(src net.Conn, dst net.Conn) {
    scanner := bufio.NewScanner(src)

    for scanner.Scan() {
        line := scanner.Text()

        // 打印从服务器返回的原始数据
        if debug {log.Printf("Server:%s -> Client Raw Data: %s", src.RemoteAddr().String(), string(line))}

        // 转发数据到客户端
        _, err := dst.Write([]byte(line + "\n")) // 确保添加换行符
        if err != nil {
            log.Printf("Error writing to client: %v", err)
            return
        }
    }

    // 检测到 EOF 或其他读取错误
    if err := scanner.Err(); err != nil {
        log.Printf("Error reading from server: %v", err)
    } else {
        log.Printf("Server closed connection to Client: %s", dst.RemoteAddr())
    }
}
