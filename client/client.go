package client

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	go_config "github.com/pefish/go-config"
	go_logger "github.com/pefish/go-logger"
	"io"
	"net"
	"os"
	"path"
	"time"
)

type Client struct {

}

func NewClient() *Client {
	return &Client{

	}
}

const (
	MAX_DATASIZE_PER_PACKAGE = 100 * 1024
)

func (c *Client) DecorateFlagSet(flagSet *flag.FlagSet) error {
	flagSet.String("tcp-address", "0.0.0.0:8000", "<addr>:<port> to listen on for TCP clients")
	flagSet.String("file", "", "file to send")
	return nil
}

func (c *Client) ParseFlagSet(flagSet *flag.FlagSet) error {
	err := flagSet.Parse(os.Args[2:])
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Start() error {
	filepath := go_config.Config.MustGetString("file")
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return err
	}
	var filename string
	newFilepath := filepath
	if fileInfo.IsDir() {
		filename = fmt.Sprintf("%s.tar.gz", time.Now().Format("2006-01-02_15-04-05"))
		fp, _ := os.Getwd()
		newFilepath = path.Join(fp, "temp", filename)
		MustPack(filepath, newFilepath)
	} else {
		filename = fileInfo.Name()
	}

	go_logger.Logger.InfoF("filename: %s, newFilepath: %s", filename, newFilepath)
	tcpAddress, err := go_config.Config.GetString("tcp-address")
	if err != nil {
		go_logger.Logger.ErrorF("get config error - %s", err)
		return err
	}
	conn, err := net.Dial("tcp", tcpAddress)
	if err != nil {
		return err
	}
	defer func() {
		conn.Close()
		go_logger.Logger.Info("connection closed!!!")
	}()
	var zeroTime time.Time
	err = conn.SetDeadline(zeroTime)
	if err != nil {
		return err
	}

	// 选择出最合适的数据包大小
	fileInfo, err = os.Stat(newFilepath)
	if err != nil {
		return err
	}
	fileSize := fileInfo.Size()
	go_logger.Logger.DebugF("fileSize: %d", fileSize)
	var dataSizePerPackage uint64
	if fileSize > MAX_DATASIZE_PER_PACKAGE {
		dataSizePerPackage = MAX_DATASIZE_PER_PACKAGE
	} else {
		dataSizePerPackage = uint64(fileSize)
	}
	go_logger.Logger.DebugF("dataSizePerPackage: %d", dataSizePerPackage)

	//发送数据包大小以及文件名到接收端
	var promiseBuf bytes.Buffer
	err = binary.Write(&promiseBuf, binary.BigEndian, dataSizePerPackage)
	if err != nil {
		return err
	}
	//return nil
	promiseBuf.Write([]byte(filename))
	go_logger.Logger.Debug("filename and dataSize package: ", promiseBuf.Bytes())
	_, err = conn.Write(promiseBuf.Bytes())
	if err != nil {
		return err
	}
	resultBuf := make([]byte, 10)
	//接收服务器返还的指令
	n, err := conn.Read(resultBuf)
	if err != nil {
		return err
	}
	//返回ok，可以传输文件
	result := string(resultBuf[:n])
	go_logger.Logger.InfoF("filename result: %s", result)
	if result != "ok" {
		return fmt.Errorf("deny!!! - %s", result)
	}
	go_logger.Logger.InfoF("开始传输文件 - %s", newFilepath)

	file, err := os.Open(newFilepath)
	if err != nil {
		return err
	}
	defer func() {
		file.Close()
		go_logger.Logger.Info("文件已关闭")
	}()
	go_logger.Logger.Info("文件已打开")
	buf := make([]byte, dataSizePerPackage)
	for {
		readCount, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				go_logger.Logger.Info("文件已读完", readCount)
				break
			}
			return err
		}
		var dataSize = uint64(readCount)
		var packageBuf bytes.Buffer
		err = binary.Write(&packageBuf, binary.BigEndian, dataSize)
		if err != nil {
			return err
		}
		go_logger.Logger.DebugF("dataSize: %d, dataSizePerPackage: %d", dataSize, dataSizePerPackage)
		packageBuf.Write(buf)  // 要全部写过去，保证packageBuf大小足够一个包的长度，避免服务端最后一个包阻塞
		go_logger.Logger.DebugF("发送 %d 个字节", len(packageBuf.Bytes()))
		_, err = conn.Write(packageBuf.Bytes())
		if err != nil {
			return err
		}

	}
	go_logger.Logger.Info("文件发送完成，现在发送结束标志")
	buf = make([]byte, dataSizePerPackage)
	var dataSize = uint64(0)
	var packageBuf bytes.Buffer
	err = binary.Write(&packageBuf, binary.BigEndian, dataSize)
	if err != nil {
		return err
	}
	packageBuf.Write(buf)
	_, err = conn.Write(packageBuf.Bytes())
	if err != nil {
		return err
	}

	go_logger.Logger.Info("等待回复接收完成")
	n, err = conn.Read(resultBuf)
	if err != nil {
		return err
	}
	//返回ok，可以传输文件
	result = string(resultBuf[:n])
	go_logger.Logger.InfoF("result: %s", result)
	if result != "done" {
		return fmt.Errorf("没有收到接收成功的回复!!! - %s", result)
	}
	go_logger.Logger.Info("文件传输成功完成")
	return nil
}
