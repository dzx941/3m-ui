package node

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/dzx941/3m-ui/backend/internal/credentials"
	"github.com/dzx941/3m-ui/backend/internal/database/models"
	mihomoConfig "github.com/dzx941/3m-ui/backend/internal/mihomo/config"
)

// ValidateNode is the backend gate for all listener configuration. The
// protocol registry lives in the Mihomo config package so generators and
// validators cannot silently drift apart.
func ValidateNode(l *models.Listener) error {
	l.Name = strings.TrimSpace(l.Name)
	if l.Name == "" { return fmt.Errorf("listener name cannot be empty") }
	proto := strings.ToLower(strings.TrimSpace(l.Protocol)); if proto == "" { proto = strings.ToLower(strings.TrimSpace(l.Type)) }
	if !mihomoConfig.IsMihomoListenerProtocol(proto) { return fmt.Errorf("unsupported Mihomo listener protocol: %s", proto) }
	l.Protocol, l.Type = proto, proto
	if l.Port < 1 || l.Port > 65535 { return fmt.Errorf("port number must be between 1 and 65535") }
	l.BindAddress = strings.TrimSpace(l.BindAddress); if l.BindAddress == "" { l.BindAddress = strings.TrimSpace(l.Listen) }; if l.BindAddress == "" { l.BindAddress = "0.0.0.0" }
	if err := credentials.EnsureListenerCredentials(l); err != nil { return err }
	cfg, err := decodeConfig(l.Config); if err != nil { return err }
	if err := validateSchema(proto, cfg); err != nil { return err }
	return validateProtocolSpecific(proto, cfg)
}

