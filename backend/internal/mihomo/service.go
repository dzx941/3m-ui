package mihomo

import (
	"fmt"
	"time"

	"github.com/dzx941/3m-ui/backend/internal/config"
)

type Service struct { pm *ProcessManager; cm *ConfigManager }
var GlobalService *Service
func InitService(cfg *config.Config){GlobalService=&Service{pm:GetProcessManager(cfg.Mihomo.Binary,cfg.Mihomo.Config),cm:NewConfigManager(cfg.Mihomo.Config)}}
func (s *Service) StartMihomo()error{if s==nil||s.pm==nil{return fmt.Errorf("mihomo service not initialized")};if err:=s.pm.ValidateConfig();err!=nil{return err};return s.pm.Start()}
func (s *Service) StopMihomo()error{if s==nil||s.pm==nil{return fmt.Errorf("mihomo service not initialized")};return s.pm.Stop()}
func (s *Service) RestartMihomo()error{if s==nil||s.pm==nil{return fmt.Errorf("mihomo service not initialized")};return s.pm.Restart()}
func (s *Service) GetStatus()(*StatusResponse,error){if s==nil||s.pm==nil{return nil,fmt.Errorf("mihomo service not initialized")};return s.pm.Status()}
func (s *Service) SaveConfig(content string)error{if s==nil||s.cm==nil{return fmt.Errorf("mihomo service not initialized")};if err:=s.cm.SaveConfig(content);err!=nil{return err};return s.pm.ValidateConfig()}
func (s *Service) ApplyConfig(content string)error{if s==nil||s.pm==nil||s.cm==nil{return fmt.Errorf("mihomo service not initialized")};if err:=s.cm.SaveConfig(content);err!=nil{return err};if err:=s.pm.ValidateConfig();err!=nil{return err};if s.pm.IsRunning(){return s.pm.Restart()};return s.pm.Start()}
func (s *Service) GetLogs()([]LogResponse,error){if s==nil||s.pm==nil{return nil,fmt.Errorf("mihomo service not initialized")};lines:=s.pm.Logs();result:=make([]LogResponse,0,len(lines));for _,line:=range lines{result=append(result,LogResponse{Timestamp:time.Now(),Level:"info",Payload:line})};return result,nil}
