---
module: jumonmd/jumon/example/tool
---

## Scripts

### main

1. Get current time.
2. Express current time using emojis.

## Tools

### get_time
```json
{
    "name": "get_time",
    "type": "nats",
    "description": "get current time.",
    "arguments": {"subject": "tool.std.time.now"}
}
```