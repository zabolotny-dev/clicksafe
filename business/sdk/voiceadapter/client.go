package voiceadapter

import (
	"context"
	"errors"
	"fmt"
	"io"

	voiceadapterv1 "github.com/zabolotny-dev/clicksafe/adapters/voice/gen/go/voiceadapter/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn  *grpc.ClientConn
	api   voiceadapterv1.VoiceAdapterServiceClient
	token string
}

func New(cfg Config) (*Client, error) {
	if cfg.Addr == "" {
		return nil, errors.New("voice adapter addr is required")
	}

	conn, err := grpc.NewClient(cfg.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("new voice adapter client: %w", err)
	}

	return &Client{
		conn:  conn,
		api:   voiceadapterv1.NewVoiceAdapterServiceClient(conn),
		token: cfg.Token,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Health(ctx context.Context) (*voiceadapterv1.HealthResponse, error) {
	resp, err := c.api.Health(c.withToken(ctx), &voiceadapterv1.HealthRequest{})
	if err != nil {
		return nil, mapErr(err)
	}
	return resp, nil
}

func (c *Client) Transcribe(ctx context.Context, desc AudioDescriptor, reader io.Reader, language string) (string, error) {
	if reader == nil {
		return "", errors.New("reader is required")
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("read audio: %w", err)
	}
	if len(data) == 0 {
		return "", errors.New("audio data is required")
	}

	resp, err := c.api.Transcribe(c.withToken(ctx), &voiceadapterv1.TranscribeRequest{
		Audio:     audioDescriptorProto(desc),
		AudioData: data,
		Language:  language,
	})
	if err != nil {
		return "", mapErr(err)
	}

	return resp.GetText(), nil
}

func (c *Client) Synthesize(ctx context.Context, req SynthesizeRequest, writer io.Writer, onProgress ProgressFunc) (SynthesizeResult, error) {
	if req.Text == "" {
		return SynthesizeResult{}, errors.New("text is required")
	}
	if req.Mode == "" {
		return SynthesizeResult{}, errors.New("mode is required")
	}
	if writer == nil {
		return SynthesizeResult{}, errors.New("writer is required")
	}
	if req.Mode == SynthesisModeVoiceClone && req.ReferenceReader == nil {
		return SynthesizeResult{}, errors.New("reference reader is required for voice clone")
	}
	if req.Mode == SynthesisModeVoiceDesign && req.Instruct == "" {
		return SynthesizeResult{}, errors.New("instruct is required for voice design")
	}

	options, err := modelOptionsProto(req.ModelOptions)
	if err != nil {
		return SynthesizeResult{}, fmt.Errorf("model options: %w", err)
	}

	stream, err := c.api.Synthesize(c.withToken(ctx))
	if err != nil {
		return SynthesizeResult{}, mapErr(err)
	}

	err = stream.Send(&voiceadapterv1.SynthesizeRequest{
		Payload: &voiceadapterv1.SynthesizeRequest_Metadata{
			Metadata: &voiceadapterv1.SynthesisMetadata{
				RequestId:        req.RequestID,
				Text:             req.Text,
				Mode:             synthesisModeProto(req.Mode),
				Language:         req.Language,
				ReferenceText:    req.ReferenceText,
				Instruct:         req.Instruct,
				ReferenceAudio:   audioDescriptorProto(req.ReferenceAudio),
				GenerationConfig: generationConfigProto(req.GenerationConfig),
				OutputConfig:     outputConfigProto(req.OutputConfig),
				ModelOptions:     options,
			},
		},
	})
	if err != nil {
		return SynthesizeResult{}, mapErr(err)
	}

	if req.ReferenceReader != nil {
		if err := c.sendReference(ctx, stream, req.ReferenceReader); err != nil {
			return SynthesizeResult{}, err
		}
	}

	if err := stream.CloseSend(); err != nil {
		return SynthesizeResult{}, mapErr(err)
	}

	var result SynthesizeResult
	completed := false
	for {
		event, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return SynthesizeResult{}, mapErr(err)
		}

		switch payload := event.GetEvent().(type) {
		case *voiceadapterv1.SynthesizeEvent_Progress:
			if onProgress != nil {
				progress := payload.Progress
				if err := onProgress(Progress{
					Stage:   progress.GetStage(),
					Message: progress.GetMessage(),
					Percent: progress.GetPercent(),
				}); err != nil {
					return SynthesizeResult{}, err
				}
			}
		case *voiceadapterv1.SynthesizeEvent_AudioChunk:
			if _, err := writer.Write(payload.AudioChunk.GetData()); err != nil {
				return SynthesizeResult{}, fmt.Errorf("write audio chunk: %w", err)
			}
		case *voiceadapterv1.SynthesizeEvent_Completed:
			response := payload.Completed
			result = SynthesizeResult{
				RequestID:     response.GetRequestId(),
				Format:        outputFormat(response.GetFormat()),
				MimeType:      response.GetMimeType(),
				Extension:     response.GetExtension(),
				SampleRateHz:  response.GetSampleRateHz(),
				DurationMs:    response.GetDurationMs(),
				TotalBytes:    response.GetTotalBytes(),
				WaveformPeaks: append([]byte(nil), response.GetWaveformPeaks()...),
			}
			completed = true
		}
	}

	if !completed {
		return SynthesizeResult{}, errors.New("voice adapter completed without result")
	}
	return result, nil
}

func (c *Client) sendReference(ctx context.Context, stream grpc.BidiStreamingClient[voiceadapterv1.SynthesizeRequest, voiceadapterv1.SynthesizeEvent], reader io.Reader) error {
	buf := make([]byte, 64*1024)
	var offset int64
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		n, readErr := reader.Read(buf)
		if n > 0 {
			chunk := append([]byte(nil), buf[:n]...)
			err := stream.Send(&voiceadapterv1.SynthesizeRequest{
				Payload: &voiceadapterv1.SynthesizeRequest_ReferenceAudioChunk{
					ReferenceAudioChunk: &voiceadapterv1.ReferenceAudioChunk{
						Data:   chunk,
						Offset: offset,
					},
				},
			})
			if err != nil {
				return mapErr(err)
			}
			offset += int64(n)
		}
		if errors.Is(readErr, io.EOF) {
			return nil
		}
		if readErr != nil {
			return fmt.Errorf("read reference audio: %w", readErr)
		}
	}
}

func (c *Client) withToken(ctx context.Context) context.Context {
	if c.token == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+c.token)
}
