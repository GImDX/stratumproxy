package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
)

// MiningAuthorizeMessage 定义了mining.authorize消息的结构
type MiningAuthorizeMessage struct {
	Params []string `json:"params"`
	ID     int      `json:"id"`
	Method string   `json:"method"`
}

// 命令行参数
var (
	serverPemFile     string
	serverKeyFile     string
	listenAddr        string
	serverAddr        string
	replacedUser      string
	replacedPassword  string
)

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
	log.Printf("Listening on %s", listenAddr)
	log.Printf("Connecting to server at %s", serverAddr)
	log.Printf("Replacing Params: [%s, %s]", replacedUser, replacedPassword)

	listener, err := tls.Listen("tcp", listenAddr, tlsConfig)
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}

	for {
		// 接受客户端连接
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %s", err)
			continue
		}

		go handleClient(clientConn, tlsConfig, serverAddr)
	}
}

// handleClient 处理从客户端到服务器的连接
func handleClient(clientConn net.Conn, tlsConfig *tls.Config, serverAddr string) {
	defer clientConn.Close()

	// 用于测试环境的TLS配置，跳过证书验证
	// 在生产环境中，请确保使用有效的证书，并去掉InsecureSkipVerify选项
	insecureTlsConfig := &tls.Config{InsecureSkipVerify: true}

	// 使用新的TLS配置连接到服务器C
	serverConn, err := tls.Dial("tcp", serverAddr, insecureTlsConfig)
	if err != nil {
		log.Printf("failed to connect to server: %s", err)
		return
	}
	defer serverConn.Close()

	done := make(chan bool)

	// 数据转发和处理
	var once sync.Once
	closeOnce := func() {
		once.Do(func() {
			close(done)
		})
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic: %v", r)
			}
			log.Printf("Client %s to Server %s connection closed", clientConn.RemoteAddr(), serverConn.RemoteAddr()) // 打印连接关闭信息和原因
			closeOnce()
		}()
		processAndForwardClientData(clientConn, serverConn)
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic: %v", r)
			}
			log.Printf("Server %s to Client %s connection closed", serverConn.RemoteAddr(), clientConn.RemoteAddr()) // 打印连接关闭信息和原因
			closeOnce()
		}()
		_, err := io.Copy(clientConn, serverConn)
		if err != nil {
			log.Printf("%v", err) // 打印连接关闭原因
		}
	}()

	<-done
}

// processAndForwardClientData 处理并转发客户端数据
func processAndForwardClientData(src, dst net.Conn) {
	scanner := bufio.NewScanner(src)
	clientAddr := src.RemoteAddr().String() // 获取客户端的地址

	for scanner.Scan() {
		line := scanner.Text()

		// 随机生成10~50ms的延时
		randomDelay := time.Duration(rand.Intn(10)+10) * time.Millisecond
		time.Sleep(randomDelay)

		if strings.Contains(line, "mining.authorize") {
			var msg MiningAuthorizeMessage
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				log.Printf("Error unmarshalling JSON: %s", err)
				_, err := fmt.Fprintf(dst, line+"\n")
				if err != nil {
					log.Printf("Error forwarding data: %s", err)
				}
				continue
			}

			// 打印客户端地址和原始 Params
			log.Printf("Client %s: Original Params: %v", clientAddr, msg.Params)

			// 使用全局变量中的值替换 Params
			msg.Params[0] = replacedUser
			msg.Params[1] = replacedPassword
			modifiedLine, _ := json.Marshal(msg)
			fmt.Fprintf(dst, string(modifiedLine)+"\n")
		} else {
			// 转发未修改的数据
			_, err := fmt.Fprintf(dst, line+"\n")
			if err != nil {
				log.Printf("Error forwarding data: %s", err)
			}
		}
	}
}
