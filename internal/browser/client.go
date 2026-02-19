package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

const (
	DebugPort  = ":9222"
	ProfileDir = ".cache/scraping"
)

var execLock sync.Mutex

type Client struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func New() *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *Client) Close() {
	if c.cancel != nil {
		c.cancel()
	}
}

func FindChrome() string {
	paths := []string{
		os.Getenv("CHROME_PATH"),
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/usr/bin/google-chrome",
		"/usr/bin/chromium",
		"/usr/bin/chromium-browser",
	}
	for _, p := range paths {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func (c *Client) Start(useProfile bool) error {
	execLock.Lock()
	defer execLock.Unlock()

	killChrome()
	time.Sleep(500 * time.Millisecond)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	profilePath := filepath.Join(homeDir, ProfileDir)

	if useProfile {
		chromeProfileDir := ""
		if os.Getenv("HOME") != "" {
			chromeProfileDir = filepath.Join(os.Getenv("HOME"), "Library/Application Support/Google/Chrome")
		}

		if chromeProfileDir != "" {
			if err := os.MkdirAll(profilePath, 0755); err != nil {
				return fmt.Errorf("creating profile directory: %w", err)
			}

			cmd := exec.Command("rsync", "-a", "--delete",
				chromeProfileDir+"/",
				profilePath+"/")
			cmd.Stdout = os.Stderr
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not sync profile: %v\n", err)
			}
		}
	}

	chromePath := FindChrome()
	if chromePath == "" {
		return fmt.Errorf("could not find Chrome executable")
	}

	if err := os.MkdirAll(profilePath, 0755); err != nil {
		return fmt.Errorf("creating profile directory: %w", err)
	}

	cmd := exec.Command(chromePath,
		"--remote-debugging-port=9222",
		"--user-data-dir="+profilePath,
		"--no-first-run",
		"--no-default-browser-check")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting Chrome: %w", err)
	}

	if err := c.waitForChrome(30); err != nil {
		return fmt.Errorf("waiting for Chrome: %w", err)
	}

	return nil
}

