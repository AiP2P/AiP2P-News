# FT Markets Skill

## Site Overview
- **Site**: Financial Times Markets
- **URL**: https://www.ft.com/markets
- **Type**: Financial and markets news
- **Current status**: Direct fetch works

## Historical Test Cases
| Date | Method | Result | Notes |
|------|----------|------|------|
| 2026-03-09 | requests + BeautifulSoup | Success | 252412 bytes, 10+ items |
| 2026-03-09 | x-reader | Success | Backup path |

## Preferred Fetch Method (Method 1)

### Direct Request + BeautifulSoup Parsing
```python
import requests
from bs4 import BeautifulSoup

headers = {'User-Agent': 'Mozilla/5.0'}
url = 'https://www.ft.com/markets'
r = requests.get(url, timeout=15, headers=headers)
soup = BeautifulSoup(r.text, 'html.parser')

for a in soup.find_all('a', href=True):
    txt = ' '.join(a.get_text(' ', strip=True).split())
    if len(txt) > 35 and a['href'].startswith('http'):
        print(txt[:200], '|', a['href'])
```

### Success Criteria
- At least 10 titles with links

## Secondary Fetch Method (Method 2)

### Fallback Page: FT World
```bash
curl -s "https://www.ft.com/world" | python3 -c "..."
```

## Third Fetch Method (Method 3)

### x-reader / Jina fallback
```bash
~/Library/Python/3.12/bin/x-reader "https://www.ft.com/markets"
```

### Success Criteria
- At least 5 summaries

## Common Failure Reasons

| Cause | Code | Mitigation |
|------|------|----------|
| Paywall | Partial | Some content requires subscription |
| Layout change | - | Update parsing rules |

## Debugging Notes

1. Some FT content is paywalled, but list pages are often fetchable.
2. x-reader may bypass some restrictions.
3. Check whether CSS class names changed.

## Latest Test Result
- 2026-03-09: Direct fetch succeeded, 252412 bytes
