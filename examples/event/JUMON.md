---
module: jumonmd/jumon/example/event
---

## Scripts

### main

1. Output as JSON of this text.

> 1984 · George Orwell · 1949
> Brave New World · Aldous Huxley · 1932


## Events

### subscribe event.yakhon.books.candidates
module: jumonmd/examples/yakhon
template:
```
{
  "title": "{{ .title }}",
  "members": ["1046947783246106715"],
  "message": "{{ .title }} -  {{ .author }}\nhttps://www.amazon.co.jp/s?k={{ .isbn }}\n\n{{ .originalTitle }} - {{ .originalAuthor }}\nhttps://www.amazon.co.jp/s?k={{ .originalIsbn }}"
  {{- if .originalIsbn }}
  ,"buttons": [
    {
      "label": "Confirm",
      "style": 1,
      "custom_id": "ok:{{ .isbn }}={{ .originalIsbn }}"
    }
  ]
  {{- end }}
}
```

### publish jumonmd/examples/yakhon
subject: event.yakhon.books.candidates