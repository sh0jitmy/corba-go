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
	"log"
	"net"

	"github.com/shjtmy/corba-go/orb/cdr"
	"github.com/shjtmy/corba-go/orb/giop"
)

// Server handles incoming IIOP connections.
type Server struct {
	listener net.Listener
}

// Handler is a function that processes an IIOP request and returns a reply payload.
type Handler func(objectKey []byte, operation string, reqPayload []byte) (replyPayload []byte, err error)

// NewServer creates and starts a new IIOP server.
func NewServer(port uint16) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Server{listener: l}, nil
}

// Serve starts accepting connections and processing requests using the given handler.
func (s *Server) Serve(handler Handler) error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}
		go s.handleConnection(conn, handler)
	}
}

// Close stops the server.
func (s *Server) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// Addr returns the address the server is listening on.
func (s *Server) Addr() net.Addr {
	return s.listener.Addr()
}

func (s *Server) handleConnection(conn net.Conn, handler Handler) {
	defer func() { _ = conn.Close() }()

	for {
		// Read GIOP Header
		headerBuf := make([]byte, 12)
		_, err := io.ReadFull(conn, headerBuf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading GIOP header: %v", err)
			}
			return
		}

		header, err := giop.ReadHeader(headerBuf)
		if err != nil {
			log.Printf("Error parsing GIOP header: %v", err)
			return
		}

		if header.MessageType == giop.LocateRequestMsg {
			// Handle LocateRequest
			bodyBuf := make([]byte, header.MessageSize)
			if _, err := io.ReadFull(conn, bodyBuf); err != nil {
				log.Printf("Error reading LocateRequest body: %v", err)
				return
			}

			var order binary.ByteOrder
			if header.Flags&0x01 == 1 {
				order = binary.LittleEndian
			} else {
				order = binary.BigEndian
			}
			dec := cdr.NewDecoder(bodyBuf, order)
			reqID, _ := dec.DecodeULong()

			// Send LocateReply with OBJECT_HERE (1)
			enc := cdr.NewEncoder(binary.LittleEndian)
			enc.EncodeULong(reqID)
			enc.EncodeULong(1) // OBJECT_HERE

			hEnc := cdr.NewEncoder(binary.LittleEndian)
			giop.WriteHeader(hEnc, &giop.Header{
				Magic:       [4]byte{'G', 'I', 'O', 'P'},
				Version:     header.Version,
				Flags:       1,
				MessageType: giop.LocateReplyMsg,
				MessageSize: uint32(len(enc.Bytes())), //#nosec G115 -- encoder output is bounded
			})

			if _, err := conn.Write(append(hEnc.Bytes(), enc.Bytes()...)); err != nil {
				log.Printf("Error writing LocateReply: %v", err)
				return
			}
			continue // keep connection open
		} else if header.MessageType != giop.RequestMsg {
			log.Printf("Unsupported message type: %v", header.MessageType)
			return
		}

		// Read Request Body
		bodyBuf := make([]byte, header.MessageSize)
		_, err = io.ReadFull(conn, bodyBuf)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			return
		}

		// Parse Request (GIOP 1.2)
		var order binary.ByteOrder
		if header.Flags&0x01 == 1 {
			order = binary.LittleEndian
		} else {
			order = binary.BigEndian
		}

		dec := cdr.NewDecoder(bodyBuf, order)
		reqID, err := dec.DecodeULong()
		if err != nil {
			log.Printf("Error decoding request ID: %v", err)
			return
		}
		_, _ = dec.DecodeOctet() // Response flags
		_, _ = dec.DecodeOctet() // Reserved
		_, _ = dec.DecodeOctet() // Reserved
		_, _ = dec.DecodeOctet() // Reserved
		_, _ = dec.DecodeShort() // TargetAddress Disposition (0 = KeyAddr)

		objKeyLen, err := dec.DecodeULong()
		if err != nil {
			log.Printf("Error decoding ObjectKey length: %v", err)
			return
		}

		objKey := make([]byte, objKeyLen)
		for i := uint32(0); i < objKeyLen; i++ {
			b, _ := dec.DecodeOctet()
			objKey[i] = b
		}

		operation, err := dec.DecodeString()
		if err != nil {
			log.Printf("Error decoding operation: %v", err)
			return
		}

		ctxCount, _ := dec.DecodeULong()
		for i := uint32(0); i < ctxCount; i++ {
			_, _ = dec.DecodeULong() // Context ID
			ctxLen, _ := dec.DecodeULong()
			for j := uint32(0); j < ctxLen; j++ {
				_, _ = dec.DecodeOctet()
			}
		}

		// Read padding for 8-byte alignment (simplified logic)
		// For proper GIOP 1.2, alignment is absolute.
		// Actually, `handler` might need to unmarshal the payload itself.
		reqPayload := dec.Rest()

		// Invoke the user handler
		replyPayload, handlerErr := handler(objKey, operation, reqPayload)

		// Create Reply Message
		enc := cdr.NewEncoder(binary.LittleEndian) // Encode reply in Little Endian

		// Encode Reply ID (must match Request ID)
		enc.EncodeULong(reqID)

		if handlerErr != nil {
			// ReplyStatus: SYSTEM_EXCEPTION = 2 (simplified)
			enc.EncodeULong(2)
			enc.EncodeULong(0) // 0 Service Contexts
			// Encode exception info... skipping for now, just encode a generic string
			enc.EncodeString(handlerErr.Error())
		} else {
			// ReplyStatus: NO_EXCEPTION = 0
			enc.EncodeULong(0)
			enc.EncodeULong(0) // 0 Service Contexts

			// Append reply payload (need alignment in proper implementation)
			currentLen := len(enc.Bytes())
			absPos := currentLen + 12
			padding := (8 - (absPos % 8)) % 8
			for i := 0; i < padding; i++ {
				enc.EncodeOctet(0)
			}
			// Let's use internal buffer for payload
			b := enc.Bytes()
			b = append(b, replyPayload...)

			// Build Header
			hEnc := cdr.NewEncoder(binary.LittleEndian)
			giop.WriteHeader(hEnc, &giop.Header{
				Magic:       [4]byte{'G', 'I', 'O', 'P'},
				Version:     giop.Version{Major: 1, Minor: 2},
				Flags:       1,
				MessageType: giop.ReplyMsg,
				MessageSize: uint32(len(b)), //#nosec G115 -- reply payload is bounded
			})

			replyMsg := append(hEnc.Bytes(), b...)

			// Send reply
			if _, err := conn.Write(replyMsg); err != nil {
				log.Printf("Error writing reply: %v", err)
				return
			}
		}
	}
}
