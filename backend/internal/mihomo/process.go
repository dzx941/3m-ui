package mihomo

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

type ProcessManager struct {
	mu sync.Mutex
	cmd *exec.Cmd
	pid int
	startTime time.Time
	binaryPath string
	configPath string
	done chan struct{}
	logs []string
}

var globalPM *ProcessManager
var pmOnce sync.Once
func GetProcessManager(binary, config string) *ProcessManager { pmOnce.Do(func(){ globalPM=&ProcessManager{binaryPath:binary,configPath:config,logs:make([]string,0,200)} }); return globalPM }

func (pm *ProcessManager) GetVersion() (*VersionInfo,error) {
	info,err:=os.Stat(pm.binaryPath);if err!=nil{if os.IsNotExist(err){return nil,fmt.Errorf("mihomo binary not found: %s",pm.binaryPath)};return nil,fmt.Errorf("failed to stat mihomo binary: %w",err)};if info.IsDir(){return nil,fmt.Errorf("mihomo binary path is a directory: %s",pm.binaryPath)}
	cmd:=exec.Command(pm.binaryPath,"-v");var out bytes.Buffer;cmd.Stdout=&out;cmd.Stderr=&out;if err:=cmd.Run();err!=nil{return nil,fmt.Errorf("failed to run mihomo -v: %w",err)}
	output:=strings.TrimSpace(out.String());parts:=strings.Fields(output);version:="unknown";if len(parts)>=2{version=parts[1]};return &VersionInfo{Version:version,Commit:"official-build",Build:output},nil
}

// ValidateConfig runs the real Mihomo parser before a configuration is ever
// activated. This prevents a bad Listener from taking down a working core.
func (pm *ProcessManager) ValidateConfig() error {
	pm.mu.Lock();binaryPath,configPath:=pm.binaryPath,pm.configPath;pm.mu.Unlock()
	if binaryPath==""||configPath==""{return fmt.Errorf("mihomo binary or config path is empty")}
	if info,err:=os.Stat(binaryPath);err!=nil||info.IsDir()||info.Mode()&0111==0{if err!=nil{return fmt.Errorf("mihomo binary unavailable: %w",err)};return fmt.Errorf("mihomo binary is not executable: %s",binaryPath)}
	info,err:=os.Stat(configPath);if err!=nil{return fmt.Errorf("mihomo config not found: %s",configPath)};if info.IsDir()||info.Size()==0{return fmt.Errorf("mihomo config is empty: %s",configPath)}
	cmd:=exec.Command(binaryPath,"-t","-d",filepath.Dir(configPath),"-f",configPath);var out bytes.Buffer;cmd.Stdout=&out;cmd.Stderr=&out
	if err:=cmd.Run();err!=nil{detail:=strings.TrimSpace(out.String());if detail!=""{return fmt.Errorf("mihomo configuration validation failed: %s",detail)};return fmt.Errorf("mihomo configuration validation failed: %w",err)}
	return nil
}

func (pm *ProcessManager) Start() error {
	pm.mu.Lock();defer pm.mu.Unlock();if pm.isRunning(){return fmt.Errorf("mihomo is already running (PID: %d)",pm.pid)}
	info,err:=os.Stat(pm.binaryPath);if err!=nil{if os.IsNotExist(err){return fmt.Errorf("mihomo binary not found: %s",pm.binaryPath)};return fmt.Errorf("failed to stat mihomo binary: %w",err)};if info.IsDir(){return fmt.Errorf("mihomo binary path is a directory: %s",pm.binaryPath)};if info.Mode()&0111==0{return fmt.Errorf("mihomo binary is not executable: %s",pm.binaryPath)}
	if pm.configPath==""{return fmt.Errorf("mihomo config path is empty")};cfgInfo,err:=os.Stat(pm.configPath);if err!=nil{return fmt.Errorf("mihomo config not found: %s",pm.configPath)};if cfgInfo.IsDir()||cfgInfo.Size()==0{return fmt.Errorf("mihomo config is empty: %s",pm.configPath)}
	cmd:=exec.Command(pm.binaryPath,"-d",filepath.Dir(pm.configPath),"-f",pm.configPath);cmd.SysProcAttr=&syscall.SysProcAttr{Setpgid:true};writer:=&processLogWriter{pm:pm};cmd.Stdout=writer;cmd.Stderr=writer
	if err:=cmd.Start();err!=nil{return fmt.Errorf("failed to start mihomo: %w",err)}
	pm.cmd=cmd;pm.pid=cmd.Process.Pid;pm.startTime=time.Now();pm.done=make(chan struct{});done:=pm.done;go func(c *exec.Cmd,finished chan struct{}){err:=c.Wait();pm.mu.Lock();if err!=nil{pm.appendLogLocked(fmt.Sprintf("process exited: %v",err))}else{pm.appendLogLocked("process exited")};pm.mu.Unlock();close(finished)}(cmd,done)
	select{case <-done:pm.cmd=nil;pm.pid=0;pm.startTime=time.Time{};pm.done=nil;return fmt.Errorf("mihomo exited immediately; check logs")
	case <-time.After(500*time.Millisecond):return nil}
}

