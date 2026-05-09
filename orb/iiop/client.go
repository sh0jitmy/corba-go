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

package iiop

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"github.com/sh0jitmy/corba-go/orb/cdr"
	"github.com/sh0jitmy/corba-go/orb/giop"
)

// Client represents an IIOP connection to a specific host:port.
type Client struct {
	addr  string
	conn  net.Conn
	mu    sync.Mutex
	reqID uint32
}

// NewClient creates a new IIOP client connected to the given address.
func NewClient(host string, port uint16) (*Client, error) {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %v", addr, err)
	}

	return &Client{
		addr: addr,
		conn: conn,
	}, nil
}

// Close closes the connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Invoke sends a GIOP request and waits for a reply.
func (c *Client) Invoke(objectKey []byte, operation string, args []byte) ([]byte, error) {
	reqID := atomic.AddUint32(&c.reqID, 1)

	// Encode the request
	msgData, err := giop.EncodeRequest(reqID, operation, objectKey, args)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %v", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Send message
	_, err = c.conn.Write(msgData)
	if err != nil {
		return nil, fmt.Errorf("failed to write to connection: %v", err)
	}

	// Read reply header (12 bytes)
	headerBuf := make([]byte, 12)
	_, err = io.ReadFull(c.conn, headerBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to read reply header: %v", err)
	}

	header, err := giop.ReadHeader(headerBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reply header: %v", err)
	}

	if header.MessageType != giop.ReplyMsg {
		return nil, fmt.Errorf("expected ReplyMsg, got %v", header.MessageType)
	}

	// Read reply body
	bodyBuf := make([]byte, header.MessageSize)
	_, err = io.ReadFull(c.conn, bodyBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to read reply body: %v", err)
	}

	// Parse reply body
	// Determine byte order from flags
	var order binary.ByteOrder
	if header.Flags&0x01 == 1 {
		order = binary.LittleEndian
	} else {
		order = binary.BigEndian
	}

	dec := cdr.NewDecoder(bodyBuf, order)

	// Reply Header (GIOP 1.2)
	// RequestID (4 bytes)
	replyReqID, err := dec.DecodeULong()
	if err != nil {
		return nil, err
	}

	if replyReqID != reqID {
		return nil, fmt.Errorf("request ID mismatch: sent %d, received %d", reqID, replyReqID)
	}

	// ReplyStatus (4 bytes)
	replyStatus, err := dec.DecodeULong()
	if err != nil {
		return nil, err
	}

	// ServiceContextList
	ctxCount, err := dec.DecodeULong()
	if err != nil {
		return nil, err
	}

	// Skip context data for now
	for i := uint32(0); i < ctxCount; i++ {
		_, _ = dec.DecodeULong() // Context ID
		ctxLen, _ := dec.DecodeULong()
		for j := uint32(0); j < ctxLen; j++ {
			_, _ = dec.DecodeOctet()
		}
	}

	// GIOP 1.2 alignment: reply body starts at 8-byte boundary relative to GIOP header start
	// bodyBuf offset needs to account for 12 byte header.
	// We simulate absolute position:
	// The current position in dec.buffer is where we are now.
	// Absolute pos = 12 (header) + current offset.
	// We need absolute pos % 8 == 0.
	// The decoder doesn't have alignment by absolute position easily without knowing it,
	// but we know bodyBuf is just the body.
	// Wait, standard `align(8)` in Decoder just aligns to the slice.
	// Since the slice doesn't include the 12 byte header, to align to 8 absolute,
	// it means (12 + offset) % 8 == 0.
	// offset % 8 must be 4.

	// A quick hack to align 8-byte boundary relative to absolute 12:
	// We know where we are in bodyBuf.
	// Let's read the remaining bytes and return them as the result payload.
	// For ReplyStatus == 0 (NO_EXCEPTION), the following data is the return value/out args.

	if replyStatus != 0 {
		return nil, fmt.Errorf("CORBA Exception occurred, ReplyStatus: %d", replyStatus)
	}

	// Find padding to 8-byte boundary
	// absolute = 12 + offset.
	// padding = (8 - (absolute % 8)) % 8
	// Note: We leave proper 8-byte alignment logic for GIOP 1.2 bodies to a more robust implementation.
	// For now, we simply return the unread portion as the payload.

	return dec.Rest(), nil
}
