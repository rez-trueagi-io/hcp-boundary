// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loopback

import (
	"context"
	"fmt"
	"io"
	"sync"

	plgpb "github.com/hashicorp/boundary/sdk/pbs/plugin"
	"google.golang.org/grpc/metadata"
)

// getObjectStreamResponse is used to mock a message sent from the server to the client.
type getObjectStreamResponse struct {
	msg *plgpb.GetObjectResponse
	err error
}

// getObjectStream is used to mock the interactions between
// the client and server for the GetObject method.
type getObjectStream struct {
	client plgpb.StoragePluginService_GetObjectClient
	server plgpb.StoragePluginService_GetObjectServer

	// messages is used to mock the server sending messages to the client.
	messages chan *getObjectStreamResponse

	m            *sync.Mutex
	streamClosed bool
}

// IsStreamClosed returns true if the stream is closed.
func (s *getObjectStream) IsStreamClosed() bool {
	s.m.Lock()
	defer s.m.Unlock()
	return s.streamClosed
}

// Close closes the channels of the stream and sets the streamClosed flag to true.
// A closeStream is used to prevent the channels from being closed multiple times.
func (s *getObjectStream) Close() {
	s.m.Lock()
	defer s.m.Unlock()
	if s.streamClosed {
		return
	}
	close(s.messages)
	s.streamClosed = true
}

// getObjectClient is used to mock the client stream
// interactions for the GetObject method.
type getObjectClient struct {
	// sentFromServer is used to mock the server sending messages to the client.
	sentFromServer chan *getObjectStreamResponse

	// closeStream is used to close the channels of the stream.
	// This is used to prevent the channels from being closed multiple times.
	// This is needed because the channel can be closed by the client or the server.
	closeStream func()

	// isStreamClosed is used to check if the stream is closed.
	// This is needed because the channel can be closed by the client or the server.
	isStreamClosed func() bool
}

// Recv will block until a message is received from the server.
// Recv will return io.EOF if the server closes the stream.
// Recv will return an error if the server sends an error.
func (c *getObjectClient) Recv() (*plgpb.GetObjectResponse, error) {
	resp, ok := <-c.sentFromServer
	if !ok {
		return nil, io.EOF
	}
	return resp.msg, resp.err
}

// Header should not be used.
// Header is implemeted to satisfy the grpc.ClientStream interface.
// Header will always return an empty metadata and an nil error.
func (c *getObjectClient) Header() (metadata.MD, error) {
	return make(metadata.MD), nil
}

// Trailer should not be used.
// Trailer is implemeted to satisfy the grpc.ClientStream interface.
// Trailer will always return an empty metadata.
func (c *getObjectClient) Trailer() metadata.MD {
	return make(metadata.MD)
}

// CloseSend will close the channel used to retrieve messages from
// the server. This will cause the Recv method to return io.EOF.
// CloseSend will return an error if the channel is already closed.
func (c *getObjectClient) CloseSend() error {
	if c.isStreamClosed() {
		return fmt.Errorf("stream is closed")
	}
	c.closeStream()
	return nil
}

// Context will always return a Background context.
func (c *getObjectClient) Context() context.Context {
	return context.Background()
}

// SendMsg should not be used.
// SendMsg is implemeted to satisfy the grpc.ClientStream interface.
// SendMsg will always return an nil error.
func (c *getObjectClient) SendMsg(m interface{}) error {
	return nil
}

// RecvMsg should not be used.
// RecvMsg is implemeted to satisfy the grpc.ClientStream interface.
// RecvMsg will always return an nil error.
func (c *getObjectClient) RecvMsg(m interface{}) error {
	return nil
}

// getObjectServer is used to mock the server stream
// interactions for the GetObject method.
type getObjectServer struct {
	// sendToClient is used to mock the server sending messages to the client.
	sendToClient chan *getObjectStreamResponse

	// closeStream is used to close the channels of the stream.
	// This is used to prevent the channels from being closed multiple times.
	// This is needed because the channel can be closed by the client or the server.
	closeStream func()

	// isStreamClosed is used to check if the stream is closed.
	// This is needed because the channel can be closed by the client or the server.
	isStreamClosed func() bool
}

