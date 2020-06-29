package server

import (
	"bytes"
	"encoding/binary"
	"flag"
	go_config "github.com/pefish/go-config"
	go_logger "github.com/pefish/go-logger"
	"io"
	"net"
	"os"
	"path"
	"strings"
	"time"
)

type Server struct {

}

func NewServer() *Server {
	return &Server{

	}
}

func (s *Server) DecorateFlagSet(flagSet *flag.FlagSet) error {
	flagSet.String("tcp-address", "0.0.0.0:8000", "<addr>:<port> to listen on for TCP clients")
	flagSet.String("target-path", "~/", "path to save file")
	return nil
}

func (s *Server) ParseFlagSet(flagSet *flag.FlagSet) error {
	err := flagSet.Parse(os.Args[2:])
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) Start() error {
	tcpAddress, err := go_config.Config.GetString("tcp-address")
	if err != nil {
		go_logger.Logger.ErrorF("get config error - %s", err)
		return err
	}
	tcpListener, err := net.Listen("tcp", tcpAddress)
	if err != nil {
		go_logger.Logger.ErrorF("listen (%s) failed - %s", tcpAddress, err)
		return err
	}
	go_logger.Logger.InfoF("listening on %s", tcpListener.Addr())

	for {
		clientConn, err := tcpListener.Accept()
		if err != nil {
			//go_logger.Logger.Error(err)
			if err, ok := err.(net.Error); ok && err.Temporary() {
				continue
			}
			break
		}

		go_logger.Logger.InfoF("new client - %s", clientConn.RemoteAddr())
		go s.receiveFile(clientConn)
	}
	return nil
}

func (s *Server) receiveFile(conn net.Conn) {
	defer func() {
		conn.Close()
		go_logger.Logger.Info("client closed!!!")
	}()
	// 开始接收协商包
	promiseBuf := make([]byte, 200)
	err := conn.SetReadDeadline(time.Now().Add(10 * time.Second)) // 这么久没收到数据，则报超时错
	if err != nil {
		go_logger.Logger.ErrorF("failed to set conn timeout - %s", err)
		return
	}
	n, err := conn.Read(promiseBuf) // 只要有数据就可以读出来，不管数据是否足够
	if err != nil {
		go_logger.Logger.Error(err)
		return
	}
	dataSizePerPackage := uint64(0)
	err = binary.Read(bytes.NewReader(promiseBuf[:8]), binary.BigEndian, &dataSizePerPackage)
	if err != nil {
		go_logger.Logger.Error(err)
		return
	}
	if dataSizePerPackage == 0 {
		go_logger.Logger.Error("dataSizePerPackage不能为0")
		return
	}
	go_logger.Logger.InfoF("dataSizePerPackage: %d",dataSizePerPackage)
	saveToPath := strings.TrimSpace(string(promiseBuf[8:40]))
	go_logger.Logger.InfoF("saveToPath: %s",saveToPath)
	filename := string(promiseBuf[40:n])
	go_logger.Logger.InfoF("filename: %s",filename)
	// 创建文件
	targetPath := go_config.Config.MustGetString("target-path")
	if strings.HasPrefix(targetPath, "~") {
		homePath, _ := os.UserHomeDir()
		targetPath = homePath + targetPath[1:]
	}
	targetDir := path.Join(targetPath, saveToPath)
	err = os.MkdirAll(targetDir, os.ModePerm)
	if err != nil {
		go_logger.Logger.Error(err)
		return
	}
	go_logger.Logger.InfoF("targetDir: %s",targetDir)
	saveToFile := path.Join(targetDir, filename)
	go_logger.Logger.InfoF("saveToFile: %s",saveToFile)
	file, err := os.Create(saveToFile)
	if err != nil {
		go_logger.Logger.Error(err)
		return
	}
	defer file.Close()
	_, err = conn.Write([]byte("ok"))
	if err != nil {
		go_logger.Logger.Error(err)
		return
	}

	// 开始接收数据
	// 数据包大小（8个字节）+数据
	buf := make([]byte, dataSizePerPackage + 8)
	for {
		err := conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		if err != nil {
			go_logger.Logger.ErrorF("failed to set conn timeout - %s", err)
			return
		}
		_, err = io.ReadFull(conn, buf)  // 这里的n不是指收到的服务端的数据个数，而是buf的容量
		if err != nil {
			go_logger.Logger.Error(err)
			return
		}
		// 读出数据大小
		var dataSize uint64
		err = binary.Read(bytes.NewReader(buf[:8]), binary.BigEndian, &dataSize)
		if err != nil {
			go_logger.Logger.Error(err)
			return
		}
		go_logger.Logger.DebugF("dataSize: %d, dataSizePerPackage: %d", dataSize, dataSizePerPackage)
		if dataSize != 0 {
			file.Write(buf[8:8+dataSize])
		}
		if dataSize < dataSizePerPackage {
			break
		}
	}

	// 发送done
	go_logger.Logger.Info("文件接收完成")
	_, err = conn.Write([]byte("done"))
	if err != nil {
		go_logger.Logger.Error(err)
		return
	}
}

