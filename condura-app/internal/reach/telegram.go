package reach

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
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
//
// baseURL is the Bot API origin. Production sets it to
// "https://api.telegram.org" (see apiCall). Tests override it via
// setBaseURL to point at an httptest.NewServer so the channel can
// be exercised end-to-end without hitting api.telegram.org.
//
// mu protects the chatID and connected fields. The receive
// goroutine (started by Manager.startReceive) writes chatID when
// the first inbound message arrives; the main goroutine reads it
// from Send / Status. Without the mutex these unsynchronized
// accesses race and -race catches them. Read paths (Send, Status)
// use the helpers below rather than touching the fields directly.
type telegramChannel struct {
	token      string
	chatID     string
	connected  bool
	httpClient *http.Client
	baseURL    string
	mu         sync.Mutex
}

func newTelegramChannel() *telegramChannel {
	return &telegramChannel{
		httpClient: &http.Client{Timeout: 60 * time.Second},
		baseURL:    "https://api.telegram.org",
	}
}

// setBaseURL overrides the Bot API origin. Pass the empty string
// to reset to the default. Production code should not call this;
// the integration test suite is the only intended caller.
func (t *telegramChannel) setBaseURL(s string) {
	if s == "" {
		s = "https://api.telegram.org"
	}
	t.baseURL = s
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
	t.mu.Lock()
	t.token = token
	t.mu.Unlock()
	resp, err := t.apiCall(ctx, "getMe", nil)
	if err != nil {
		t.mu.Lock()
		t.token = ""
		t.mu.Unlock()
		return fmt.Errorf("reach: telegram connect: %w", err)
	}
	var me struct {
		OK     bool `json:"ok"`
		Result struct {
			Username string `json:"username"`
		} `json:"result"`
	}
	if err := json.Unmarshal(resp, &me); err != nil || !me.OK {
		t.mu.Lock()
		t.token = ""
		t.mu.Unlock()
		return fmt.Errorf("reach: invalid telegram token")
	}
	t.mu.Lock()
	t.connected = true
	t.mu.Unlock()
	return nil
}

func (t *telegramChannel) Disconnect(_ context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.token = ""
	t.chatID = ""
	t.connected = false
	return nil
}

// setChatID stores the captured chat id under the channel's mutex.
// Called from Manager.startReceive after the first inbound message
// is observed, so subsequent Send("") calls route correctly.
func (t *telegramChannel) setChatID(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.chatID = id
}

// getChatID returns the current chat id under the channel's mutex.
func (t *telegramChannel) getChatID() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.chatID
}

func (t *telegramChannel) Send(ctx context.Context, chatID, text string) error {
	if !t.connected {
		return fmt.Errorf("reach: telegram not connected")
	}
	if chatID == "" {
		chatID = t.getChatID()
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
	t.mu.Lock()
	defer t.mu.Unlock()
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
	u := fmt.Sprintf("%s/bot%s/%s", t.baseURL, t.token, method)
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
