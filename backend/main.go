package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"smart-chat/internal/handler"
	"smart-chat/internal/store"
)

func main() {
	// 命令行参数
	port := flag.String("port", "8000", "服务端口")
	dbPath := flag.String("db", "./data/smart_chat.db", "数据库路径")
	apiKey := flag.String("api-key", "", "MiniMax API Key")
	baseURL := flag.String("base-url", "https://api.minimax.io/v1", "MiniMax API 地址")
	modelName := flag.String("model", "MiniMax-M2.7", "模型名称")
	corsOrigin := flag.String("cors", "*", "CORS 允许的源")
	flag.Parse()

	// 从环境变量覆盖
	if key := os.Getenv("MINIMAX_API_KEY"); key != "" && *apiKey == "" {
		*apiKey = key
	}
	if url := os.Getenv("MINIMAX_BASE_URL"); url != "" {
		*baseURL = url
	}

	// 确保数据库目录存在
	if err := os.MkdirAll(filepath.Dir(*dbPath), 0755); err != nil {
		log.Fatalf("创建数据目录失败: %v", err)
	}

	// 初始化存储
	s, err := store.NewStore(*dbPath)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer s.Close()
	log.Printf("✅ 数据库初始化完成: %s", *dbPath)

	// 创建处理器
	h, err := handler.NewHandler(s, *apiKey, *baseURL, *modelName)
	if err != nil {
		log.Fatalf("创建处理器失败: %v", err)
	}
	log.Printf("✅ AI 模型初始化完成: %s", *modelName)

	// 创建 Fiber 应用
	app := fiber.New(fiber.Config{
		AppName: "智能客服 API v1.0",
	})

	// 中间件
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: *corsOrigin,
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// 路由
	api := app.Group("/api")

	// 健康检查
	api.Get("/health", h.Health)

	// 知识库管理
	kb := api.Group("/knowledge")
	kb.Post("/", h.AddKnowledge)
	kb.Get("/", h.ListKnowledge)
	kb.Put("/:id", h.UpdateKnowledge)
	kb.Delete("/:id", h.DeleteKnowledge)

	// 聊天
	api.Post("/chat", h.Chat)
	api.Post("/chat/stream", h.StreamChat)

	// 对话
	api.Get("/conversations", h.ListConversations)
	api.Get("/conversations/:id", h.GetConversation)

	// 统计
	api.Get("/stats", h.GetStats)

	// 启动服务
	log.Printf("🚀 智能客服服务已启动: http://localhost:%s", *port)
	log.Printf("📚 API 文档: http://localhost:%s/api/health", *port)
	if err := app.Listen(":" + *port); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}
