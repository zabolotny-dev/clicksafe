package voiceapp

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/voiceadapter"
)

type StatusResponse struct {
	Available     bool   `json:"available"`
	Message       string `json:"message,omitempty"`
	Status        string `json:"status,omitempty"`
	ModelProvider string `json:"model_provider,omitempty"`
	ModelID       string `json:"model_id,omitempty"`
	Device        string `json:"device,omitempty"`
}

type TranscribeResponse struct {
	Text string `json:"text"`
}

type cloneRequest struct {
	text              string
	referenceText     string
	language          string
	instruct          string
	reference         multipartFile
	referenceAudio    voiceadapter.AudioDescriptor
	speed             *float64
	durationSeconds   *float64
	inferenceSteps    *int32
	guidanceScale     *float64
	denoise           *bool
	preprocessPrompt  *bool
	postprocessOutput *bool
}

type multipartFile interface {
	Read([]byte) (int, error)
	Close() error
}

func parseCloneRequest(c *echo.Context) (cloneRequest, error) {
	var fieldErrors errs.FieldErrors

	text := strings.TrimSpace(c.FormValue("text"))
	if text == "" {
		fieldErrors.Add("text", errors.New("text is required"))
	}

	if len(fieldErrors) > 0 {
		return cloneRequest{}, fieldErrors.ToError(errs.InvalidArgument, "validation failed")
	}

	file, desc, err := parseReferenceAudio(c)
	if err != nil {
		return cloneRequest{}, err
	}

	speed, err := optionalFloat(c, "speed")
	if err != nil {
		file.Close()
		return cloneRequest{}, err
	}

	durationSeconds, err := optionalPositiveFloat(c, "duration_seconds")
	if err != nil {
		file.Close()
		return cloneRequest{}, err
	}

	inferenceSteps, err := optionalInt32(c, "inference_steps")
	if err != nil {
		file.Close()
		return cloneRequest{}, err
	}

	guidanceScale, err := optionalFloat(c, "guidance_scale")
	if err != nil {
		file.Close()
		return cloneRequest{}, err
	}

	denoise, err := optionalBool(c, "denoise")
	if err != nil {
		file.Close()
		return cloneRequest{}, err
	}

	preprocessPrompt, err := optionalBool(c, "preprocess_prompt")
	if err != nil {
		file.Close()
		return cloneRequest{}, err
	}

	postprocessOutput, err := optionalBool(c, "postprocess_output")
	if err != nil {
		file.Close()
		return cloneRequest{}, err
	}

	return cloneRequest{
		text:              text,
		referenceText:     strings.TrimSpace(c.FormValue("reference_text")),
		language:          strings.TrimSpace(c.FormValue("language")),
		instruct:          strings.TrimSpace(c.FormValue("instruct")),
		reference:         file,
		referenceAudio:    desc,
		speed:             speed,
		durationSeconds:   durationSeconds,
		inferenceSteps:    inferenceSteps,
		guidanceScale:     guidanceScale,
		denoise:           denoise,
		preprocessPrompt:  preprocessPrompt,
		postprocessOutput: postprocessOutput,
	}, nil
}

func parseReferenceAudio(c *echo.Context) (multipartFile, voiceadapter.AudioDescriptor, error) {
	fileHeader, err := c.FormFile("reference_audio")
	if err != nil {
		return nil, voiceadapter.AudioDescriptor{}, errs.NewFieldErrors("reference_audio", errors.New("reference audio is required"), errs.InvalidArgument, "validation failed")
	}

	base := filepath.Base(fileHeader.Filename)
	extension := strings.ToLower(filepath.Ext(base))
	attachmentType, err := attachmentbus.Parse(extension)
	if err != nil || !attachmentType.IsAudio() {
		return nil, voiceadapter.AudioDescriptor{}, errs.NewFieldErrors("reference_audio", fmt.Errorf("unsupported audio type: %s", extension), errs.InvalidArgument, "validation failed")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, voiceadapter.AudioDescriptor{}, errs.Errorf(errs.InvalidArgument, "open reference audio: %s", err)
	}

	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = attachmentType.MIMEType()
	}

	return file, voiceadapter.AudioDescriptor{Filename: base, MimeType: mimeType, Extension: extension, Size: fileHeader.Size}, nil
}

func optionalFloat(c *echo.Context, field string) (*float64, error) {
	raw := strings.TrimSpace(c.FormValue(field))
	if raw == "" {
		return nil, nil
	}

	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil, errs.NewFieldErrors(field, err, errs.InvalidArgument, "validation failed")
	}
	return &value, nil
}

func optionalPositiveFloat(c *echo.Context, field string) (*float64, error) {
	value, err := optionalFloat(c, field)
	if err != nil || value == nil {
		return value, err
	}
	if *value <= 0 {
		return nil, nil
	}
	return value, nil
}

func optionalInt32(c *echo.Context, field string) (*int32, error) {
	raw := strings.TrimSpace(c.FormValue(field))
	if raw == "" {
		return nil, nil
	}

	value, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return nil, errs.NewFieldErrors(field, err, errs.InvalidArgument, "validation failed")
	}

	parsed := int32(value)
	return &parsed, nil
}

func optionalBool(c *echo.Context, field string) (*bool, error) {
	raw := strings.TrimSpace(c.FormValue(field))
	if raw == "" {
		return nil, nil
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, errs.NewFieldErrors(field, err, errs.InvalidArgument, "validation failed")
	}
	return &value, nil
}
