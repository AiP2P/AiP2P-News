# Bloomberg Skill

## Site Overview
- **Site**: Bloomberg Markets
- **URL**: https://www.bloomberg.com/markets
- **Type**: Financial and markets news
- **Current status**: Requires x-reader or Jina bypass

## Historical Test Cases
| Date | Method | Result | Notes |
|------|----------|------|------|
| 2026-03-09 | direct requests | Failed | 403, anti-bot protection |
| 2026-03-09 | x-reader / Jina | Success | 5+ summaries |
| 2026-03-09 | Agent-Reach | Untested | Alternative path |

## Preferred Fetch Method (Method 1)

### x-reader / Jina (Recommended)
```bash
~/Library/Python/3.12/bin/x-reader "https://www.bloomberg.com/markets"
```

### Success Criteria
- At least 5 summaries or headlines

## Secondary Fetch Method (Method 2)

### Direct x-reader Request
```bash
~/Library/Python/3.12/bin/x-reader "https://www.bloomberg.com"
```

## Third Fetch Method (Method 3)

### Agent-Reach (Alternative)
```bash
# Agent-Reach must be configured first
agent-reach fetch "https://www.bloomberg.com/markets"
```

## Why Direct Requests Fail

| Cause | Code | Mitigation |
|------|------|----------|
| Anti-bot blocking | 403 | Use x-reader or Jina |
| Paywall | - | Some content requires subscription |

## Common Failure Reasons

| Cause | Code | Mitigation |
|------|------|----------|
| 403 anti-bot response | 403 | Route through x-reader |
| Page changes | - | Update the fetch approach |

## Debugging Notes

1. Bloomberg aggressively blocks direct `curl` and `requests`.
2. Prefer the Jina fallback inside x-reader.
3. If x-reader also fails, check whether authentication or geography is involved.

## Latest Test Result
- 2026-03-09: Direct requests returned 403, x-reader succeeded
