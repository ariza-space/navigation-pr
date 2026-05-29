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

//go:embed web/dist/*
var staticFiles embed.FS

func main() {
	cfg := config.ParseFlags()

	store, err := storage.NewSQLiteSiteStore(cfg.DataPath, cfg.LegacyJSONPath)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer store.Close()

	svc := service.NewSiteService(store)
	noteFiles, err := storage.NewNoteFileStore(cfg.DataDir)
	if err != nil {
		log.Fatalf("初始化笔记目录失败: %v", err)
	}
	noteSvc := service.NewNoteService(store, noteFiles)
	authSvc, err := service.NewAuthService(store)
	if err != nil {
		log.Fatalf("初始化账号失败: %v", err)
	}
	if cfg.ResetAuth {
		if err := authSvc.ResetDefaultUser(); err != nil {
			log.Fatalf("重置账号失败: %v", err)
		}
		if credential, ok := authSvc.InitialCredential(); ok {
			log.Printf("账号密码已重置为: %s/%s", credential.Username, credential.Password)
		}
		return
	}
	if credential, ok := authSvc.InitialCredential(); ok {
		log.Printf("首次启动已生成初始账号密码: %s/%s", credential.Username, credential.Password)
		log.Print("请登录后立即修改初始密码")
	}
	handler := httptransport.NewHandler(svc, authSvc, noteSvc, staticFiles, httptransport.WithSecureCookies(cfg.SecureCookie))

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("导航站已启动: http://localhost%s", addr)
	log.Printf("SQLite 数据库: %s", cfg.DataPath)
	if err := http.ListenAndServe(addr, handler.Routes()); err != nil {
		log.Fatal(err)
	}
}
