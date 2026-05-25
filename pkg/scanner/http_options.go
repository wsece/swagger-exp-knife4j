// http_options.go: HTTP client config for scans (headers, cookies, proxy, timeout, concurrency) and Validate/Client/Apply.
package scanner

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// DefaultUserAgent is used when -A/--user-agent is not set
const DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36"

const defaultConnectTimeout = 30 * time.Second

// HTTPOptions customizes HTTP requests (curl -H / -A / -b / -x plus concurrency/timeout flags)
type HTTPOptions struct {
	Headers   []string
	UserAgent string
	Cookies   []string
	Proxy     string

	Delay          time.Duration
	RequestTimeout time.Duration // 0 = unlimited
	ConnectTimeout time.Duration
	Parallel       int
}

// HTTPRequestMeta is request metadata for one scan (persisted to DB)
type HTTPRequestMeta struct {
	Cookie        string
	UserAgent     string
	Authorization string
}

// Meta returns Cookie / UA / Authorization for persistence
func (o *HTTPOptions) Meta() (HTTPRequestMeta, error) {
	cookie, err := o.cookieHeader()
	if err != nil {
		return HTTPRequestMeta{}, err
	}
	return HTTPRequestMeta{
		Cookie:        cookie,
		UserAgent:     userAgentFor(o),
		Authorization: o.headerValue("Authorization"),
	}, nil
}

func (o *HTTPOptions) headerValue(name string) string {
	if o == nil {
		return ""
	}
	for _, h := range o.Headers {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		key, val, ok := strings.Cut(h, ":")
		if !ok {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(key), name) {
			return strings.TrimSpace(val)
		}
	}
	return ""
}

// Client builds an *http.Client with proxy and timeouts.
// fallbackTimeout is used as Client.Timeout when RequestTimeout is 0; 0 means no overall request limit.
// Returns the configured client or an error if the proxy URL is invalid.
func (o *HTTPOptions) Client(fallbackTimeout time.Duration) (*http.Client, error) {
	proxy := ""
	connectTimeout := defaultConnectTimeout
	reqTimeout := fallbackTimeout

	if o != nil {
		proxy = o.Proxy
		if o.ConnectTimeout > 0 {
			connectTimeout = o.ConnectTimeout
		}
		if o.RequestTimeout > 0 {
			reqTimeout = o.RequestTimeout
		}
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	dialer := &net.Dialer{Timeout: connectTimeout}
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, addr)
	}
	if proxy != "" {
		proxyURL, err := parseProxyURL(proxy)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	return &http.Client{Timeout: reqTimeout, Transport: transport}, nil
}

// Validate checks concurrency, delay, timeout flags; returns error if invalid.
func (o *HTTPOptions) Validate() error {
	if o == nil {
		return nil
	}
	if o.Parallel < 1 {
		return fmt.Errorf("parallel (-P) must be >= 1")
	}
	if o.Delay < 0 {
		return fmt.Errorf("delay must be >= 0")
	}
	if o.ConnectTimeout < 0 {
		return fmt.Errorf("connect-timeout must be >= 0")
	}
	if o.RequestTimeout < 0 {
		return fmt.Errorf("request timeout (-m) must be >= 0")
	}
	return nil
}

// Apply sets User-Agent, custom headers, and cookies on req; returns error on bad header format.
func (o *HTTPOptions) Apply(req *http.Request) error {
	req.Header.Set("User-Agent", userAgentFor(o))

	if o == nil {
		return nil
	}

	for _, h := range o.Headers {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		key, val, ok := strings.Cut(h, ":")
		if !ok {
			return fmt.Errorf("invalid header %q, expected \"Key: Value\"", h)
		}
		req.Header.Set(strings.TrimSpace(key), strings.TrimSpace(val))
	}

	cookie, err := o.cookieHeader()
	if err != nil {
		return err
	}
	if cookie != "" {
		if existing := req.Header.Get("Cookie"); existing != "" {
			req.Header.Set("Cookie", existing+"; "+cookie)
		} else {
			req.Header.Set("Cookie", cookie)
		}
	}

	return nil
}

func userAgentFor(o *HTTPOptions) string {
	if o != nil && strings.TrimSpace(o.UserAgent) != "" {
		return o.UserAgent
	}
	return DefaultUserAgent
}

func (o *HTTPOptions) cookieHeader() (string, error) {
	if o == nil || len(o.Cookies) == 0 {
		return "", nil
	}

	parts := make([]string, 0, len(o.Cookies))
	for _, raw := range o.Cookies {
		val, err := loadCookieValue(strings.TrimSpace(raw))
		if err != nil {
			return "", err
		}
		if val != "" {
			parts = append(parts, val)
		}
	}
	return strings.Join(parts, "; "), nil
}

func loadCookieValue(flag string) (string, error) {
	if flag == "" {
		return "", nil
	}

	if st, err := os.Stat(flag); err == nil && !st.IsDir() {
		data, err := os.ReadFile(flag)
		if err != nil {
			return "", fmt.Errorf("read cookie file %q: %w", flag, err)
		}
		return strings.TrimSpace(string(data)), nil
	}

	return flag, nil
}

func parseProxyURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("proxy url is empty")
	}
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy %q: %w", raw, err)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("invalid proxy %q", raw)
	}
	return u, nil
}

func fetchWithOptions(client *http.Client, httpOpts *HTTPOptions, rawURL string) ([]byte, string, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", err
	}
	if err := httpOpts.Apply(req); err != nil {
		return nil, "", err
	}
	return doRequest(client, req)
}

func doRequest(client *http.Client, req *http.Request) ([]byte, string, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := readLimitedBody(resp.Body)
	if err != nil {
		return nil, "", err
	}
	return body, resp.Header.Get("Content-Type"), nil
}

func readLimitedBody(r io.Reader) ([]byte, error) {
	return io.ReadAll(io.LimitReader(r, maxSwaggerFetchBytes))
}
