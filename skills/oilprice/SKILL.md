# Oilprice Skill

## Site Overview
- **Site**: Oilprice.com
- **URL**: https://oilprice.com/
- **Type**: Energy news
- **Current status**: Direct fetch works

## Historical Test Cases
| Date | Method | Result | Notes |
|------|----------|------|------|
| 2026-03-09 | requests + BeautifulSoup | Success | 254462 bytes, 10+ items |
| 2026-03-09 | x-reader | Success | Backup path |

## Preferred Fetch Method (Method 1)

### Direct Request + BeautifulSoup Parsing
```python
import requests
from bs4 import BeautifulSoup

headers = {'User-Agent': 'Mozilla/5.0'}
url = 'https://oilprice.com/'
r = requests.get(url, timeout=15, headers=headers)
soup = BeautifulSoup(r.text, 'html.parser')

for a in soup.find_all('a', href=True):
    txt = ' '.join(a.get_text(' ', strip=True).split())
    if len(txt) > 35 and 'oilprice.com' in a['href']:
        print(txt[:200], '|', a['href'])
```

### Success Criteria
- At least 10 titles with links

## Secondary Fetch Method (Method 2)

### Fallback Energy Page
```bash
curl -s "https://oilprice.com/oil-price-charts/" | python3 -c "..."
```

## Third Fetch Method (Method 3)

### x-reader / Jina fallback
```bash
~/Library/Python/3.12/bin/x-reader "https://oilprice.com/"
```

### Success Criteria
- At least 5 summaries

## Common Failure Reasons

| Cause | Code | Mitigation |
|------|------|----------|
| Heavy ads or JS | - | Filter non-news links |
| Layout change | - | Update parsing rules |

## Debugging Notes

1. Filter non-news links such as `oil-price-charts`.
2. Check whether the page structure changed.
3. Use x-reader if ad-heavy pages degrade the result.

## Latest Test Result
- 2026-03-09: Direct fetch succeeded, 254462 bytes
