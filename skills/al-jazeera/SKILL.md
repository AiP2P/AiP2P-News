# Al Jazeera Skill

## Site Overview
- **Site**: Al Jazeera News
- **URL**: https://www.aljazeera.com/news
- **Type**: International news
- **Current status**: Direct fetch works

## Historical Test Cases
| Date | Method | Result | Notes |
|------|----------|------|------|
| 2026-03-09 | requests + BeautifulSoup | Success | 268981 bytes, 10+ items |
| 2026-03-09 | x-reader | Success | Backup path |

## Preferred Fetch Method (Method 1)

### Direct Request + BeautifulSoup Parsing
```python
import requests
from bs4 import BeautifulSoup

headers = {'User-Agent': 'Mozilla/5.0'}
url = 'https://www.aljazeera.com/news'
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

### Fallback Page
```bash
curl -s "https://www.aljazeera.com/" | python3 -c "..."
```

## Third Fetch Method (Method 3)

### x-reader / Jina fallback
```bash
~/Library/Python/3.12/bin/x-reader "https://www.aljazeera.com/news"
```

### Success Criteria
- At least 5 summaries

## Common Failure Reasons

| Cause | Code | Mitigation |
|------|------|----------|
| Layout change | - | Update parsing rules |

## Latest Test Result
- 2026-03-09: Direct fetch succeeded, 268981 bytes
