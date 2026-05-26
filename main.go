package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"

	"navigation/internal/config"
	"navigation/internal/service"
	"navigation/internal/storage"
	httptransport "navigation/internal/transport/http"
)

//go:embed index.html
var staticFiles embed.FS

func main() {
	cfg := config.ParseFlags()

	store, err := storage.NewSQLiteSiteStore(cfg.DataPath, cfg.LegacyJSONPath)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer store.Close()

	svc := service.NewSiteService(store)
	handler := httptransport.NewHandler(svc, staticFiles)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("导航站已启动: http://localhost%s", addr)
	log.Printf("SQLite 数据库: %s", cfg.DataPath)
	if err := http.ListenAndServe(addr, handler.Routes()); err != nil {
		log.Fatal(err)
	}
}
