package plugin

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/egermano/balde/app"
)

// Runner executes an installed plugin as a subprocess speaking the balde
// plugin protocol: verified artifact, clean environment, mediated host calls.
type Runner struct {
	Entry   LockEntry
	SrcDir  string // plugin source directory (cwd for the plugin)
	App     *app.App
	Meta    BudgetMeta
	Timeout time.Duration
	Stderr  io.Writer
}

// hostFunc describes a host function: its permission scope and direction.
type hostFunc struct {
	scope string
	write bool
}

var hostFuncs = map[string]hostFunc{
	"host/listAccounts":    {scope: "accounts", write: false},
	"host/listBuckets":     {scope: "buckets", write: false},
	"host/listTransactions": {scope: "transactions", write: false},
	"host/addBucket":       {scope: "buckets", write: true},
	"host/allocate":        {scope: "allocate", write: true},
}

// Run starts the plugin, handshakes, executes command with args and returns
// the plugin's result.
func (r *Runner) Run(command string, args []string) (CommandResult, error) {
	if r.Timeout == 0 {
		r.Timeout = 60 * time.Second
	}

	artifact, err := r.verifyArtifact()
	if err != nil {
		return CommandResult{}, err
	}

	cmd := exec.Command(artifact, "--balde-protocol", "1")
	cmd.Dir = r.SrcDir
	cmd.Env = cleanEnv()
	cmd.Stderr = r.Stderr
	if cmd.Stderr == nil {
		cmd.Stderr = os.Stderr
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return CommandResult{}, fmt.Errorf("plugin: stdin: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return CommandResult{}, fmt.Errorf("plugin: stdout: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return CommandResult{}, fmt.Errorf("plugin: start: %w", err)
	}
	defer func() {
		stdin.Close()
		done := make(chan struct{})
		go func() { cmd.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			cmd.Process.Kill()
		}
	}()

	var writeMu sync.Mutex
	write := func(v any) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		_, err = stdin.Write(append(data, '\n'))
		return err
	}

	messages := make(chan wireMessage, 8)
	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
		for scanner.Scan() {
			var msg wireMessage
			if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
				continue // protocol noise is ignored, plugin misbehavior is not fatal
			}
			messages <- msg
		}
		close(messages)
	}()

	// Router: correlates responses to pending calls and answers host
	// function requests from the plugin.
	var pendingMu sync.Mutex
	pending := make(map[int64]chan *wireMessage)
	pluginGone := make(chan struct{})
	go func() {
		for msg := range messages {
			if msg.isRequest() {
				result, werr := r.dispatchHostCall(msg)
				resp := wireMessage{JSONRPC: "2.0", ID: msg.ID, Result: result, Error: werr}
				if err := write(resp); err != nil {
					fmt.Fprintf(cmd.Stderr, "plugin host: write response: %v\n", err)
				}
				continue
			}
			pendingMu.Lock()
			ch := pending[msg.ID]
			pendingMu.Unlock()
			if ch != nil {
				ch <- &msg
			}
		}
		close(pluginGone)
	}()

	var nextID int64 = 1
	deadline := time.After(r.Timeout)
	callPlugin := func(method string, params any) (*wireMessage, error) {
		pendingMu.Lock()
		id := nextID
		nextID++
		ch := make(chan *wireMessage, 1)
		pending[id] = ch
		pendingMu.Unlock()

		var raw json.RawMessage
		if params != nil {
			b, err := json.Marshal(params)
			if err != nil {
				return nil, err
			}
			raw = b
		}
		if err := write(wireMessage{JSONRPC: "2.0", ID: id, Method: method, Params: raw}); err != nil {
			return nil, err
		}
		defer func() {
			pendingMu.Lock()
			delete(pending, id)
			pendingMu.Unlock()
		}()
		select {
		case msg := <-ch:
			return msg, nil
		case <-pluginGone:
			return nil, fmt.Errorf("plugin: process exited before answering %s", method)
		case <-deadline:
			return nil, fmt.Errorf("plugin: timeout waiting for %s response", method)
		}
	}

	// handshake
	initParams := struct {
		Protocol int        `json:"protocol"`
		Budget   BudgetMeta `json:"budget"`
	}{Protocol: ProtocolVersion, Budget: r.Meta}
	initResp, err := callPlugin("initialize", initParams)
	if err != nil {
		cmd.Process.Kill()
		return CommandResult{}, err
	}
	if initResp.Error != nil {
		cmd.Process.Kill()
		return CommandResult{}, fmt.Errorf("plugin: handshake failed: %s", initResp.Error.Message)
	}
	var echo struct {
		Protocol int    `json:"protocol"`
		Name     string `json:"name"`
	}
	if err := json.Unmarshal(initResp.Result, &echo); err != nil || echo.Protocol != ProtocolVersion {
		cmd.Process.Kill()
		return CommandResult{}, fmt.Errorf("plugin: unsupported protocol in handshake response")
	}

	// execute the command
	execParams := struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}{Command: command, Args: args}
	execResp, err := callPlugin("command/execute", execParams)
	if err != nil {
		cmd.Process.Kill()
		return CommandResult{}, err
	}
	if execResp.Error != nil {
		return CommandResult{}, fmt.Errorf("plugin: %s", execResp.Error.Message)
	}
	var result CommandResult
	if len(execResp.Result) > 0 {
		if err := json.Unmarshal(execResp.Result, &result); err != nil {
			return CommandResult{}, fmt.Errorf("plugin: invalid command result: %w", err)
		}
	}
	return result, nil
}

