---
module: jumonmd/jumon/example/tool
---

## Scripts

### main

1. Get weather from Tokyo.
2. Explain in Japanese.


## Tools

### get_weather
```json
{
    "name": "get_weather",
    "type": "nats",
    "description": "get weather at the city.",
    "version": "0.1.0",
    "input_schema": {
        "type": "object",
        "properties": {
            "city": {
                "type": "string"
            }
        },
        "required": ["city"]
    },
    "arguments": {"subject": "tool.example.get_weather"}
}
```