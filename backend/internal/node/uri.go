package node

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/dzx941/3m-ui/backend/internal/database/models"
	"golang.org/x/crypto/curve25519"
)

func ClientURIs(listener models.Listener, host string) ([]string, error) {
	host = normalizeExportHost(host, listener.BindAddress, listener.Listen)
	if host == "" { return nil, fmt.Errorf("cannot determine public host for listener") }
	cfg, err := decodeURIConfig(listener.Config); if err != nil { return nil, err }
	port := strconv.Itoa(listener.Port)
	switch strings.ToLower(listener.Protocol) {
	case "shadowsocks": return shadowsocksURIs(listener.Name, host, port, cfg)
	case "vless": return vlessURIs(listener.Name, host, port, cfg)
	case "vmess": return vmessURIs(listener.Name, host, port, cfg)
	case "trojan": return trojanURIs(listener.Name, host, port, cfg)
	case "hysteria2": return hysteria2URIs(listener.Name, host, port, cfg)
	case "tuic": return tuicURIs(listener.Name, host, port, cfg)
	case "anytls": return anytlsURIs(listener.Name, host, port, cfg)
	default: return nil, fmt.Errorf("URI export is not supported for listener protocol %q", listener.Protocol)
	}
}

func decodeURIConfig(raw string) (map[string]interface{}, error) { if strings.TrimSpace(raw) == "" { return map[string]interface{}{}, nil }; var cfg map[string]interface{}; if err := json.Unmarshal([]byte(raw), &cfg); err != nil { return nil, fmt.Errorf("invalid listener configuration: %w", err) }; if cfg == nil { return map[string]interface{}{}, nil }; return cfg, nil }
func normalizeExportHost(requestHost, bind, listen string) string { h := strings.TrimSpace(requestHost); if h != "" { if host, _, err := net.SplitHostPort(h); err == nil { return strings.Trim(host, "[]") }; return strings.Trim(h, "[]") }; for _, candidate := range []string{bind, listen} { candidate = strings.TrimSpace(candidate); if candidate == "" || candidate == "0.0.0.0" || candidate == "::" || candidate == "*" || candidate == "127.0.0.1" || candidate == "::1" { continue }; if host, _, err := net.SplitHostPort(candidate); err == nil { return strings.Trim(host, "[]") }; return strings.Trim(candidate, "[]") }; return "" }
func userMap(cfg map[string]interface{}) map[string]interface{} { if users, ok := cfg["users"].(map[string]interface{}); ok { return users }; return nil }
func userRows(cfg map[string]interface{}) []map[string]interface{} { users, _ := cfg["users"].([]interface{}); rows := make([]map[string]interface{}, 0, len(users)); for _, raw := range users { if row, ok := raw.(map[string]interface{}); ok { rows = append(rows, row) } }; return rows }
func query(base string, values map[string]string) string { q := url.Values{}; for k, v := range values { if v != "" { q.Set(k, v) } }; if encoded := q.Encode(); encoded != "" { return base + "?" + encoded }; return base }

func shadowsocksURIs(name, host, port string, cfg map[string]interface{}) ([]string, error) { cipher, _ := cfg["cipher"].(string); password, _ := cfg["password"].(string); if cipher == "" || password == "" { return nil, fmt.Errorf("shadowsocks listener requires cipher and password for URI export") }; encoded := base64.RawStdEncoding.EncodeToString([]byte(cipher+":"+password)); return []string{"ss://"+encoded+"@"+net.JoinHostPort(host, port)+"#"+url.PathEscape(name)}, nil }