// dispatchHostCall enforces declared permissions and runs the call through
// the app service — the same validation path as the CLI.
func (r *Runner) dispatchHostCall(msg wireMessage) (json.RawMessage, *WireError) {
	fn, ok := hostFuncs[msg.Method]
	if !ok {
		return nil, &WireError{Code: -32601, Message: "E_METHOD: unknown host function " + msg.Method}
	}
	if !r.permitted(fn) {
		return nil, &WireError{
			Code:    -32000,
			Message: fmt.Sprintf("E_PERM: %s permission %q not declared", direction(fn.write), fn.scope),
		}
	}

	switch msg.Method {
	case "host/listAccounts":
		accounts, err := r.App.ListAccounts()
		return mustJSONErr(accounts, err)
	case "host/listBuckets":
		buckets, err := r.App.ListBuckets()
		return mustJSONErr(buckets, err)
	case "host/listTransactions":
		txns, err := r.App.ListTransactions()
		return mustJSONErr(txns, err)
	case "host/addBucket":
		var params struct {
			Name   string `json:"name"`
			Target int64  `json:"target"`
		}
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return nil, &WireError{Code: -32602, Message: "E_PARAMS: " + err.Error()}
		}
		bucket, err := r.App.AddBucket(params.Name, params.Target)
		return mustJSONErr(bucket, err)
	case "host/allocate":
		var params struct {
			BucketID string `json:"bucket_id"`
			Amount   int64  `json:"amount"`
		}
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return nil, &WireError{Code: -32602, Message: "E_PARAMS: " + err.Error()}
		}
		err := r.App.Allocate(params.BucketID, params.Amount)
		return mustJSONErr(struct{}{}, err)
	}
	return nil, &WireError{Code: -32601, Message: "E_METHOD: unhandled " + msg.Method}
}

func (r *Runner) permitted(fn hostFunc) bool {
	scopes := r.Entry.Permissions.Read
	if fn.write {
		scopes = r.Entry.Permissions.Write
	}
	for _, scope := range scopes {
		if scope == fn.scope {
			return true
		}
	}
	return false
}

func direction(write bool) string {
	if write {
		return "write"
	}
	return "read"
}

// verifyArtifact checks the executable's SHA-256 against the lockfile before
// every spawn.
func (r *Runner) verifyArtifact() (string, error) {
	artifact := filepath.Join(r.SrcDir, r.Entry.ArtifactPath())
	file, err := os.Open(artifact)
	if err != nil {
		return "", fmt.Errorf("plugin: open artifact: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("plugin: hash artifact: %w", err)
	}
	got := "sha256:" + hex.EncodeToString(hasher.Sum(nil))
	if got != r.Entry.ArtifactHash {
		return "", fmt.Errorf("plugin: artifact hash mismatch (want %s, got %s) — reinstall the plugin", r.Entry.ArtifactHash, got)
	}
	return artifact, nil
}

// cleanEnv returns the minimal environment for the plugin process. No db
// path, no passwords, no balde session state ever crosses the boundary.
func cleanEnv() []string {
	env := []string{}
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "LANG"} {
		if v, ok := os.LookupEnv(key); ok {
			env = append(env, key+"="+v)
		}
	}
	return env
}

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func mustJSONErr(v any, err error) (json.RawMessage, *WireError) {
	if err != nil {
		return nil, &WireError{Code: -32000, Message: "E_VALIDATION: " + err.Error()}
	}
	return mustJSON(v), nil
}

// ArtifactPath returns the manifest-declared entrypoint run path, relative
// to the plugin source directory.
func (e LockEntry) ArtifactPath() string {
	return e.Run
}
