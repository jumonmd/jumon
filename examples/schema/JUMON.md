---
module: jumonmd/jumon/example/schema
---

## Scripts

### main

1. Output as JSON of this text.

> 1984 · George Orwell · 1949
> Brave New World · Aldous Huxley · 1932


## Schemas

### main.output
```
{
  "type": "object",
  "properties": {
    "books": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "title": {
            "type": "string"
          },
          "author": {
            "type": "string"
          },
          "publish_year": {
            "type": "number"
          }
        },
        "required": [
          "title",
          "author",
          "publish_year"
        ]
      }
    }
  },
  "required": ["books"]
}
```