func (c *Client) waitForChrome(seconds int) error {
	for i := 0; i < seconds*2; i++ {
		ctx, cancel := context.WithTimeout(c.ctx, 500*time.Millisecond)
		defer cancel()

		task := chromedp.Navigate("about:blank")
		if err := chromedp.Run(ctx, task); err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for Chrome")
}

func killChrome() {
	cmds := []string{
		"killall 'Google Chrome'",
		"killall 'Chromium'",
		"killall chrome",
	}
	for _, cmd := range cmds {
		exec.Command("sh", "-c", cmd).Run()
	}
}

func (c *Client) Connect() (context.Context, context.CancelFunc, error) {
	ctx, cancel := chromedp.NewRemoteAllocator(c.ctx, "http://localhost"+DebugPort)
	ctx, cancel = chromedp.NewContext(ctx)
	return ctx, cancel, nil
}

func (c *Client) Navigate(ctx context.Context, url string, newTab bool) error {
	if newTab {
		_, cancel := chromedp.NewContext(ctx)
		defer cancel()
	}
	return chromedp.Run(ctx, chromedp.Navigate(url))
}

func (c *Client) Eval(ctx context.Context, code string) (interface{}, error) {
	var result []byte
	task := chromedp.Evaluate(code, &result)
	if err := chromedp.Run(ctx, task); err != nil {
		return nil, err
	}
	return string(result), nil
}

func (c *Client) EvalAsync(ctx context.Context, code string) (interface{}, error) {
	js := fmt.Sprintf(`
		(async () => { return (async () => { return (%s) })() })()
	`, code)

	var result []byte
	task := chromedp.Evaluate(js, &result)
	if err := chromedp.Run(ctx, task); err != nil {
		return nil, err
	}
	return string(result), nil
}

func (c *Client) Screenshot(ctx context.Context) (string, error) {
	var buf []byte
	task := chromedp.FullScreenshot(&buf, 100)
	if err := chromedp.Run(ctx, task); err != nil {
		return "", err
	}

	tmpDir := os.TempDir()
	filename := fmt.Sprintf("screenshot-%s.png", time.Now().Format("2006-01-02-15-04-05"))
	filepath := filepath.Join(tmpDir, filename)

	if err := os.WriteFile(filepath, buf, 0644); err != nil {
		return "", err
	}

	return filepath, nil
}

type ElementInfo struct {
	Tag     string `json:"tag"`
	ID      string `json:"id,omitempty"`
	Class   string `json:"class,omitempty"`
	Text    string `json:"text,omitempty"`
	HTML    string `json:"html,omitempty"`
	Parents string `json:"parents,omitempty"`
}

func (c *Client) Pick(ctx context.Context, message string) ([]ElementInfo, error) {
	pickerCode := `
		(function() {
			return new Promise((resolve) => {
				const selections = [];
				const selectedElements = new Set();

				const overlay = document.createElement('div');
				overlay.style.cssText = 'position:fixed;top:0;left:0;width:100%;height:100%;z-index:2147483647;pointer-events:none';

				const highlight = document.createElement('div');
				highlight.style.cssText = 'position:absolute;border:2px solid #3b82f6;background:rgba(59,130,246,0.1);transition:all 0.1s';
				overlay.appendChild(highlight);

				const banner = document.createElement('div');
				banner.style.cssText = 'position:fixed;bottom:20px;left:50%;transform:translateX(-50%);background:#1f2937;color:white;padding:12px 24px;border-radius:8px;font:14px sans-serif;box-shadow:0 4px 12px rgba(0,0,0,0.3);pointer-events:auto;z-index:2147483647';

				const updateBanner = () => {
					banner.textContent = arguments[0] + ' (' + selections.length + ' selected, Cmd/Ctrl+click to add, Enter to finish, ESC to cancel)';
				};
				updateBanner();

				document.body.append(banner, overlay);

				const cleanup = () => {
					document.removeEventListener('mousemove', onMove, true);
					document.removeEventListener('click', onClick, true);
					document.removeEventListener('keydown', onKey, true);
					overlay.remove();
					banner.remove();
					selectedElements.forEach((el) => { el.style.outline = ''; });
				};

				const onMove = (e) => {
					const el = document.elementFromPoint(e.clientX, e.clientY);
					if (!el || overlay.contains(el) || banner.contains(el)) return;
					const r = el.getBoundingClientRect();
					highlight.style.cssText = 'position:absolute;border:2px solid #3b82f6;background:rgba(59,130,246,0.1);top:'+r.top+'px;left:'+r.left+'px;width:'+r.width+'px;height:'+r.height+'px';
				};

				const buildElementInfo = (el) => {
					const parents = [];
					let current = el.parentElement;
					while (current && current !== document.body) {
						const parentInfo = current.tagName.toLowerCase();
						const id = current.id ? '#'+current.id : '';
						const cls = current.className ? '.'+current.className.trim().split(/\s+/).join('.') : '';
						parents.push(parentInfo + id + cls);
						current = current.parentElement;
					}

					return {
						tag: el.tagName.toLowerCase(),
						id: el.id || null,
						class: el.className || null,
						text: el.textContent ? el.textContent.trim().slice(0, 200) : null,
						html: el.outerHTML ? el.outerHTML.slice(0, 500) : null,
						parents: parents.join(' > ')
					};
				};

				const onClick = (e) => {
					if (banner.contains(e.target)) return;
					e.preventDefault();
					e.stopPropagation();
					const el = document.elementFromPoint(e.clientX, e.clientY);
					if (!el || overlay.contains(el) || banner.contains(el)) return;

					if (e.metaKey || e.ctrlKey) {
						if (!selectedElements.has(el)) {
							selectedElements.add(el);
							el.style.outline = '3px solid #10b981';
							selections.push(buildElementInfo(el));
							updateBanner();
						}
					} else {
						cleanup();
						const info = buildElementInfo(el);
						resolve(selections.length > 0 ? selections : info);
					}
				};

				const onKey = (e) => {
					if (e.key === 'Escape') {
						e.preventDefault();
						cleanup();
						resolve(null);
					} else if (e.key === 'Enter' && selections.length > 0) {
						e.preventDefault();
						cleanup();
						resolve(selections);
					}
				};

				document.addEventListener('mousemove', onMove, true);
				document.addEventListener('click', onClick, true);
				document.addEventListener('keydown', onKey, true);
			});
		})
	`

	var result []byte
	task := chromedp.Evaluate(pickerCode+"('"+message+"')", &result)
	if err := chromedp.Run(ctx, task); err != nil {
		return nil, err
	}

	var parsed interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		return nil, err
	}

	if parsed == nil {
		return nil, nil
	}

	var elements []ElementInfo
	switch v := parsed.(type) {
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				elements = append(elements, convertToElementInfo(m))
			}
		}
	case map[string]interface{}:
		elements = append(elements, convertToElementInfo(v))
	}

	return elements, nil
}

