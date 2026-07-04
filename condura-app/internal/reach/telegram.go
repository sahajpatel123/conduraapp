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
)

// telegramChannel implements the Channel interface for Telegram
// Bot API.
type telegramChannel struct {
	token     string
	chatID    string
	connected bool
}

func newTelegramChannel() *telegramChannel {
	return &telegramChannel{}
}

func (t *telegramChannel) Connect(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("reach: telegram token is empty")
	}
	// Validate the token by calling getMe.
	resp, err := t.apiCall(ctx, "getMe", nil)
	if err != nil {
		return fmt.Errorf("reach: telegram connect: %w", err)
	}
	var me struct {
		OK     bool `json:"ok"`
		Result struct {
			Username string `json:"username"`
		} `json:"result"`
	}
	if err := json.Unmarshal(resp, &me); err != nil || !me.OK {
		return fmt.Errorf("reach: invalid telegram token")
	}
	t.token = token
	t.connected = true
	return nil
}

func (t *telegramChannel) Disconnect(ctx context.Context) error {
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
	go t.longPoll(ctx, ch)
	return ch, nil
}

func (t *telegramChannel) Status(ctx context.Context) (ChannelStatus, error) {
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
			time.Sleep(2 * time.Second)
			continue
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
			time.Sleep(2 * time.Second)
			continue
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
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	return io.ReadAll(resp.Body)
}
