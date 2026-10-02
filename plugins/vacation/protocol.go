package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

const protocolVersion = 1

type message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type client struct {
	reader *bufio.Reader
	writer *bufio.Writer
	nextID int64
}

func newClient(in io.Reader, out io.Writer) *client {
	return &client{reader: bufio.NewReader(in), writer: bufio.NewWriter(out), nextID: 100}
}

func (c *client) read() (message, error) {
	line, err := c.reader.ReadBytes('\n')
	if err != nil {
		return message{}, err
	}
	var msg message
	if err := json.Unmarshal(line, &msg); err != nil {
		return message{}, fmt.Errorf("invalid protocol message: %w", err)
	}
	return msg, nil
}

func (c *client) send(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err := c.writer.Write(append(data, '\n')); err != nil {
		return err
	}
	return c.writer.Flush()
}

func (c *client) respond(id int64, result any) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return c.send(message{JSONRPC: "2.0", ID: id, Result: data})
}

func (c *client) respondError(id int64, code int, text string) error {
	return c.send(message{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: text}})
}

// call asks the host for a permitted function result. Command execution is
// synchronous, so no other host messages are expected while awaiting this
// response.
func (c *client) call(method string, params any, output any) error {
	id := c.nextID
	c.nextID++
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	if err := c.send(message{JSONRPC: "2.0", ID: id, Method: method, Params: data}); err != nil {
		return err
	}

	msg, err := c.read()
	if err != nil {
		return err
	}
	if msg.ID != id {
		return fmt.Errorf("unexpected host response id %d, want %d", msg.ID, id)
	}
	if msg.Error != nil {
		return fmt.Errorf("%s", msg.Error.Message)
	}
	if output == nil {
		return nil
	}
	return json.Unmarshal(msg.Result, output)
}
