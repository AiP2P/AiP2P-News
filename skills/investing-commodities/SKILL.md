# Investing Commodities Skill

## Site Overview
- **Site**: Investing.com Commodities
- **URL**: https://www.investing.com/commodities/
- **Type**: Commodities news
- **Current status**: Direct fetch works

## Historical Test Cases
| Date | Method | Result | Notes |
|------|----------|------|------|
| 2026-03-09 | requests + BeautifulSoup | Success | 993687 bytes, 15+ items |
| 2026-03-09 | x-reader | Success | More stable when JS is heavy |

## Preferred Fetch Method (Method 1)

### Direct Request + BeautifulSoup Parsing
```python
import requests
from bs4 import BeautifulSoup

headers = {'User-Agent': 'Mozilla/5.0'}
url = 'https://www.investing.com/commodities/'
r = requests.get(url, timeout=15, headers=headers)
soup = BeautifulSoup(r.text, 'html.parser')

for a in soup.find_all('a', href=True):
    txt = ' '.join(a.get_text(' ', strip=True).split())
    if len(txt) > 35 and 'investing.com' in a['href']:
        print(txt[:200], '|', a['href'])
```

### Success Criteria
- At least 15 titles with links

## Secondary Fetch Method (Method 2)

### Fallback Page: Energy And Real-Time Prices
```bash
curl -s "https://www.investing.com/commodities/real-time-prices" | python3 -c "..."
```

## Third Fetch Method (Method 3)

### x-reader / Jina fallback
```bash
~/Library/Python/3.12/bin/x-reader "https://www.investing.com/commodities/"
```

### Success Criteria
- At least 5 summaries

## Common Failure Reasons

| Cause | Code | Mitigation |
|------|------|----------|
| Heavy JS rendering | - | Wait for JS or use x-reader |
| Layout change | - | Update parsing rules |

## Debugging Notes

1. Investing pages use a lot of JS, so prefer x-reader when needed.
2. Filter duplicate links.
3. Check whether login gates were introduced.

## Latest Test Result
- 2026-03-09: Direct fetch succeeded, 993687 bytes
