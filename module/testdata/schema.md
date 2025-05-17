---
module: schema
---

## Schemas
### main.input
```
{
    "type": "object",
    "properties": {
        "query": {
            "type": "string"
        }
    },
    "required": [
        "query"
    ]
}
```

### main.output
```
{
    "type": "object",
    "properties": {
        "result": {
            "type": "string"
        }
    },
    "required": [
        "result"
    ]
}
```