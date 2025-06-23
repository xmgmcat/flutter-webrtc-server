package main

import (
	"os"

	"github.com/flutter-webrtc/flutter-webrtc-server/pkg/logger"
	"github.com/flutter-webrtc/flutter-webrtc-server/pkg/signaler"
	"github.com/flutter-webrtc/flutter-webrtc-server/pkg/turn"
	"github.com/flutter-webrtc/flutter-webrtc-server/pkg/websocket"
	"gopkg.in/ini.v1"
)

// 根据配置文件初始化并配置 TURN 服务器、信令服务器和 WebSocket 服务器，
// 同时设置 SSL 证书并绑定 WebSocket 服务器。
func main() {

	// 加载配置文件。
	cfg, err := ini.Load("configs/config.ini")
	if err != nil {
		logger.Errorf("读取配置文件失败: %v", err)
		os.Exit(1)
	}

	// 从配置文件中更新 TURN 服务器的相关配置。
	publicIP := cfg.Section("turn").Key("public_ip").String()
	stunPort, err := cfg.Section("turn").Key("port").Int()
	if err != nil {
		stunPort = 19302 // 如果配置文件中不存在指定端口，则使用默认的 STUN 端口。
	}
	realm := cfg.Section("turn").Key("realm").String()

	// 使用解析的配置初始化 TURN 服务器。
	turnConfig := turn.DefaultConfig()
	turnConfig.PublicIP = publicIP
	turnConfig.Port = stunPort
	turnConfig.Realm = realm
	turn := turn.NewTurnServer(turnConfig)

	// 使用 TURN 服务器实例初始化信令服务器。
	signaler := signaler.NewSignaler(turn)

	// 初始化 WebSocket 服务器，并设置处理新 WebSocket 连接和 TURN 凭证请求的处理器。
	wsServer := websocket.NewWebSocketServer(signaler.HandleNewWebSocket, signaler.HandleTurnServerCredentials)

	// 从配置文件中解析通用服务器配置。
	sslCert := cfg.Section("general").Key("cert").String()
	sslKey := cfg.Section("general").Key("key").String()
	bindAddress := cfg.Section("general").Key("bind").String()

	port, err := cfg.Section("general").Key("port").Int()
	if err != nil {
		port = 6656 // 如果配置文件中不存在指定端口，则使用默认端口。
	}

	htmlRoot := cfg.Section("general").Key("html_root").String()

	// 更新server配置初始化 WebSocket 服务器。
	config := websocket.DefaultConfig()
	config.Host = bindAddress
	config.Port = port
	config.CertFile = sslCert
	config.KeyFile = sslKey
	config.HTMLRoot = htmlRoot

	// 将 WebSocket 服务器绑定到指定的地址和端口。
	wsServer.Bind(config)
}
