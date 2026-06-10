package messagebus

import (
	"context"
	"fmt"
	"math/rand"
	"net/mail"

	"github.com/zabolotny-dev/clicksafe/business/types/label"
	"github.com/zabolotny-dev/clicksafe/business/types/subject"
)

// TestSeedEmailMessages creates n email messages via the business layer.
func TestSeedEmailMessages(ctx context.Context, n int, api *Business) ([]Message, error) {
	msgs := make([]Message, 0, n)

	idx := rand.Intn(10000)
	for range n {
		idx++

		subj, _ := subject.ParseNull(fmt.Sprintf("Subject %d", idx))

		nm := NewMessage{
			Type:      EmailMessage,
			Label:     label.MustParse(fmt.Sprintf("Message %d", idx)),
			FromEmail: mail.Address{Address: fmt.Sprintf("sender%d@example.com", idx)},
			Subject:   subj,
		}

		msg, err := api.Save(ctx, nm)
		if err != nil {
			return nil, fmt.Errorf("seeding email message idx[%d]: %w", idx, err)
		}

		msgs = append(msgs, msg)
	}

	return msgs, nil
}
