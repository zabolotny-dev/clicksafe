package voiceapp

import (
	"bytes"
	"context"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/sdk/voiceadapter"
)

type app struct {
	voice           *voiceadapter.Client
	cloneTimeout    time.Duration
	asrTimeout      time.Duration
	healthTTL       time.Duration
	outputChunkSize int32
}

func newApp(voice *voiceadapter.Client, cfg RuntimeConfig) *app {
	return &app{
		voice:           voice,
		cloneTimeout:    cfg.CloneTimeout,
		asrTimeout:      cfg.ASRTimeout,
		healthTTL:       cfg.HealthTTL,
		outputChunkSize: cfg.OutputChunkSize,
	}
}

func (a *app) status(c *echo.Context) error {
	if a.voice == nil {
		return c.JSON(http.StatusOK, StatusResponse{
			Available: false,
			Message:   "voice service is not configured",
		})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), a.healthTTL)
	defer cancel()

	resp, err := a.voice.Health(ctx)
	if err != nil {
		return c.JSON(http.StatusOK, StatusResponse{
			Available: false,
			Message:   "voice service is unavailable",
		})
	}

	return c.JSON(http.StatusOK, StatusResponse{
		Available:     strings.EqualFold(resp.GetStatus(), "ok"),
		Status:        resp.GetStatus(),
		ModelProvider: resp.GetModelProvider(),
		ModelID:       resp.GetModelId(),
		Device:        resp.GetDevice(),
	})
}

func (a *app) transcribe(c *echo.Context) error {
	if a.voice == nil {
		return errs.Errorf(errs.Unavailable, "voice service is not configured")
	}

	reference, desc, err := parseReferenceAudio(c)
	if err != nil {
		return err
	}
	defer reference.Close()

	ctx, cancel := context.WithTimeout(c.Request().Context(), a.asrTimeout)
	defer cancel()

	text, err := a.voice.Transcribe(ctx, desc, reference, strings.TrimSpace(c.FormValue("language")))
	if err != nil {
		return mapBusErr(err, "transcribe")
	}

	return c.JSON(http.StatusOK, TranscribeResponse{Text: text})
}

func (a *app) clone(c *echo.Context) error {
	if a.voice == nil {
		return errs.Errorf(errs.Unavailable, "voice service is not configured")
	}

	req, err := parseCloneRequest(c)
	if err != nil {
		return err
	}
	defer req.reference.Close()

	ctx, cancel := context.WithTimeout(c.Request().Context(), a.cloneTimeout)
	defer cancel()

	var out bytes.Buffer
	result, err := a.voice.Synthesize(ctx, voiceadapter.SynthesizeRequest{
		RequestID:       uuid.NewString(),
		Text:            req.text,
		Mode:            voiceadapter.SynthesisModeVoiceClone,
		Language:        req.language,
		ReferenceText:   req.referenceText,
		Instruct:        req.instruct,
		ReferenceAudio:  req.referenceAudio,
		ReferenceReader: req.reference,
		GenerationConfig: voiceadapter.GenerationConfig{
			Speed:             req.speed,
			DurationSeconds:   req.durationSeconds,
			InferenceSteps:    req.inferenceSteps,
			GuidanceScale:     req.guidanceScale,
			Denoise:           req.denoise,
			PreprocessPrompt:  req.preprocessPrompt,
			PostprocessOutput: req.postprocessOutput,
		},
		OutputConfig: voiceadapter.OutputConfig{
			Format:    voiceadapter.OutputFormatOGGOpus,
			ChunkSize: a.outputChunkSize,
		},
	}, &out, nil)
	if err != nil {
		return mapBusErr(err, "clone")
	}

	extension := result.Extension
	if extension == "" {
		extension = ".ogg"
	}

	mimeType := result.MimeType
	if mimeType == "" {
		mimeType = "audio/ogg"
	}

	disposition := mime.FormatMediaType("inline", map[string]string{
		"filename": "voice-clone" + extension,
	})

	c.Response().Header().Set(echo.HeaderContentDisposition, disposition)
	c.Response().Header().Set(echo.HeaderXContentTypeOptions, "nosniff")
	c.Response().Header().Set("X-Voice-Request-Id", result.RequestID)
	c.Response().Header().Set("X-Voice-Extension", extension)
	c.Response().Header().Set("X-Voice-Duration-Ms", strconv.FormatInt(result.DurationMs, 10))
	c.Response().Header().Set("X-Voice-Sample-Rate-Hz", strconv.FormatInt(int64(result.SampleRateHz), 10))
	c.Response().Header().Set("X-Voice-Total-Bytes", strconv.FormatInt(result.TotalBytes, 10))

	return c.Stream(http.StatusOK, mimeType, bytes.NewReader(out.Bytes()))
}