func (pm *ProcessManager) Stop() error {pm.mu.Lock();if !pm.isRunning(){pm.mu.Unlock();return fmt.Errorf("mihomo is not running")};cmd,pid,done:=pm.cmd,pm.pid,pm.done;pm.mu.Unlock();pgid,err:=syscall.Getpgid(pid);if err==nil&&pgid>0{_ = syscall.Kill(-pgid,syscall.SIGTERM)}else if cmd.Process!=nil{_ = cmd.Process.Signal(syscall.SIGTERM)};select{case <-done:case <-time.After(5*time.Second):if pgid>0{_ = syscall.Kill(-pgid,syscall.SIGKILL)}else if cmd.Process!=nil{_ = cmd.Process.Kill()};<-done};pm.mu.Lock();defer pm.mu.Unlock();if pm.cmd==cmd{pm.cmd=nil;pm.pid=0;pm.startTime=time.Time{};pm.done=nil};return nil}
func (pm *ProcessManager) Restart() error {if pm.IsRunning(){if err:=pm.Stop();err!=nil{return fmt.Errorf("failed to stop before restart: %w",err)}};if err:=pm.ValidateConfig();err!=nil{return err};return pm.Start()}
func (pm *ProcessManager) IsRunning() bool {pm.mu.Lock();defer pm.mu.Unlock();return pm.isRunning()}
func (pm *ProcessManager) isRunning() bool {if pm.cmd==nil||pm.cmd.Process==nil||pm.pid==0{return false};if pm.done!=nil{select{case <-pm.done:return false;default:}};process,err:=os.FindProcess(pm.pid);if err!=nil{return false};return process.Signal(syscall.Signal(0))==nil}
func (pm *ProcessManager) Status()(*StatusResponse,error){pm.mu.Lock();running:=pm.isRunning();pid:=pm.pid;startTime:=pm.startTime;pm.mu.Unlock();versionStr:="unknown";if vInfo,err:=pm.GetVersion();err==nil&&vInfo!=nil{versionStr=vInfo.Version};uptime:="0s";if running&&!startTime.IsZero(){uptime=formatDuration(time.Since(startTime))};return &StatusResponse{Running:running,Version:versionStr,PID:pid,Uptime:uptime},nil}
func (pm *ProcessManager) Logs()[]string{pm.mu.Lock();defer pm.mu.Unlock();out:=make([]string,len(pm.logs));copy(out,pm.logs);return out}
func (pm *ProcessManager) appendLogLocked(line string){line=strings.TrimSpace(line);if line==""{return};pm.logs=append(pm.logs,time.Now().Format(time.RFC3339)+" "+line);if len(pm.logs)>200{pm.logs=pm.logs[len(pm.logs)-200:]}}
type processLogWriter struct{pm *ProcessManager}
func(w *processLogWriter)Write(p []byte)(int,error){lines:=strings.Split(string(p),"\n");w.pm.mu.Lock();defer w.pm.mu.Unlock();for _,line:=range lines{w.pm.appendLogLocked(line)};return len(p),nil}
func formatDuration(d time.Duration)string{d=d.Round(time.Second);h:=d/time.Hour;d-=h*time.Hour;m:=d/time.Minute;d-=m*time.Minute;s:=d/time.Second;if h>0{return fmt.Sprintf("%dh %dm %ds",h,m,s)};if m>0{return fmt.Sprintf("%dm %ds",m,s)};return fmt.Sprintf("%ds",s)}
