package main

import (
	"github.com/pefish/file-transfer-tool/client"
	"github.com/pefish/file-transfer-tool/version"
	"github.com/pefish/go-commander"
	"github.com/pefish/file-transfer-tool/server"
)

func main() {
	commanderInstance := commander.NewCommander(version.AppName, version.Version, "file-transfer-tool 是一款文件传输工具，祝你玩得开心。作者：pefish")
	commanderInstance.RegisterSubcommand("client", client.NewClient())
	commanderInstance.RegisterSubcommand("server", server.NewServer())
	commanderInstance.Run()
}
