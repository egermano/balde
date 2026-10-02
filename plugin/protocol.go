package plugin

import "encoding/json"

// Protocol version spoken by the host. Bumping this is a breaking change;
// while the plugin system is experimental it may move without notice.
const ProtocolVersion = 1

// Wire types: newline-delimited JSON-RPC 2.0 over the plugin's stdin/stdout.
// Requests travel both directions; ids correlate responses.

type wireMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *WireError      `json:"error,omitempty"`
}

func (m wireMessage) isRequest() bool { return m.Method != "" }

type WireError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *WireError) Error() string { return e.Message }

// CommandResult is the plugin's final answer to command/execute. Text is
// human output; JSON is the structured form emitted for --json.
type CommandResult struct {
	Text string          `json:"text"`
	JSON json.RawMessage `json:"json,omitempty"`
}

// BudgetMeta is the non-sensitive budget context sent at handshake. It never
// includes the db path, password or any secret.
type BudgetMeta struct {
	CurrencySymbol     string `json:"currency_symbol"`
	DecimalSeparator   string `json:"decimal_separator"`
	ThousandsSeparator string `json:"thousands_separator"`
	Frequency          string `json:"frequency"`
}
