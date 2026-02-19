# Browser Tools

Minimal CDP tools for collaborative site exploration.

## Usage

```bash
./gogogo browser <command>
```

## Commands

### Start Chrome

```bash
./gogogo browser start              # Fresh profile
./gogogo browser start --profile   # Copy your profile (cookies, logins)
```

Start Chrome on `:9222` with remote debugging.

### Navigate

```bash
./gogogo browser navigate <url>        # Navigate current tab
./gogogo browser navigate <url> --new  # Open in new tab
```

Navigate current tab or open new tab.

### Evaluate JavaScript

```bash
./gogogo browser eval 'document.title'
./gogogo browser eval 'document.querySelectorAll("a").length'
```

Execute JavaScript in active tab (async context).

### Screenshot

```bash
./gogogo browser screenshot
```

Screenshot current viewport, returns temp file path.

### Pick Elements

```bash
./gogogo browser pick "Click the submit button"
```

Interactive element picker. Click to select, Cmd/Ctrl+Click for multi-select, Enter to finish.

### Cookies

```bash
./gogogo browser cookies
```

Get cookies from current page (as JSON).

## Examples

### Start a scraping session

```bash
# 1. Start Chrome with your profile (to be logged in)
./gogogo browser start --profile

# 2. Navigate to a site
./gogogo browser navigate https://example.com

# 3. Take a screenshot
./gogogo browser screenshot

# 4. Get page title
./gogogo browser eval 'document.title'

# 5. Get all links
./gogogo browser eval 'Array.from(document.querySelectorAll("a")).map(a => a.href)'

# 6. Get cookies for authenticated requests
./gogogo browser cookies
```

### Interactive scraping with element picking

```bash
# Start Chrome
./gogogo browser start

# Navigate to target site
./gogogo browser navigate https://example.com

# Pick specific elements (runs interactive picker)
./gogogo browser pick "Click the main content container"

# Then use the returned selectors to scrape
./gogogo browser eval 'document.querySelector("div.main-content").innerText'
```

## Agent Integration

These tools can be used by Claude Code or other agents. The agent should:

1. Read this README when it needs to interact with a browser
2. Use Bash to invoke the commands
3. Parse the output (file paths, JSON, etc.)

All scripts are composable - output from one command can be used as input to another.