func vlessURIs(name, host, port string, cfg map[string]interface{}) ([]string, error) {
	rows := userRows(cfg); if len(rows) == 0 { return nil, fmt.Errorf("vless listener requires at least one user for URI export") }; result := make([]string, 0, len(rows))
	for _, row := range rows { uuid, _ := row["uuid"].(string); if uuid == "" { return nil, fmt.Errorf("vless user uuid is required") }; params := map[string]string{"type":"tcp", "security":"none"}; if flow, _ := row["flow"].(string); flow != "" { params["flow"] = flow }; if decryption, _ := cfg["decryption"].(string); decryption != "" { params["encryption"] = decryption }; if rc, ok := cfg["reality-config"].(map[string]interface{}); ok { params["security"]="reality"; publicKey, err := realityPublicKey(rc); if err != nil { return nil, err }; params["pbk"]=publicKey; params["sid"], _ = rc["short-id"].(string); params["sni"], _ = firstString(rc["server-names"]) }; if ws, ok := cfg["ws-path"].(string); ok && ws != "" { params["type"]="ws"; params["path"]=ws }; if grpc, ok := cfg["grpc-service-name"].(string); ok && grpc != "" { params["type"]="grpc"; params["serviceName"]=grpc }; result = append(result, query("vless://"+url.PathEscape(uuid)+"@"+net.JoinHostPort(host, port), params)+"#"+url.PathEscape(name)) }
	return result, nil
}
func vmessURIs(name, host, port string, cfg map[string]interface{}) ([]string, error) { rows := userRows(cfg); if len(rows)==0 { return nil, fmt.Errorf("vmess listener requires at least one user for URI export") }; result:=make([]string,0,len(rows)); for _, row:=range rows { obj:=map[string]string{"v":"2","ps":name,"add":host,"port":port,"id":fmt.Sprint(row["uuid"]),"aid":"0","scy":"auto","net":"tcp","type":"none","tls":""}; if ws,ok:=cfg["ws-path"].(string);ok&&ws!=""{obj["net"]="ws";obj["path"]=ws};if grpc,ok:=cfg["grpc-service-name"].(string);ok&&grpc!=""{obj["net"]="grpc";obj["path"]=grpc};data,_:=json.Marshal(obj);result=append(result,"vmess://"+base64.RawStdEncoding.EncodeToString(data)) };return result,nil }
func trojanURIs(name, host, port string, cfg map[string]interface{}) ([]string,error){rows:=userRows(cfg);if len(rows)==0{return nil,fmt.Errorf("trojan listener requires at least one user for URI export")};result:=make([]string,0,len(rows));for _,row:=range rows{password,_:=row["password"].(string);if password==""{return nil,fmt.Errorf("trojan user password is required")};result=append(result,query("trojan://"+url.PathEscape(password)+"@"+net.JoinHostPort(host,port),map[string]string{})+"#"+url.PathEscape(name))};return result,nil}
func hysteria2URIs(name, host, port string, cfg map[string]interface{}) ([]string,error){users:=userMap(cfg);if len(users)==0{return nil,fmt.Errorf("hysteria2 listener requires at least one user for URI export")};result:=make([]string,0,len(users));for user,raw:=range users{password,_:=raw.(string);if password==""{return nil,fmt.Errorf("hysteria2 user %q has empty password",user)};result=append(result,"hysteria2://"+url.PathEscape(user)+":"+url.PathEscape(password)+"@"+net.JoinHostPort(host,port)+"#"+url.PathEscape(name))};return result,nil}
func tuicURIs(name, host, port string, cfg map[string]interface{}) ([]string,error){users:=userMap(cfg);if len(users)==0{return nil,fmt.Errorf("tuic V5 listener requires at least one user for URI export")};result:=make([]string,0,len(users));for user,raw:=range users{password,_:=raw.(string);if password==""{return nil,fmt.Errorf("tuic user %q has empty password",user)};result=append(result,"tuic://"+url.PathEscape(user)+":"+url.PathEscape(password)+"@"+net.JoinHostPort(host,port)+"#"+url.PathEscape(name))};return result,nil}
func anytlsURIs(name, host, port string, cfg map[string]interface{}) ([]string,error){users:=userMap(cfg);if len(users)==0{return nil,fmt.Errorf("anytls listener requires at least one user for URI export")};result:=make([]string,0,len(users));for user,raw:=range users{password,_:=raw.(string);if password==""{return nil,fmt.Errorf("anytls user %q has empty password",user)};result=append(result,"anytls://"+url.PathEscape(user)+":"+url.PathEscape(password)+"@"+net.JoinHostPort(host,port)+"#"+url.PathEscape(name))};return result,nil}

func realityPublicKey(cfg map[string]interface{}) (string,error){if public,ok:=cfg["public-key"].(string);ok&&strings.TrimSpace(public)!=""{return public,nil};private,ok:=cfg["private-key"].(string);if !ok||strings.TrimSpace(private)==""{return "",fmt.Errorf("reality listener URI export requires reality-config.public-key or private-key")};var raw []byte;var err error;for _,decode:=range []func(string)([]byte,error){base64.RawStdEncoding.DecodeString,base64.StdEncoding.DecodeString,base64.RawURLEncoding.DecodeString,base64.URLEncoding.DecodeString}{raw,err=decode(strings.TrimSpace(private));if err==nil&&len(raw)==32{break}};if len(raw)!=32{return "",fmt.Errorf("invalid Reality private key: expected 32 decoded bytes")};public,err:=curve25519.X25519(raw,curve25519.Basepoint);if err!=nil{return "",fmt.Errorf("failed to derive Reality public key: %w",err)};return base64.RawStdEncoding.EncodeToString(public),nil}
func firstString(v interface{})(string,bool){if s,ok:=v.(string);ok{return s,s!=""};if a,ok:=v.([]interface{});ok&&len(a)>0{s,_:=a[0].(string);return s,s!=""};return "",false}