// Legacy helpers remain here for compatibility with older tests and callers.
// New credential generation is centralized in internal/credentials so this
// package no longer owns cross-package credential lifecycle state.
func ensureClientCredentials(l *models.Listener) error { return credentials.EnsureListenerCredentials(l) }
func requiresUserCredentials(proto string) bool { switch strings.ToLower(proto) { case "vless", "vmess", "trojan", "hysteria2", "anytls", "mieru", "shadowquic", "tuic": return true; default: return false } }
func hasExportCredentials(proto string, cfg map[string]interface{}) bool { users, ok := cfg["users"]; if !ok || users == nil { return false }; switch strings.ToLower(proto) { case "hysteria2", "anytls", "mieru", "tuic": m, ok := users.(map[string]interface{}); return ok && len(m)>0; default: list, ok := users.([]interface{}); return ok && len(list)>0 } }
func randomSecret(length int) (string,error) { b:=make([]byte,length); if _,err:=rand.Read(b); err!=nil{return "",err}; return hex.EncodeToString(b),nil }
func randomUUID() (string,error) { b:=make([]byte,16); if _,err:=rand.Read(b);err!=nil{return "",err}; b[6]=(b[6]&0x0f)|0x40; b[8]=(b[8]&0x3f)|0x80; return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",b[0:4],b[4:6],b[6:8],b[8:10],b[10:16]),nil }

func decodeConfig(raw string) (map[string]interface{}, error) { if strings.TrimSpace(raw)=="" { return map[string]interface{}{},nil }; var cfg map[string]interface{}; if err:=json.Unmarshal([]byte(raw),&cfg);err!=nil{return nil,fmt.Errorf("invalid listener configuration: %w",err)}; if cfg==nil{return map[string]interface{}{},nil}; return cfg,nil }
func validateSchema(proto string,cfg map[string]interface{}) error { schema,ok:=mihomoConfig.GetMihomoListenerSchema(proto);if !ok{return fmt.Errorf("no listener schema registered for protocol %s",proto)};flat:=make(map[string]interface{});flattenConfig("",cfg,flat);for key:=range flat{top:=strings.SplitN(key,".",2)[0];if _,ok:=schema.Fields[top];!ok{return fmt.Errorf("%s listener: unsupported field %q",proto,key)}};return nil }
func flattenConfig(prefix string,value interface{},out map[string]interface{}) { if m,ok:=value.(map[string]interface{});ok{for key,child:=range m{next:=key;if prefix!=""{next=prefix+"."+key};if childMap,ok:=child.(map[string]interface{});ok{flattenConfig(next,childMap,out)}else{out[next]=child}};return};if prefix!=""{out[prefix]=value} }
func validateProtocolSpecific(proto string,cfg map[string]interface{}) error { if err:=validateCertificateMode(proto,cfg);err!=nil{return err};switch proto{case "snell":if value,ok:=numeric(cfg["version"]);ok&&(value<1||value>5){return fmt.Errorf("snell version must be between 1 and 5")};case "hysteria2":if obfs,ok:=cfg["obfs"].(string);ok&&obfs!=""&&obfs!="salamander"{return fmt.Errorf("hysteria2 obfs must be salamander")};if !hasCertificatePair(cfg)&&!boolValue(cfg["allow-insecure"]){return fmt.Errorf("hysteria2 listener requires certificate/private-key or allow-insecure")};if users,ok:=cfg["users"].(map[string]interface{});ok&&len(users)==0{return fmt.Errorf("hysteria2 users cannot be empty")};case "anytls":if _,ok:=cfg["reality-config"];ok{return fmt.Errorf("anytls does not support reality-config")};if !hasCertificatePair(cfg)&&!boolValue(cfg["allow-insecure"]){return fmt.Errorf("anytls listener requires certificate/private-key or allow-insecure")};if users,ok:=cfg["users"].(map[string]interface{});ok&&len(users)==0{return fmt.Errorf("anytls users cannot be empty")};case "trusttunnel":if !hasCertificatePair(cfg){return fmt.Errorf("trusttunnel listener requires certificate and private-key")};case "tuic":users:=hasNonEmpty(cfg["users"]);token:=hasNonEmpty(cfg["token"]);if users==token{return fmt.Errorf("tuic listener must configure exactly one of users (TUIC V5) or token (TUIC V4)")};case "vless","vmess":if users,ok:=cfg["users"].([]interface{});ok{for i,user:=range users{if err:=validateUserRow(proto,i,user,true);err!=nil{return err}}};case "trojan","shadowquic":if users,ok:=cfg["users"].([]interface{});ok{for i,user:=range users{if err:=validateUserRow(proto,i,user,false);err!=nil{return err}}};case "mieru":if users,ok:=cfg["users"].(map[string]interface{});ok&&len(users)==0{return fmt.Errorf("mieru users cannot be empty")};case "sudoku":min,minOK:=numeric(cfg["padding-min"]);max,maxOK:=numeric(cfg["padding-max"]);if minOK&&maxOK&&max<min{return fmt.Errorf("sudoku padding-max must be greater than or equal to padding-min")}};return nil }
func validateCertificateMode(proto string,cfg map[string]interface{}) error {cert:=hasString(cfg["certificate"]);key:=hasString(cfg["private-key"])||hasString(cfg["private_key"]);if cert!=key{return fmt.Errorf("%s listener requires certificate and private-key together",proto)};modes:=make([]string,0,5);if cert&&key{modes=append(modes,"certificate")};for _,name:=range []string{"reality-config","shadow-tls","res-tls","jls-config"}{if hasNonEmpty(cfg[name]){modes=append(modes,name)}};if len(modes)>1{return fmt.Errorf("%s listener has mutually exclusive TLS modes configured: %s",proto,strings.Join(modes,", "))};if proto=="anytls"&&hasNonEmpty(cfg["reality-config"]){return fmt.Errorf("anytls listener does not support reality-config")};if (proto=="anytls"||proto=="hysteria2"||proto=="tuic"||proto=="trusttunnel")&&(hasNonEmpty(cfg["shadow-tls"])||hasNonEmpty(cfg["res-tls"])||hasNonEmpty(cfg["jls-config"])){return fmt.Errorf("%s listener does not support the selected TLS alternative",proto)};return nil }
func validateUserRow(proto string,index int,raw interface{},uuidMode bool) error {row,ok:=raw.(map[string]interface{});if !ok{return fmt.Errorf("%s listener users[%d] must be an object",proto,index)};if !hasString(row["username"]){return fmt.Errorf("%s listener users[%d] requires username",proto,index)};if uuidMode{if !hasString(row["uuid"]){return fmt.Errorf("%s listener users[%d] requires uuid",proto,index)};if proto=="vmess"&&hasNonEmpty(row["flow"]){return fmt.Errorf("vmess listener users[%d] does not support flow",index)};if proto=="vless"&&hasNonEmpty(row["alterId"]){return fmt.Errorf("vless listener users[%d] does not support alterId",index)}}else if !hasString(row["password"]){return fmt.Errorf("%s listener users[%d] requires password",proto,index)};return nil }
func hasCertificatePair(cfg map[string]interface{})bool{return hasString(cfg["certificate"])&&(hasString(cfg["private-key"])||hasString(cfg["private_key"]))}
func hasString(value interface{})bool{s,ok:=value.(string);return ok&&strings.TrimSpace(s)!=""}
func hasNonEmpty(value interface{})bool{if value==nil{return false};if s,ok:=value.(string);ok{return strings.TrimSpace(s)!=""};v:=reflect.ValueOf(value);switch v.Kind(){case reflect.Slice,reflect.Map:return v.Len()>0;default:return true}}
func boolValue(v interface{})bool{b,_:=v.(bool);return b}
func numeric(v interface{})(float64,bool){n,ok:=v.(float64);return n,ok}