// Send will send a message to the client.
// Send will return an error if the client closes the stream.
// Send will return an error if the response is nil.
func (s *getObjectServer) Send(resp *plgpb.GetObjectResponse) error {
	if resp == nil {
		return fmt.Errorf(`parameter arg "resp GetObjectResponse" cannot be nil`)
	}
	if s.isStreamClosed() {
		return fmt.Errorf("stream is closed")
	}
	s.sendToClient <- &getObjectStreamResponse{msg: resp}
	return nil
}

// SetHeader should not be used.
// SetHeader is implemeted to satisfy the grpc.ServerStream interface.
// SetHeader will always return an nil error.
func (s *getObjectServer) SetHeader(metadata.MD) error {
	return nil
}

// SendHeader should not be used.
// SendHeader is implemeted to satisfy the grpc.ServerStream interface.
// SendHeader will always return an nil error.
func (s *getObjectServer) SendHeader(metadata.MD) error {
	return nil
}

// SetTrailer should not be used.
// SetTrailer is implemeted to satisfy the grpc.ServerStream interface.
// SetTrailer will always return an nil error.
func (s *getObjectServer) SetTrailer(metadata.MD) {
}

// Context will always return a Background context.
func (s *getObjectServer) Context() context.Context {
	return context.Background()
}

// SendMsg allows sending GetObjectResponse messages to the client.
// SendMsg allows sending errors other than io.EOF to the client.
// Sending an error message will close the stream.
// SendMsg returns an invalid argument error if the message is not
// an error or GetObjectResponse.
// SendMsg will return an error if the stream is closed.
func (s *getObjectServer) SendMsg(m interface{}) error {
	switch msg := m.(type) {
	case *plgpb.GetObjectResponse:
		if s.isStreamClosed() {
			return fmt.Errorf("stream is closed")
		}
		s.sendToClient <- &getObjectStreamResponse{msg: msg}
	case error:
		if s.isStreamClosed() {
			return fmt.Errorf("stream is closed")
		}
		defer s.closeStream()
		s.sendToClient <- &getObjectStreamResponse{err: msg}
	default:
		return fmt.Errorf("invalid argument %v", m)
	}
	return nil
}

// RecvMsg should not be used.
// RecvMsg is implemeted to satisfy the grpc.ServerStream interface.
// RecvMsg will always return an nil error.
func (s *getObjectServer) RecvMsg(m interface{}) error {
	return nil
}

// newGetObjectStream will create a mock stream for the GetObject method.
// The client and server stream is mocked by creating a GetObjectResponse
// channel and an error channel that is shared between the client and server.
func newGetObjectStream() *getObjectStream {
	stream := &getObjectStream{
		m:        new(sync.Mutex),
		messages: make(chan *getObjectStreamResponse),
	}
	stream.client = &getObjectClient{
		sentFromServer: stream.messages,
		closeStream:    stream.Close,
		isStreamClosed: stream.IsStreamClosed,
	}
	stream.server = &getObjectServer{
		sendToClient:   stream.messages,
		closeStream:    stream.Close,
		isStreamClosed: stream.IsStreamClosed,
	}
	return stream
}

// putObjectStreamRequest is used to mock a message sent from the client to the server.
type putObjectStreamRequest struct {
	msg *plgpb.PutObjectRequest
	err error
}

// puObjectStreamResponse is used to mock a message sent from the server to the client.
type putObjectStreamResponse struct {
	msg *plgpb.PutObjectResponse
	err error
}

// putObjectStream is used to mock the interactions between
// the client and server for the PutObject method.
type putObjectStream struct {
	client plgpb.StoragePluginService_PutObjectClient
	server plgpb.StoragePluginService_PutObjectServer

	// requests is used to mock the client sending messages to the server.
	requests chan *putObjectStreamRequest

	// responses is used to mock the server sending messages to the client.
	responses chan *putObjectStreamResponse

	m            *sync.Mutex
	clientClosed bool
	serverClosed bool
}

