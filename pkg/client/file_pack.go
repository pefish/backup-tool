package client

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func MustPack(targetPath string, dst string) {
	err := os.MkdirAll(filepath.Dir(dst), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	fw, err := os.Create(dst)
	if err != nil {
		log.Fatal(err)
	}
	defer fw.Close()
	tw := tar.NewWriter(fw)
	defer tw.Close()
	err = filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("walk error - %s\n", err)
			return err
		}
		if path == targetPath {
			return nil
		}
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			fmt.Printf("FileInfoHeader error - %s\n", err)
			return err
		}
		hdr.Name = strings.TrimPrefix(path, strings.TrimPrefix(targetPath, "./"))
		if err := tw.WriteHeader(hdr); err != nil {
			fmt.Printf("WriteHeader error - %s\n", err)
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		// 打开文件
		fr, err := os.Open(path)
		defer fr.Close()
		if err != nil {
			fmt.Printf("Open error - %s\n", err)
			return err
		}

		// copy 文件数据到 tw
		_, err = io.Copy(tw, fr)
		if err != nil {
			fmt.Printf("Copy error - %s\n", err)
			return err
		}

		//log.Printf("成功打包 %s ，共写入了 %d 字节的数据\n", path, n)

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
