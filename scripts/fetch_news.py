#!/usr/bin/env python3
"""
PC75 multi-source news fetcher
Usage: python3 fetch_news.py
"""

import requests, json
from bs4 import BeautifulSoup
from datetime import datetime

headers = {'User-Agent': 'Mozilla/5.0'}

SITES = [
    ('CNBC Markets', 'https://www.cnbc.com/markets/'),
    ('BBC News', 'https://www.bbbc.com/news'),
    ('Oilprice', 'https://oilprice.com/'),
    # Add more sites here.
]

def fetch_site(name, url):
    try:
        r = requests.get(url, timeout=15, headers=headers)
        if r.status_code != 200:
            return None
        soup = BeautifulSoup(r.text, 'html.parser')
        results = []
        for a in soup.find_all('a', href=True):
            txt = ' '.join(a.get_text(' ', strip=True).split())
            if len(txt) > 35:
                results.append({'title': txt[:200], 'url': a['href']})
        return results[:10]
    except Exception as e:
        print(f"Error fetching {name}: {e}")
        return None

def main():
    all_news = []
    for name, url in SITES:
        print(f"Fetching {name}...")
        items = fetch_site(name, url)
        if items:
            all_news.extend([{'source': name, **i} for i in items])
    
    # Save results.
    date = datetime.now().strftime('%Y-%m-%d')
    with open(f'news_{date}.json', 'w') as f:
        json.dump(all_news, f, ensure_ascii=False, indent=2)
    print(f"Saved {len(all_news)} items")

if __name__ == '__main__':
    main()