func (s *putObjectStream) CloseClient() {
	s.m.Lock()
	defer s.m.Unlock()
	if s.clientClosed {
		return
	}
	close(s.requests)
	s.clientClosed = true
}

func (s *putObjectStream) IsClientClosed() bool {
	s.m.Lock()
	defer s.m.Unlock()
	return s.clientClosed
}

func (s *putObjectStream) CloseServer() {
	s.m.Lock()
	defer s.m.Unlock()
	if s.serverClosed {
		return
	}
	close(s.responses)
	s.serverClosed = true
}

func (s *putObjectStream) IsServerClosed() bool {
	s.m.Lock()
	defer s.m.Unlock()
	return s.serverClosed
}

// putObjectClient is used to mock the client stream
// interactions for the PutObject method.
type putObjectClient struct {
	sendToServer   chan *putObjectStreamRequest
	sentFromServer chan *putObjectStreamResponse

	// closeStream is used to close the channels of the stream.
	// This is used to prevent the channels from being closed multiple times.
	// This is needed because the channel can be closed by the client or the server.
	closeStream func()

	// isStreamClosed is used to check if the stream is closed.
	// This is needed because the channel can be closed by the client or the server.
	isStreamClosed func() bool
}

// Send will send a message to the server.
// Send will return an error if the request is nil.
// Send will return an error if the stream is closed.
func (c *putObjectClient) Send(req *plgpb.PutObjectRequest) error {
	if req == nil {
		return fmt.Errorf(`parameter arg "req PutObjectRequest" cannot be nil`)
	}
	if c.isStreamClosed() {
		return fmt.Errorf("stream is closed")
	}
	c.sendToServer <- &putObjectStreamRequest{msg: req}
	return nil
}

// CloseAndRecv will return an PutObjectResponse if the server returns one.
// CloseAndRecv will return an error if the server returns an error.
// CloseAndRecv will return an io.EOF error if the server closes the stream.
// CloseAndRecv will close the stream used to send messages to the server.
func (c *putObjectClient) CloseAndRecv() (*plgpb.PutObjectResponse, error) {
	c.closeStream()
	resp, ok := <-c.sentFromServer
	if !ok {
		return nil, io.EOF
	}
	return resp.msg, resp.err
}

// Header should not be used.
// Header is implemeted to satisfy the grpc.ClientStream interface.
// Header will always return an empty metadata and an nil error.
func (c *putObjectClient) Header() (metadata.MD, error) {
	return make(metadata.MD), nil
}

// Trailer should not be used.
// Trailer is implemeted to satisfy the grpc.ClientStream interface.
// Trailer will always return an empty metadata.
func (c *putObjectClient) Trailer() metadata.MD {
	return make(metadata.MD)
}

// CloseSend will close the channel used to retrieve messages from
// the server. This will cause the CloseAndRecv method to return io.EOF.
// CloseSend will always return a nill error.
func (c *putObjectClient) CloseSend() error {
	if c.isStreamClosed() {
		return fmt.Errorf("stream is closed")
	}
	c.closeStream()
	return nil
}

// Context will always return a Background context.
func (c *putObjectClient) Context() context.Context {
	return context.Background()
}

// SendMsg should not be used.
// SendMsg is implemeted to satisfy the grpc.ClientStream interface.
// SendMsg will always return an nil error.
func (c *putObjectClient) SendMsg(m interface{}) error {
	return nil
}

// RecvMsg should not be used.
// RecvMsg is implemeted to satisfy the grpc.ClientStream interface.
// RecvMsg will always return an nil error.
func (c *putObjectClient) RecvMsg(m interface{}) error {
	return nil
}

