package reach

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/safego"
)

// telegramChannel implements the Channel interface for Telegram
// Bot API.
//
// httpClient is injectable so tests can swap it for an httptest
// transport. In production it is a *http.Client with a 60s timeout,
// which is comfortably larger than Telegram's long-poll window
// (30s by default — see the timeout param on getUpdates below).
type telegramChannel struct {
	token      string
	chatID     string
	connected  bool
	httpClient *http.Client
}

func newTelegramChannel() *telegramChannel {
	return &telegramChannel{
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// setHTTPClient replaces the HTTP client used for Telegram API
// calls. Tests use this to wire in an httptest.NewServer; production
// code should not need to call it.
func (t *telegramChannel) setHTTPClient(c *http.Client) {
	if c == nil {
		return
	}
	t.httpClient = c
}

func (t *telegramChannel) Connect(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("reach: telegram token is empty")
	}
	// Set t.token before the validation call: apiCall reads
	// t.token to build the bot URL. If validation fails we
	// roll back below so an invalid token never leaves a
	// half-configured channel behind.
	t.token = token
	resp, err := t.apiCall(ctx, "getMe", nil)
	if err != nil {
		t.token = ""
		return fmt.Errorf("reach: telegram connect: %w", err)
	}
	var me struct {
		OK     bool `json:"ok"`
		Result struct {
			Username string `json:"username"`
		} `json:"result"`
	}
	if err := json.Unmarshal(resp, &me); err != nil || !me.OK {
		t.token = ""
		return fmt.Errorf("reach: invalid telegram token")
	}
	t.connected = true
	return nil
}

func (t *telegramChannel) Disconnect(_ context.Context) error {
	t.token = ""
	t.chatID = ""
	t.connected = false
	return nil
}

func (t *telegramChannel) Send(ctx context.Context, chatID, text string) error {
	if !t.connected {
		return fmt.Errorf("reach: telegram not connected")
	}
	if chatID == "" {
		chatID = t.chatID
	}
	if chatID == "" {
		return fmt.Errorf("reach: telegram chat_id unknown — message the bot once to establish")
	}
	_, err := t.apiCall(ctx, "sendMessage", url.Values{
		"chat_id": {chatID},
		"text":    {text},
	})
	return err
}

func (t *telegramChannel) Receive(ctx context.Context) (<-chan Message, error) {
	if !t.connected {
		return nil, fmt.Errorf("reach: telegram not connected")
	}
	ch := make(chan Message, 10)
	safego.Go(func() { t.longPoll(ctx, ch) })
	return ch, nil
}

func (t *telegramChannel) Status(_ context.Context) (ChannelStatus, error) {
	return ChannelStatus{
		Name:      "telegram",
		Connected: t.connected,
		ChatID:    t.chatID,
	}, nil
}

func (t *telegramChannel) longPoll(ctx context.Context, out chan<- Message) {
	defer close(out)
	var offset int64
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		params := url.Values{"timeout": {"30"}}
		if offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", offset+1))
		}
		resp, err := t.apiCall(ctx, "getUpdates", params)
		if err != nil {
			// Wait briefly on error, then retry. Bounded so a
			// daemon shutdown can stop us within ~2s.
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
		}
		var updates struct {
			OK     bool `json:"ok"`
			Result []struct {
				UpdateID int64 `json:"update_id"`
				Message  struct {
					Chat struct {
						ID int64 `json:"id"`
					} `json:"chat"`
					From struct {
						Username string `json:"username"`
					} `json:"from"`
					Text string `json:"text"`
				} `json:"message"`
			} `json:"result"`
		}
		if err := json.Unmarshal(resp, &updates); err != nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
		}
		for _, upd := range updates.Result {
			if upd.UpdateID > offset {
				offset = upd.UpdateID
			}
			if upd.Message.Text != "" {
				select {
				case out <- Message{
					ChatID:  fmt.Sprintf("%d", upd.Message.Chat.ID),
					Sender:  upd.Message.From.Username,
					Text:    upd.Message.Text,
					Channel: "telegram",
				}:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func (t *telegramChannel) apiCall(ctx context.Context, method string, params url.Values) ([]byte, error) {
	if t.token == "" {
		return nil, fmt.Errorf("reach: telegram token not set")
	}
	u := fmt.Sprintf("https://api.telegram.org/bot%s/%s", t.token, method)
	var req *http.Request
	var err error
	if params != nil {
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(params.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	}
	if err != nil {
		return nil, err
	}
	client := t.httpClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	return io.ReadAll(resp.Body)
}
