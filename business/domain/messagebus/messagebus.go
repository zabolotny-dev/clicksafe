package messagebus

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
	"github.com/zabolotny-dev/clicksafe/business/types/file"
)

var (
	ErrUniqueLabel               = errors.New("Message with this label already exists")
	ErrContentNotFound           = errors.New("message content not found")
	ErrEmptyContent              = errors.New("message content is empty")
	ErrUnsupportedTemplateSyntax = errors.New("unsupported template syntax")
)

type Storer interface {
	Save(ctx context.Context, msg Message) error
	Update(ctx context.Context, msg Message) error
	Delete(ctx context.Context, msg Message) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Message, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

type FileStorage interface {
	Save(ctx context.Context, r io.Reader, ext string) (file.Path, error)
	Read(ctx context.Context, p file.Path) ([]byte, error)
	Delete(ctx context.Context, p file.Path) error
}

type Business struct {
	storer    Storer
	fileStore FileStorage
}

func NewBusiness(s Storer, fileStore FileStorage) *Business {
	return &Business{storer: s, fileStore: fileStore}
}

func (b *Business) Save(ctx context.Context, msg NewMessage) (Message, error) {
	message := Message{
		ID:        uuid.New(),
		Label:     msg.Label,
		FromEmail: msg.FromEmail,
		FromName:  msg.FromName,
		Subject:   msg.Subject,
	}

	if err := b.storer.Save(ctx, message); err != nil {
		return Message{}, fmt.Errorf("save: %w", err)
	}

	return message, nil
}

func (b *Business) Update(ctx context.Context, msg Message, up UpdateMessage) (Message, error) {
	if up.Label != nil {
		msg.Label = *up.Label
	}
	if up.FromEmail != nil {
		msg.FromEmail = *up.FromEmail
	}
	if up.FromName != nil {
		msg.FromName = *up.FromName
	}
	if up.Subject != nil {
		msg.Subject = *up.Subject
	}

	if err := b.storer.Update(ctx, msg); err != nil {
		return Message{}, fmt.Errorf("update: %w", err)
	}

	return msg, nil
}

func (b *Business) Delete(ctx context.Context, msg Message) error {
	if err := b.storer.Delete(ctx, msg); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Message, error) {
	messages, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return messages, nil
}

func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}

func (b *Business) SaveContent(ctx context.Context, msg Message, r io.Reader) (Message, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return Message{}, fmt.Errorf("savecontent: read: %w", err)
	}

	if len(bytes.TrimSpace(content)) == 0 {
		return Message{}, ErrEmptyContent
	}

	requiredVars, err := validateAndExtractRequiredVars(content)
	if err != nil {
		return Message{}, fmt.Errorf("savecontent: %w", err)
	}

	newPath, err := b.fileStore.Save(ctx, bytes.NewReader(content), ".html")
	if err != nil {
		return Message{}, fmt.Errorf("savecontent: save file: %w", err)
	}

	oldPath := msg.ContentPath

	msg.ContentPath = file.NewNullPath(newPath)
	msg.RequiredVars = requiredVars

	if err := b.storer.Update(ctx, msg); err != nil {
		_ = b.fileStore.Delete(ctx, newPath)
		return Message{}, fmt.Errorf("savecontent: update: %w", err)
	}

	if oldPath.Valid() && !oldPath.Path().Equal(newPath) {
		_ = b.fileStore.Delete(ctx, oldPath.Path())
	}

	return msg, nil
}

func (b *Business) ReadContent(ctx context.Context, msg Message) ([]byte, error) {
	if !msg.ContentPath.Valid() {
		return nil, ErrContentNotFound
	}

	content, err := b.fileStore.Read(ctx, msg.ContentPath.Path())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrContentNotFound
		}
		return nil, fmt.Errorf("readcontent: read file: %w", err)
	}

	return content, nil
}
