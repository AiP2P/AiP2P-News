# Multi-Source International News Fetching For The Last 24 Hours

## Overview

This skill bundle is for fetching international news from the last 24 hours with multi-source collection, automatic fallback, and lightweight parsing guidance.

## Overall Workflow

```
1. Read the site skill first, for example `cnbc-markets/SKILL.md`.
2. Check the historical test cases and choose the most recently successful method.
3. Try methods in order: Method 1 -> Method 2 -> Method 3.
4. Parse HTML or JSON and extract title + link pairs.
5. Deduplicate and merge similar events.
6. Map stories into broad commodity or macro groups if needed.
```

## Preferred News Sources (By Priority)

### Tier 1: Most Reliable (Direct Fetch Usually Works)
- CNBC Markets: https://www.cnbc.com/markets/
- CNBC World: https://www.cnbc.com/world/?region=world
- BBC News: https://www.bbc.com/news
- FT Markets: https://www.ft.com/markets
- Oilprice: https://oilprice.com/
- Al Jazeera: https://www.aljazeera.com/news
- AP News: https://apnews.com/hub/world-news
- TechCrunch: https://techcrunch.com/

### Tier 2: Usually Needs x-reader Or Jina
- Bloomberg: https://www.bloomberg.com/markets (x-reader is commonly needed to bypass 403)

### Tier 3: Currently Unreliable Or Not Available
- Reuters (451 legal restriction in some environments)
- WSJ (401/451)
- The Economist (403)
- Some RSS sources

## Fallback Order

```
Method 1: direct requests + BeautifulSoup
    ↓ (if 403/401)
Method 2: change user agent / fallback page / lighter parsing
    ↓ (if still failing)
Method 3: x-reader / Jina / Agent-Reach
```

## Tools And Methods Used In Practice

### 1. Python requests + BeautifulSoup
- Used for direct page fetches
- Requires: `pip install requests beautifulsoup4`

### 2. x-reader (GitHub Project)
- Project: https://github.com/runesleo/x-reader
- Install: `pip install --user git+https://github.com/runesleo/x-reader.git`
- Used to bypass anti-bot protection for sources like Bloomberg or The Economist

### 3. Agent-Reach (GitHub Project)
- Project: https://github.com/Panniantong/Agent-Reach
- Install: through a Python environment
- Provides Jina Reader, RSS, Twitter, and other integrations

## Installed Dependencies

```bash
# Core dependencies
pip install requests beautifulsoup4 feedparser

# x-reader (Jina-based URL reader)
pip install --user git+https://github.com/runesleo/x-reader.git

# x-reader Telegram extension
pip install --user 'x-reader[telegram]@ git+https://github.com/runesleo/x-reader.git'
```

## Moving To Another Machine

### Minimum Files To Copy

1. **`skills/` directory**
   - All site skills such as `bbc-news/`, `cnbc-markets/`, and `oilprice/`
   - Each skill includes `SKILL.md` with fetch methods and tested cases

2. **`scripts/` directory**
   - Fetch scripts, if any

3. **`docs/` directory**
   - Installation notes
   - Dependency lists

### Required Installation

```bash
# Python environment (3.10+)
pip install requests beautifulsoup4 feedparser

# x-reader
pip install --user git+https://github.com/runesleo/x-reader.git
```

## Usage Example

```python
# Method 1: direct fetch (Tier 1 source)
import requests
from bs4 import BeautifulSoup

headers = {'User-Agent': 'Mozilla/5.0'}
url = 'https://www.cnbc.com/markets/'
r = requests.get(url, timeout=15, headers=headers)
soup = BeautifulSoup(r.text, 'html.parser')
# Extract title + link pairs here

# Method 2: x-reader (Tier 2 source)
~/Library/Python/3.12/bin/x-reader "https://www.bloomberg.com/markets"
```

## Included Site Skills

| Site | Status | Primary Method |
|------|------|----------|
| bbc-news | ✅ | requests + BS |
| cnbc-markets | ✅ | requests + BS |
| cnbc-world | ✅ | requests + BS |
| oilprice | ✅ | requests + BS |
| investing-commodities | ✅ | requests + BS |
| ft-markets | ✅ | requests + BS |
| ap-world | ✅ | requests + BS |
| al-jazeera | ✅ | requests + BS |
| techcrunch | ✅ | requests + BS |
| bloomberg | ⚠️ | x-reader bypass |

## Adding More Sites

To add a new site:
1. Create a new directory under `skills/`.
2. Add a `SKILL.md` that includes:
   - site overview
   - historical test cases
   - Method 1 / 2 / 3
   - success criteria
   - common failure reasons
   - debugging notes
3. After testing, update the historical cases table.
