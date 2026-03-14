# BBC News Skill

## Site Overview
- **Site**: BBC News World
- **URL**: https://www.bbc.com/news
- **Type**: General international news
- **Current status**: Direct fetch works

## Historical Test Cases
| Date | Method | Result | Notes |
|------|----------|------|------|
| 2026-03-09 | curl + BeautifulSoup | Success | 28 links |
| 2026-03-09 | x-reader | Success | Backup path |

## Preferred Fetch Method (Method 1)

### Direct Request + BeautifulSoup Parsing
```bash
curl -s -A "Mozilla/5.0" "https://www.bbc.com/news" | python3 -c "
import sys, re
from bs4 import BeautifulSoup
soup = BeautifulSoup(sys.stdin.read(), 'html.parser')
for a in soup.find_all('a', href=True):
    txt = ' '.join(a.get_text(' ', strip=True).split())
    if len(txt) > 30 and a['href'].startswith('http'):
        print(txt[:200], '|', a['href'])
"
```

### Success Criteria
- At least 10 titles with links

## Secondary Fetch Method (Method 2)

### Retry With A More Complete User Agent
```bash
curl -s -A "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36" "https://www.bbc.com/news" | python3 -c "..."
```

### Success Criteria
- Same as above

## Third Fetch Method (Method 3)

### x-reader / Jina fallback
```bash
~/Library/Python/3.12/bin/x-reader "https://www.bbc.com/news"
```

### Success Criteria
- At least 5 summaries

## Common Failure Reasons

| Cause | Code | Mitigation |
|------|------|----------|
| Anti-bot blocking | 403 | Change the user agent or use x-reader |
| DNS issue | - | Check network connectivity |
| Layout change | - | Update parsing rules |

## Debugging Notes

1. First verify that `curl` returns `200`.
2. Check whether the HTML structure changed.
3. Use x-reader as the final fallback.

## Latest Test Result
- 2026-03-09: Direct fetch succeeded, 28 links
