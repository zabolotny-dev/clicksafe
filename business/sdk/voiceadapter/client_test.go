package voiceadapter

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"testing"

	voiceadapterv1 "github.com/zabolotny-dev/clicksafe/adapters/voice/gen/go/voiceadapter/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestSynthesizeStreamsMetadataReferenceAndWritesAudio(t *testing.T) {
	server := &recordingVoiceServer{}
	client, cleanup := newTestClient(t, server)
	defer cleanup()

	var output bytes.Buffer
	var progress []Progress

	result, err := client.Synthesize(
		context.Background(),
		SynthesizeRequest{
			RequestID: "request-1",
			Text:      "Hello",
			Mode:      SynthesisModeVoiceClone,
			ReferenceAudio: AudioDescriptor{
				Filename:  "ref.mp3",
				MimeType:  "audio/mpeg",
				Extension: ".mp3",
				Size:      6,
			},
			ReferenceReader: strings.NewReader("abcdef"),
			OutputConfig: OutputConfig{
				Format:    OutputFormatOGGOpus,
				ChunkSize: 3,
			},
		},
		&output,
		func(p Progress) error {
			progress = append(progress, p)
			return nil
		},
	)
	if err != nil {
		t.Fatalf("synthesize: %v", err)
	}

	if got := output.String(); got != "hello world" {
		t.Fatalf("output = %q, want hello world", got)
	}
	if result.RequestID != "request-1" {
		t.Fatalf("request id = %q", result.RequestID)
	}
	if result.Format != OutputFormatOGGOpus {
		t.Fatalf("format = %q", result.Format)
	}
	if len(progress) != 1 || progress[0].Stage != "synthesizing" {
		t.Fatalf("progress = %#v", progress)
	}
	if len(server.payloads) != 2 {
		t.Fatalf("payload count = %d, want 2", len(server.payloads))
	}
	if server.payloads[0] != "metadata" || server.payloads[1] != "reference_audio_chunk" {
		t.Fatalf("payload order = %#v", server.payloads)
	}
	if server.metadata.GetText() != "Hello" {
		t.Fatalf("metadata text = %q", server.metadata.GetText())
	}
	if got := server.reference.String(); got != "abcdef" {
		t.Fatalf("reference bytes = %q", got)
	}
}

func TestSynthesizePropagatesGRPCErrors(t *testing.T) {
	server := &recordingVoiceServer{err: status.Error(codes.InvalidArgument, "bad request")}
	client, cleanup := newTestClient(t, server)
	defer cleanup()

	var output bytes.Buffer
	_, err := client.Synthesize(
		context.Background(),
		SynthesizeRequest{
			RequestID: "request-1",
			Text:      "Hello",
			Mode:      SynthesisModeAutoVoice,
		},
		&output,
		nil,
	)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "InvalidArgument") {
		t.Fatalf("error = %v, want InvalidArgument", err)
	}
}

func TestTranscribeSendsAudio(t *testing.T) {
	server := &recordingVoiceServer{}
	client, cleanup := newTestClient(t, server)
	defer cleanup()

	text, err := client.Transcribe(
		context.Background(),
		AudioDescriptor{
			Filename:  "ref.wav",
			MimeType:  "audio/wav",
			Extension: ".wav",
			Size:      6,
		},
		strings.NewReader("abcdef"),
		"ru",
	)
	if err != nil {
		t.Fatalf("transcribe: %v", err)
	}

	if text != "recognized text" {
		t.Fatalf("text = %q, want recognized text", text)
	}
	if server.transcribe.GetAudio().GetFilename() != "ref.wav" {
		t.Fatalf("filename = %q", server.transcribe.GetAudio().GetFilename())
	}
	if got := string(server.transcribe.GetAudioData()); got != "abcdef" {
		t.Fatalf("audio = %q, want abcdef", got)
	}
	if server.transcribe.GetLanguage() != "ru" {
		t.Fatalf("language = %q, want ru", server.transcribe.GetLanguage())
	}
}

