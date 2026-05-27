package voiceadapter

import (
	"fmt"
	"io"

	voiceadapterv1 "github.com/zabolotny-dev/clicksafe/adapters/voice/gen/go/voiceadapter/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type Config struct {
	Addr  string
	Token string
}

type SynthesisMode string

const (
	SynthesisModeVoiceClone  SynthesisMode = "VOICE_CLONE"
	SynthesisModeVoiceDesign SynthesisMode = "VOICE_DESIGN"
	SynthesisModeAutoVoice   SynthesisMode = "AUTO_VOICE"
)

type OutputFormat string

const (
	OutputFormatWAVPCM16 OutputFormat = "WAV_PCM16"
	OutputFormatOGGOpus  OutputFormat = "OGG_OPUS"
)

type AudioDescriptor struct {
	Filename  string
	MimeType  string
	Extension string
	Size      int64
}

type GenerationConfig struct {
	Speed               *float64
	DurationSeconds     *float64
	InferenceSteps      *int32
	GuidanceScale       *float64
	Denoise             *bool
	PreprocessPrompt    *bool
	PostprocessOutput   *bool
	AudioChunkDuration  *float64
	AudioChunkThreshold *float64
}

type OutputConfig struct {
	Format       OutputFormat
	SampleRateHz int32
	ChunkSize    int32
}

type SynthesizeRequest struct {
	RequestID        string
	Text             string
	Mode             SynthesisMode
	Language         string
	ReferenceText    string
	Instruct         string
	ReferenceAudio   AudioDescriptor
	ReferenceReader  io.Reader
	GenerationConfig GenerationConfig
	OutputConfig     OutputConfig
	ModelOptions     map[string]any
}

type Progress struct {
	Stage   string
	Message string
	Percent int32
}

type SynthesizeResult struct {
	RequestID     string
	Format        OutputFormat
	MimeType      string
	Extension     string
	SampleRateHz  int32
	DurationMs    int64
	TotalBytes    int64
	WaveformPeaks []byte
}

type ProgressFunc func(Progress) error

func synthesisModeProto(mode SynthesisMode) voiceadapterv1.SynthesisMode {
	switch mode {
	case SynthesisModeVoiceClone:
		return voiceadapterv1.SynthesisMode_SYNTHESIS_MODE_VOICE_CLONE
	case SynthesisModeVoiceDesign:
		return voiceadapterv1.SynthesisMode_SYNTHESIS_MODE_VOICE_DESIGN
	case SynthesisModeAutoVoice:
		return voiceadapterv1.SynthesisMode_SYNTHESIS_MODE_AUTO_VOICE
	default:
		return voiceadapterv1.SynthesisMode_SYNTHESIS_MODE_UNSPECIFIED
	}
}

func outputFormatProto(format OutputFormat) voiceadapterv1.AudioOutputFormat {
	switch format {
	case OutputFormatOGGOpus:
		return voiceadapterv1.AudioOutputFormat_AUDIO_OUTPUT_FORMAT_OGG_OPUS
	case OutputFormatWAVPCM16, "":
		return voiceadapterv1.AudioOutputFormat_AUDIO_OUTPUT_FORMAT_WAV_PCM16
	default:
		return voiceadapterv1.AudioOutputFormat_AUDIO_OUTPUT_FORMAT_UNSPECIFIED
	}
}

func outputFormat(format voiceadapterv1.AudioOutputFormat) OutputFormat {
	switch format {
	case voiceadapterv1.AudioOutputFormat_AUDIO_OUTPUT_FORMAT_OGG_OPUS:
		return OutputFormatOGGOpus
	case voiceadapterv1.AudioOutputFormat_AUDIO_OUTPUT_FORMAT_WAV_PCM16:
		return OutputFormatWAVPCM16
	default:
		return ""
	}
}

func generationConfigProto(cfg GenerationConfig) *voiceadapterv1.GenerationConfig {
	return &voiceadapterv1.GenerationConfig{
		Speed:               cfg.Speed,
		DurationSeconds:     cfg.DurationSeconds,
		InferenceSteps:      cfg.InferenceSteps,
		GuidanceScale:       cfg.GuidanceScale,
		Denoise:             cfg.Denoise,
		PreprocessPrompt:    cfg.PreprocessPrompt,
		PostprocessOutput:   cfg.PostprocessOutput,
		AudioChunkDuration:  cfg.AudioChunkDuration,
		AudioChunkThreshold: cfg.AudioChunkThreshold,
	}
}

func outputConfigProto(cfg OutputConfig) *voiceadapterv1.OutputConfig {
	return &voiceadapterv1.OutputConfig{
		Format:       outputFormatProto(cfg.Format),
		SampleRateHz: cfg.SampleRateHz,
		ChunkSize:    cfg.ChunkSize,
	}
}

func audioDescriptorProto(desc AudioDescriptor) *voiceadapterv1.AudioDescriptor {
	return &voiceadapterv1.AudioDescriptor{
		Filename:  desc.Filename,
		MimeType:  desc.MimeType,
		Extension: desc.Extension,
		Size:      desc.Size,
	}
}

func modelOptionsProto(options map[string]any) (*structpb.Struct, error) {
	if len(options) == 0 {
		return &structpb.Struct{}, nil
	}
	return structpb.NewStruct(options)
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("voice adapter failed: %w", err)
	}
	if st.Code() == codes.Canceled || st.Code() == codes.DeadlineExceeded {
		return err
	}
	return fmt.Errorf("voice adapter %s: %s", st.Code(), st.Message())
}
