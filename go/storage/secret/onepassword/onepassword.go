package onepassword

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strings"

	"hop.top/kit/go/storage/secret"
)

type Mode int

const (
	ModeCLI Mode = iota
	ModeConnect
)

type Store struct {
	mode       Mode
	vault      string
	connectURL string
	token      string
	client     *http.Client
}

func NewCLI(vault string) *Store { return &Store{mode: ModeCLI, vault: vault} }

func NewConnect(url, token, vault string) *Store {
	return &Store{mode: ModeConnect, vault: vault, connectURL: strings.TrimRight(url, "/"), token: token, client: &http.Client{}}
}

func (s *Store) Get(ctx context.Context, key string) (*secret.Secret, error) {
	if s.mode == ModeCLI {
		return s.cliGet(ctx, key)
	}
	return s.connectGet(ctx, key)
}

func (s *Store) List(ctx context.Context, prefix string) ([]string, error) {
	if s.mode == ModeCLI {
		return s.cliList(ctx, prefix)
	}
	return s.connectList(ctx, prefix)
}

func (s *Store) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.Get(ctx, key)
	if err == secret.ErrNotFound {
		return false, nil
	}
	return err == nil, err
}

func (s *Store) Set(ctx context.Context, key string, value []byte) error {
	if s.mode == ModeCLI {
		return secret.ErrNotSupported
	}
	return s.connectSet(ctx, key, value)
}

func (s *Store) Delete(ctx context.Context, key string) error {
	if s.mode == ModeCLI {
		return secret.ErrNotSupported
	}
	return s.connectDelete(ctx, key)
}

func (s *Store) cliGet(ctx context.Context, key string) (*secret.Secret, error) {
	out, err := exec.CommandContext(ctx, "op", "item", "get", "--vault", s.vault, "--fields", "password", "--format", "json", "--", key).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return nil, secret.ErrNotFound
		}
		return nil, fmt.Errorf("onepassword: op exec: %w", err)
	}
	var f struct{ Value string }
	if err := json.Unmarshal(out, &f); err != nil {
		return nil, fmt.Errorf("onepassword: parse: %w", err)
	}
	return &secret.Secret{Key: key, Value: []byte(f.Value)}, nil
}

func (s *Store) cliList(ctx context.Context, prefix string) ([]string, error) {
	raw, err := exec.CommandContext(ctx, "op", "item", "list", "--vault", s.vault, "--format", "json").Output()
	if err != nil {
		return nil, fmt.Errorf("onepassword: list: %w", err)
	}
	var items []struct{ Title string }
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("onepassword: parse list: %w", err)
	}
	var result []string
	for _, it := range items {
		if strings.HasPrefix(it.Title, prefix) {
			result = append(result, it.Title)
		}
	}
	return result, nil
}

type connectItem struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Fields []struct {
		Label string `json:"label"`
		Value string `json:"value"`
	} `json:"fields"`
}

func (s *Store) doReq(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, s.connectURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Content-Type", "application/json")
	return s.client.Do(req)
}

func (s *Store) connectGet(ctx context.Context, key string) (*secret.Secret, error) {
	params := url.Values{"filter": {fmt.Sprintf("title eq %q", key)}}
	resp, err := s.doReq(ctx, http.MethodGet, "/v1/vaults/"+url.PathEscape(s.vault)+"/items?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var items []connectItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("onepassword: decode: %w", err)
	}
	if len(items) == 0 {
		return nil, secret.ErrNotFound
	}
	for _, f := range items[0].Fields {
		if f.Label == "password" {
			return &secret.Secret{Key: key, Value: []byte(f.Value)}, nil
		}
	}
	return &secret.Secret{Key: key, Value: nil}, nil
}

func (s *Store) connectList(ctx context.Context, prefix string) ([]string, error) {
	resp, err := s.doReq(ctx, http.MethodGet, "/v1/vaults/"+url.PathEscape(s.vault)+"/items", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var items []connectItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("onepassword: decode list: %w", err)
	}
	var out []string
	for _, it := range items {
		if strings.HasPrefix(it.Title, prefix) {
			out = append(out, it.Title)
		}
	}
	return out, nil
}

func (s *Store) connectSet(ctx context.Context, key string, value []byte) error {
	body, err := json.Marshal(connectItem{
		Title: key,
		Fields: []struct {
			Label string `json:"label"`
			Value string `json:"value"`
		}{
			{Label: "password", Value: string(value)},
		},
	})
	if err != nil {
		return fmt.Errorf("onepassword: marshal: %w", err)
	}
	resp, err := s.doReq(ctx, http.MethodPost, "/v1/vaults/"+url.PathEscape(s.vault)+"/items", bytes.NewReader(body))
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("onepassword: set: status %d", resp.StatusCode)
	}
	return nil
}

func (s *Store) connectDelete(ctx context.Context, key string) error {
	params := url.Values{"filter": {fmt.Sprintf("title eq %q", key)}}
	resp, err := s.doReq(ctx, http.MethodGet, "/v1/vaults/"+url.PathEscape(s.vault)+"/items?"+params.Encode(), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var items []connectItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return err
	}
	if len(items) == 0 {
		return secret.ErrNotFound
	}
	del, err := s.doReq(ctx, http.MethodDelete, "/v1/vaults/"+url.PathEscape(s.vault)+"/items/"+url.PathEscape(items[0].ID), nil)
	if err != nil {
		return err
	}
	del.Body.Close()
	return nil
}

var _ secret.Store = (*Store)(nil)
var _ secret.MutableStore = (*Store)(nil)