type recordingVoiceServer struct {
	voiceadapterv1.UnimplementedVoiceAdapterServiceServer
	payloads   []string
	metadata   *voiceadapterv1.SynthesisMetadata
	transcribe *voiceadapterv1.TranscribeRequest
	reference  bytes.Buffer
	err        error
}

func (s *recordingVoiceServer) Health(context.Context, *voiceadapterv1.HealthRequest) (*voiceadapterv1.HealthResponse, error) {
	return &voiceadapterv1.HealthResponse{Status: "ok"}, nil
}

func (s *recordingVoiceServer) Transcribe(_ context.Context, req *voiceadapterv1.TranscribeRequest) (*voiceadapterv1.TranscribeResponse, error) {
	s.transcribe = req
	if s.err != nil {
		return nil, s.err
	}
	return &voiceadapterv1.TranscribeResponse{Text: "recognized text"}, nil
}

func (s *recordingVoiceServer) Synthesize(stream grpc.BidiStreamingServer[voiceadapterv1.SynthesizeRequest, voiceadapterv1.SynthesizeEvent]) error {
	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		switch payload := req.GetPayload().(type) {
		case *voiceadapterv1.SynthesizeRequest_Metadata:
			s.payloads = append(s.payloads, "metadata")
			s.metadata = payload.Metadata
		case *voiceadapterv1.SynthesizeRequest_ReferenceAudioChunk:
			s.payloads = append(s.payloads, "reference_audio_chunk")
			_, _ = s.reference.Write(payload.ReferenceAudioChunk.GetData())
		}
	}

	if s.err != nil {
		return s.err
	}

	if err := stream.Send(&voiceadapterv1.SynthesizeEvent{
		Event: &voiceadapterv1.SynthesizeEvent_Progress{
			Progress: &voiceadapterv1.SynthesisProgress{
				Stage:   "synthesizing",
				Message: "running",
				Percent: 50,
			},
		},
	}); err != nil {
		return err
	}
	if err := stream.Send(&voiceadapterv1.SynthesizeEvent{
		Event: &voiceadapterv1.SynthesizeEvent_AudioChunk{
			AudioChunk: &voiceadapterv1.AudioChunk{Data: []byte("hello ")},
		},
	}); err != nil {
		return err
	}
	if err := stream.Send(&voiceadapterv1.SynthesizeEvent{
		Event: &voiceadapterv1.SynthesizeEvent_AudioChunk{
			AudioChunk: &voiceadapterv1.AudioChunk{Data: []byte("world")},
		},
	}); err != nil {
		return err
	}
	return stream.Send(&voiceadapterv1.SynthesizeEvent{
		Event: &voiceadapterv1.SynthesizeEvent_Completed{
			Completed: &voiceadapterv1.SynthesisCompleted{
				RequestId:    "request-1",
				Format:       voiceadapterv1.AudioOutputFormat_AUDIO_OUTPUT_FORMAT_OGG_OPUS,
				MimeType:     "audio/ogg; codecs=opus",
				Extension:    ".ogg",
				SampleRateHz: 48000,
				DurationMs:   1000,
				TotalBytes:   11,
			},
		},
	})
}

func newTestClient(t *testing.T, srv voiceadapterv1.VoiceAdapterServiceServer) (*Client, func()) {
	t.Helper()

	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	voiceadapterv1.RegisterVoiceAdapterServiceServer(grpcServer, srv)

	go func() {
		_ = grpcServer.Serve(listener)
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("new grpc client: %v", err)
	}

	cleanup := func() {
		_ = conn.Close()
		grpcServer.Stop()
		_ = listener.Close()
	}

	return &Client{
		conn: conn,
		api:  voiceadapterv1.NewVoiceAdapterServiceClient(conn),
	}, cleanup
}
