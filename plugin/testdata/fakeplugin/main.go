// Command fakeplugin is a test fixture: a minimal balde plugin that speaks
// the balde plugin protocol over stdin/stdout. Its command names select
// scripted behaviors used by the runtime tests.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type wireMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

var (
	writeMu   sync.Mutex
	stdoutW   *bufio.Writer
	hostMu    sync.Mutex
	pending   = map[int64]chan wireMessage{}
	nextReqID = int64(100) // ids for plugin->host calls
)

func send(v any) {
	b, _ := json.Marshal(v)
	writeMu.Lock()
	defer writeMu.Unlock()
	stdoutW.Write(b)
	stdoutW.WriteByte('\n')
	stdoutW.Flush()
}

// hostCall sends a host function request and waits for its response, which
// the main read loop routes to the pending channel.
func hostCall(method, params string) json.RawMessage {
	hostMu.Lock()
	id := nextReqID
	nextReqID++
	ch := make(chan wireMessage, 1)
	pending[id] = ch
	hostMu.Unlock()

	send(wireMessage{JSONRPC: "2.0", ID: id, Method: method, Params: json.RawMessage(params)})

	resp := <-ch
	if resp.Error != nil {
		return json.RawMessage(fmt.Sprintf(`{"error":%q}`, resp.Error.Message))
	}
	return resp.Result
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	stdoutW = bufio.NewWriter(os.Stdout)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		var msg wireMessage
		if json.Unmarshal([]byte(line), &msg) != nil {
			continue
		}

		if msg.Method == "" {
			// response to one of our host calls
			hostMu.Lock()
			ch := pending[msg.ID]
			delete(pending, msg.ID)
			hostMu.Unlock()
			if ch != nil {
				ch <- msg
			}
			continue
		}

		switch msg.Method {
		case "initialize":
			send(wireMessage{
				JSONRPC: "2.0", ID: msg.ID,
				Result: json.RawMessage(`{"name":"fake","version":"0.1.0","protocol":1,"capabilities":[{"type":"command","name":"fake"}]}`),
			})
		case "command/execute":
			var params struct {
				Command string   `json:"command"`
				Args    []string `json:"args"`
			}
			json.Unmarshal(msg.Params, &params)
			go handleCommand(msg.ID, params.Command, params.Args)
		default:
			send(wireMessage{
				JSONRPC: "2.0", ID: msg.ID,
				Error: &struct {
					Code    int    `json:"code"`
					Message string `json:"message"`
				}{Code: -32601, Message: "method not found: " + msg.Method},
			})
		}
	}
}

func handleCommand(id int64, command string, args []string) {
	switch command {
	case "ok":
		result(id, `{"text":"hello from plugin"}`)
	case "echo-args":
		result(id, fmt.Sprintf(`{"text":%q}`, join(args)))
	case "list-buckets":
		resp := hostCall("host/listBuckets", `{}`)
		result(id, fmt.Sprintf(`{"text":%q}`, "buckets: "+string(resp)))
	case "add-bucket":
		resp := hostCall("host/addBucket", `{"name":"from-plugin","target":100}`)
		result(id, fmt.Sprintf(`{"text":%q}`, "add: "+string(resp)))
	case "fail":
		send(wireMessage{JSONRPC: "2.0", ID: id, Error: &struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}{Code: 1, Message: "E_PLUGIN: planned failure"}})
	default:
		send(wireMessage{JSONRPC: "2.0", ID: id, Error: &struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}{Code: 1, Message: "E_PLUGIN: unknown command " + command}})
	}
}

func join(args []string) string {
	out := ""
	for i, a := range args {
		if i > 0 {
			out += " "
		}
		out += a
	}
	return out
}

func result(id int64, resultJSON string) {
	send(wireMessage{JSONRPC: "2.0", ID: id, Result: json.RawMessage(resultJSON)})
}