func convertToElementInfo(m map[string]interface{}) ElementInfo {
	info := ElementInfo{}
	if v, ok := m["tag"].(string); ok {
		info.Tag = v
	}
	if v, ok := m["id"]; ok && v != nil {
		if s, ok := v.(string); ok {
			info.ID = s
		}
	}
	if v, ok := m["class"]; ok && v != nil {
		if s, ok := v.(string); ok {
			info.Class = s
		}
	}
	if v, ok := m["text"]; ok && v != nil {
		if s, ok := v.(string); ok {
			info.Text = s
		}
	}
	if v, ok := m["html"]; ok && v != nil {
		if s, ok := v.(string); ok {
			info.HTML = s
		}
	}
	if v, ok := m["parents"]; ok && v != nil {
		if s, ok := v.(string); ok {
			info.Parents = s
		}
	}
	return info
}

type Cookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain,omitempty"`
	Path     string `json:"path,omitempty"`
	Secure   bool   `json:"secure"`
	HTTPOnly bool   `json:"httpOnly"`
}

func (c *Client) GetCookies(ctx context.Context) ([]Cookie, error) {
	var result []byte
	task := chromedp.Evaluate(`JSON.stringify(document.cookie)`, &result)
	if err := chromedp.Run(ctx, task); err != nil {
		return nil, err
	}

	var cookies []Cookie
	cookieStr := string(result)
	if cookieStr != "" {
		pairs := splitCookieString(cookieStr)
		for _, pair := range pairs {
			parts := splitKeyValue(pair)
			if len(parts) == 2 {
				cookies = append(cookies, Cookie{Name: parts[0], Value: parts[1]})
			}
		}
	}
	return cookies, nil
}

func splitCookieString(s string) []string {
	var pairs []string
	var current string
	inQuote := false
	for _, c := range s {
		if c == '"' {
			inQuote = !inQuote
		}
		if c == ';' && !inQuote {
			pairs = append(pairs, current)
			current = ""
		} else if c != ' ' || current != "" {
			current += string(c)
		}
	}
	if current != "" {
		pairs = append(pairs, current)
	}
	return pairs
}

func splitKeyValue(s string) []string {
	for i, c := range s {
		if c == '=' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s, ""}
}

func (c *Client) SetCookies(ctx context.Context, cookies []Cookie) error {
	for _, ck := range cookies {
		js := fmt.Sprintf("document.cookie = '%s=%s'", ck.Name, ck.Value)
		var result []byte
		task := chromedp.Evaluate(js, &result)
		if err := chromedp.Run(ctx, task); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) GetCookiesJSON(ctx context.Context) (string, error) {
	cookies, err := c.GetCookies(ctx)
	if err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(cookies, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
