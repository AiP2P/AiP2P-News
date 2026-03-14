# CNBC World Skill

## Site Overview
- **Site**: CNBC World
- **URL**: https://www.cnbc.com/world/?region=world
- **Type**: International news
- **Current status**: Direct fetch works

## Historical Test Cases
| Date | Method | Result | Notes |
|------|----------|------|------|
| 2026-03-09 | requests + BeautifulSoup | Success | 1627577 bytes, 10+ items |
| 2026-03-09 | x-reader | Success | Backup path |

## Preferred Fetch Method (Method 1)

### Direct Request + BeautifulSoup Parsing
```python
import requests
from bs4 import BeautifulSoup

headers = {'User-Agent': 'Mozilla/5.0'}
url = 'https://www.cnbc.com/world/?region=world'
r = requests.get(url, timeout=15, headers=headers)
soup = BeautifulSoup(r.text, 'html.parser')

for a in soup.find_all('a', href=True):
    txt = ' '.join(a.get_text(' ', strip=True).split())
    if len(txt) > 35 and a['href'].startswith('http'):
        print(txt[:200], '|', a['href'])
```

### Success Criteria
- At least 15 titles with links

## Secondary Fetch Method (Method 2)

### Fallback Page: CNBC Markets
```bash
curl -s "https://www.cnbc.com/markets/" | python3 -c "..."
```

## Third Fetch Method (Method 3)

### x-reader / Jina fallback
```bash
~/Library/Python/3.12/bin/x-reader "https://www.cnbc.com/world/?region=world"
```

### Success Criteria
- At least 5 summaries

## Common Failure Reasons

| Cause | Code | Mitigation |
|------|------|----------|
| Anti-bot blocking | 403 | Change the user agent or use x-reader |
| Parameter issue | - | Verify the `region` parameter |

## Latest Test Result
- 2026-03-09: Direct fetch succeeded, 1627577 bytes
