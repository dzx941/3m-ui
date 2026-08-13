package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dzx941/3m-ui/backend/internal/auth"
	"github.com/dzx941/3m-ui/backend/internal/config"
	dbconfig "github.com/dzx941/3m-ui/backend/internal/mihomo/config"
	"github.com/dzx941/3m-ui/backend/internal/database"
	"github.com/dzx941/3m-ui/backend/internal/listener"
	"github.com/dzx941/3m-ui/backend/internal/mihomo"
	"github.com/dzx941/3m-ui/backend/internal/node"
	"github.com/dzx941/3m-ui/backend/internal/router"
	"github.com/dzx941/3m-ui/backend/internal/security"
	"github.com/dzx941/3m-ui/backend/internal/traffic"
	"github.com/dzx941/3m-ui/backend/internal/user"
	"github.com/gin-gonic/gin"
)

//go:embed web/dist/*
var frontendFiles embed.FS

func main(){
	if len(os.Args)>1&&(os.Args[1]=="--version"||os.Args[1]=="version"){fmt.Println(versionString());return}
	configPath:=defaultConfigPath();cfg,err:=config.LoadConfig(configPath);if err!=nil{log.Fatalf("load config: %v",err)}
	if _,err:=database.InitDB(cfg.Database.Path);err!=nil{log.Fatalf("initialize database: %v",err)}
	if created,username,password,err:=auth.EnsureAdmin(database.GlobalDB,cfg.Database.Path);err!=nil{log.Fatalf("initialize administrator: %v",err)}else if created{log.Printf("initial administrator created: username=%s",username);passwordFile:=filepath.Join(filepath.Dir(cfg.Database.Path),".initial_admin_password");if err:=os.WriteFile(passwordFile,[]byte(password+"\n"),0600);err!=nil{log.Printf("warning: could not write initial admin password file: %v",err)}else{log.Printf("initial administrator password saved to %s",passwordFile)}}
	security.InitCredentialKey(cfg.Security.CredentialKey)
	mihomo.InitService(cfg);listener.InitService(database.GlobalDB,cfg.Mihomo.Config);node.InitService(database.GlobalDB,cfg.Mihomo.Config);user.InitService(database.GlobalDB)
	dbconfig.CredentialProvider=func()(map[uint][]dbconfig.Credential,error){if user.GlobalService==nil{return map[uint][]dbconfig.Credential{},nil};provided,err:=user.GlobalService.ActiveCredentialsByListener();if err!=nil{return nil,err};result:=make(map[uint][]dbconfig.Credential,len(provided));for listenerID,credentials:=range provided{converted:=make([]dbconfig.Credential,0,len(credentials));for _,credential:=range credentials{converted=append(converted,dbconfig.Credential{Username:credential.Username,Password:credential.Password,UUID:credential.UUID})};result[listenerID]=converted};return result,nil}

	// Build and validate the actual Mihomo configuration before serving HTTP.
	// The running core is therefore always backed by the same database-driven
	// Listener configuration exposed by the panel.
	engine:=dbconfig.NewConfigEngine(database.GlobalDB);generatedConfig,err:=engine.GenerateFinalConfig();if err!=nil{log.Fatalf("generate Mihomo configuration: %v",err)}
	if mihomo.GlobalService==nil{log.Fatal("initialize Mihomo service: service is nil")}
	if err:=mihomo.GlobalService.SaveConfig(generatedConfig);err!=nil{log.Fatalf("validate Mihomo configuration: %v",err)}
	if err:=mihomo.GlobalService.StartMihomo();err!=nil{log.Fatalf("start Mihomo core: %v",err)}
	log.Printf("Mihomo core started successfully")

	traffic.InitGlobalService();r:=router.SetupRouter(cfg);mountFrontend(r);addr:=fmt.Sprintf(":%d",cfg.Server.Port);log.Printf("3m-ui listening on %s",addr);if err:=r.Run(addr);err!=nil{log.Fatalf("run server: %v",err)}
}
func defaultConfigPath()string{if value:=os.Getenv("THREE_M_UI_CONFIG");value!=""{return value};if _,err:=os.Stat("/etc/3m-ui/config.yaml");err==nil{return "/etc/3m-ui/config.yaml"};return "backend/config/config.yaml"}
func mountFrontend(r *gin.Engine){staticFS,err:=fs.Sub(frontendFiles,"web/dist");if err!=nil{log.Printf("frontend assets unavailable: %v",err);return};fileServer:=http.FileServer(http.FS(staticFS));r.RedirectTrailingSlash=false;r.RedirectFixedPath=false;r.NoRoute(func(c *gin.Context){path:=c.Request.URL.Path;if len(path)>=4&&path[:4]=="/api"{c.Status(http.StatusNotFound);return};if path=="/"{c.Data(http.StatusOK,"text/html; charset=utf-8",mustReadFile(staticFS,"index.html"));return};f,err:=staticFS.Open(path[1:]);if err==nil{defer f.Close();fileServer.ServeHTTP(c.Writer,c.Request);return};c.Data(http.StatusOK,"text/html; charset=utf-8",mustReadFile(staticFS,"index.html"))})}
func mustReadFile(fsys fs.FS,name string)[]byte{data,err:=fs.ReadFile(fsys,name);if err!=nil{log.Printf("read frontend %s failed: %v",name,err);return []byte("3m-ui frontend unavailable")};return data}