// putObjectServer is used to mock the server stream
// interactions for the PutObject method.
type putObjectServer struct {
	sentFromClient chan *putObjectStreamRequest
	sentToClient   chan *putObjectStreamResponse

	// closeStream is used to close the channels of the stream.
	// This is used to prevent the channels from being closed multiple times.
	// This is needed because the channel can be closed by the client or the server.
	closeStream func()

	// isStreamClosed is used to check if the stream is closed.
	// This is needed because the channel can be closed by the client or the server.
	isStreamClosed func() bool
}

// SendAndClose will send a PutObjectResponse to the client.
// SendAndClose will return an error if the response is nil.
// SendAndClose will close the stream used to send messages from the client.
// SendAndClose will return an error if the stream is closed.
func (s *putObjectServer) SendAndClose(resp *plgpb.PutObjectResponse) error {
	if resp == nil {
		return fmt.Errorf(`parameter arg "resp PutObjectResponse" cannot be nil`)
	}
	if s.isStreamClosed() {
		return fmt.Errorf("stream is closed")
	}
	s.sentToClient <- &putObjectStreamResponse{msg: resp}
	s.closeStream()
	return nil
}

// Recv will read a PutObjectRequest from the stream.
// Recv will return an io.EOF error if the stream is closed.
func (s *putObjectServer) Recv() (*plgpb.PutObjectRequest, error) {
	req, ok := <-s.sentFromClient
	if !ok {
		return nil, io.EOF
	}
	return req.msg, req.err
}

// SetHeader should not be used.
// SetHeader is implemeted to satisfy the grpc.ServerStream interface.
// SetHeader will always return an nil error.
func (s *putObjectServer) SetHeader(metadata.MD) error {
	return nil
}

// SendHeader should not be used.
// SendHeader is implemeted to satisfy the grpc.ServerStream interface.
// SendHeader will always return an nil error.
func (s *putObjectServer) SendHeader(metadata.MD) error {
	return nil
}

// SetTrailer should not be used.
// SetTrailer is implemeted to satisfy the grpc.ServerStream interface.
func (s *putObjectServer) SetTrailer(metadata.MD) {
}

// Context will always return a Background context.
func (s *putObjectServer) Context() context.Context {
	return context.Background()
}

// SendMsg allows sending errors other than io.EOF to the client.
// Sending an error message will close the stream.
// SendMsg allows sending PutObjectResponse messages to the client.
// SendMsg returns an invalid argument error if the message is not
// an error or PutObjectResponse.
// SendMsg will return an error if the stream is closed.
func (s *putObjectServer) SendMsg(m interface{}) error {
	switch msg := m.(type) {
	case *plgpb.PutObjectResponse:
		if s.isStreamClosed() {
			return fmt.Errorf("stream is closed")
		}
		s.sentToClient <- &putObjectStreamResponse{msg: msg}
	case error:
		if s.isStreamClosed() {
			return fmt.Errorf("stream is closed")
		}
		s.sentToClient <- &putObjectStreamResponse{err: msg}
		defer s.closeStream()
	default:
		return fmt.Errorf("invalid argument %v", m)
	}
	return nil
}

// RecvMsg should not be used.
// RecvMsg is implemeted to satisfy the grpc.ServerStream interface.
// RecvMsg will always return an nil error.
func (s *putObjectServer) RecvMsg(m interface{}) error {
	return nil
}

// newPutObjectStream will create a mock stream for the PutObject method.
// The client and server stream is mocked by creating a PutObjectRequest,
// PutObjectResponse and an error channel that is shared between the client
// and server.
func newPutObjectStream() *putObjectStream {
	stream := &putObjectStream{
		m:         new(sync.Mutex),
		requests:  make(chan *putObjectStreamRequest),
		responses: make(chan *putObjectStreamResponse),
	}
	stream.client = &putObjectClient{
		sendToServer:   stream.requests,
		sentFromServer: stream.responses,
		closeStream:    stream.CloseClient,
		isStreamClosed: stream.IsClientClosed,
	}
	stream.server = &putObjectServer{
		sentFromClient: stream.requests,
		sentToClient:   stream.responses,
		closeStream:    stream.CloseServer,
		isStreamClosed: stream.IsServerClosed,
	}
	return stream
}