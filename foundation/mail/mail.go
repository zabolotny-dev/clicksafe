package mail

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	gomail "github.com/wneessen/go-mail"
)

var (
	ErrHostRequired       = errors.New("mail host is required")
	ErrInvalidCredentials = errors.New("mail username and password must be provided together")
	ErrInvalidTLSPolicy   = errors.New("invalid mail tls policy")
	ErrClientRequired     = errors.New("mail client is required")
)

type TLSPolicy string

const (
	TLSMandatory     TLSPolicy = "mandatory"
	TLSOpportunistic TLSPolicy = "opportunistic"
	NoTLS            TLSPolicy = "none"
)

type Config struct {
	Host      string
	Port      int
	Username  string
	Password  string
	Timeout   time.Duration
	TLSPolicy TLSPolicy
	SSL       bool
}

type Client struct {
	client *gomail.Client
}

func New(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.Host) == "" {
		return nil, ErrHostRequired
	}

	if (cfg.Username == "") != (cfg.Password == "") {
		return nil, ErrInvalidCredentials
	}

	policy, err := toGoMailTLSPolicy(cfg.TLSPolicy)
	if err != nil {
		return nil, err
	}

	opts := []gomail.Option{}

	if cfg.Port != 0 {
		opts = append(opts, gomail.WithPort(cfg.Port))
	}

	if cfg.SSL {
		if cfg.Port == 0 {
			opts = append(opts, gomail.WithSSLPort(false))
		} else {
			opts = append(opts, gomail.WithSSL())
		}
	} else if cfg.Port == 0 {
		opts = append(opts, gomail.WithTLSPortPolicy(policy))
	} else {
		opts = append(opts, gomail.WithTLSPolicy(policy))
	}

	if cfg.Timeout != 0 {
		opts = append(opts, gomail.WithTimeout(cfg.Timeout))
	}

	if cfg.Username != "" {
		opts = append(opts,
			gomail.WithSMTPAuth(gomail.SMTPAuthAutoDiscover),
			gomail.WithUsername(cfg.Username),
			gomail.WithPassword(cfg.Password),
		)
	}

	client, err := gomail.NewClient(cfg.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf("new mail client: %w", err)
	}

	return &Client{client: client}, nil
}

func (c *Client) Send(ctx context.Context, to, from, subject, body string) error {
	if c == nil || c.client == nil {
		return ErrClientRequired
	}

	msg := gomail.NewMsg()
	if err := msg.From(from); err != nil {
		return fmt.Errorf("from: %w", err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("to: %w", err)
	}
	msg.Subject(subject)
	msg.SetBodyString(gomail.TypeTextHTML, body)

	if err := c.client.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("send: %w", err)
	}

	return nil
}

func toGoMailTLSPolicy(policy TLSPolicy) (gomail.TLSPolicy, error) {
	switch strings.ToLower(strings.TrimSpace(string(policy))) {
	case "", "mandatory", "tlsmandatory":
		return gomail.TLSMandatory, nil
	case "opportunistic", "tlsopportunistic":
		return gomail.TLSOpportunistic, nil
	case "none", "no_tls", "notls", "disabled":
		return gomail.NoTLS, nil
	default:
		return 0, fmt.Errorf("%w: %q", ErrInvalidTLSPolicy, policy)
	}
}
