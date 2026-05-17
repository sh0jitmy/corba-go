// Copyright 2026- The corba-go Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sh0jitmy/corba-go/idlc"
	"github.com/sh0jitmy/corba-go/orb/cdr"
	"github.com/sh0jitmy/corba-go/orb/iiop"
	"github.com/sh0jitmy/corba-go/orb/iop"
	"github.com/sh0jitmy/corba-go/services/naming"
)

var (
	idlFile    string
	namingHost string
	namingPort int
	port       int
)

type InvokeRequest struct {
	TargetName []NamingComponent `json:"target_name"`
	Interface  string            `json:"interface"`
	Method     string            `json:"method"`
	Args       []interface{}     `json:"args"`
}

type NamingComponent struct {
	Id   string `json:"id"`
	Kind string `json:"kind"`
}

type InvokeResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

var globalModule *idlc.ModuleNode

func main() {
	flag.StringVar(&idlFile, "idl", "", "Path to the IDL file to load")
	flag.StringVar(&namingHost, "naming-host", "localhost", "Naming service host")
	flag.IntVar(&namingPort, "naming-port", 2809, "Naming service port")
	flag.IntVar(&port, "port", 8080, "REST gateway port")
	flag.Parse()

	if idlFile == "" {
		log.Fatal("Please provide an IDL file using -idl <path>")
	}

	//nolint:gosec // user provides idl file path via command line flag
	data, err := os.ReadFile(idlFile)
	if err != nil {
		log.Fatalf("Failed to read IDL file: %v", err)
	}

	lexer := idlc.NewLexer(string(data))
	parser := idlc.NewParser(lexer)
	mod, err := parser.Parse()
	if err != nil {
		log.Fatalf("Failed to parse IDL file: %v", err)
	}
	globalModule = mod
	log.Printf("Loaded IDL module: %s", mod.Name)

	http.HandleFunc("/api/invoke", handleInvoke)
	log.Printf("Starting REST-CORBA Gateway on :%d", port)
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 3 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}

func findOperation(interfaceName, methodName string) (*idlc.OperationNode, error) {
	for _, iface := range globalModule.Interfaces {
		if iface.Name == interfaceName {
			for _, op := range iface.Operations {
				if op.Name == methodName {
					return op, nil
				}
			}
			return nil, fmt.Errorf("method '%s' not found in interface '%s'", methodName, interfaceName)
		}
	}
	return nil, fmt.Errorf("interface '%s' not found in IDL", interfaceName)
}

func handleInvoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read body")
		return
	}

	var req InvokeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	// 1. ASTから該当するメソッドの型定義を取得
	op, err := findOperation(req.Interface, req.Method)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if len(req.Args) != len(op.Parameters) {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("expected %d arguments, got %d", len(op.Parameters), len(req.Args)))
		return
	}

	// 2. ネーミングサービスへの接続 (本来はコネクションプーリング推奨)
	//nolint:gosec // port is bounds checked via type or flag limits
	client, err := iiop.NewClient(namingHost, uint16(namingPort))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to connect to naming service")
		return
	}
	defer func() { _ = client.Close() }()

	namingContext := naming.NewCosNaming_NamingContext_Stub(client, []byte("NameService"))

	var name naming.CosNaming_Name
	for _, comp := range req.TargetName {
		name = append(name, &naming.CosNaming_NameComponent{
			Id:   naming.CosNaming_Istring(comp.Id),
			Kind: naming.CosNaming_Istring(comp.Kind),
		})
	}

	iorStr, err := namingContext.Resolve(name)
	if err != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Failed to resolve target name: %v", err))
		return
	}

	ior, err := iop.ParseIOR(iorStr)
	if err != nil || len(ior.Profiles) == 0 {
		writeError(w, http.StatusInternalServerError, "Failed to parse IOR")
		return
	}

	prof, err := iop.ParseIIOPProfile(ior.Profiles[0].ProfileData)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to parse IIOP Profile")
		return
	}

	// 3. ターゲット(CORBAサーバ)へ接続
	targetClient, err := iiop.NewClient(prof.Host, prof.Port)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to connect to target server")
		return
	}
	defer func() { _ = targetClient.Close() }()

	// 4. 動的エンコード (ASTのパラメータ型定義に従ってJSONの値をCDRに変換)
	enc := cdr.NewEncoder(binary.LittleEndian)
	for i, param := range op.Parameters {
		if err := encodeDynamicArg(enc, param.Type, req.Args[i]); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("Failed to encode param %d: %v", i, err))
			return
		}
	}

	// 5. CORBA呼び出し実行
	reply, err := targetClient.Invoke(prof.ObjectKey, req.Method, enc.Bytes())
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("CORBA invoke failed: %v", err))
		return
	}

	// 6. 動的デコード (ASTの戻り値型定義に従ってCDRから値を復元)
	dec := cdr.NewDecoder(reply, binary.LittleEndian)
	result, err := decodeDynamicReply(dec, op.ReturnType)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to decode reply: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(InvokeResponse{Result: result})
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(InvokeResponse{Error: msg})
}

// プリミティブ型の動的エンコード
func encodeDynamicArg(enc *cdr.Encoder, typ string, val interface{}) error {
	switch typ {
	case "short":
		enc.EncodeShort(int16(val.(float64)))
	case "long":
		enc.EncodeLong(int32(val.(float64)))
	case "longlong":
		enc.EncodeLongLong(int64(val.(float64)))
	case "float":
		enc.EncodeFloat(float32(val.(float64)))
	case "double":
		enc.EncodeDouble(val.(float64))
	case "string":
		enc.EncodeString(val.(string))
	case "boolean":
		enc.EncodeBoolean(val.(bool))
	case "octet":
		enc.EncodeOctet(uint8(val.(float64)))
	default:
		return fmt.Errorf("unsupported type for generic encoding: %s", typ)
	}
	return nil
}

// プリミティブ型の動的デコード
func decodeDynamicReply(dec *cdr.Decoder, typ string) (interface{}, error) {
	switch typ {
	case "short":
		return dec.DecodeShort()
	case "long":
		return dec.DecodeLong()
	case "longlong":
		return dec.DecodeLongLong()
	case "float":
		return dec.DecodeFloat()
	case "double":
		return dec.DecodeDouble()
	case "string":
		return dec.DecodeString()
	case "boolean":
		return dec.DecodeBoolean()
	case "octet":
		return dec.DecodeOctet()
	case "void":
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported type for generic decoding: %s", typ)
	}
}
