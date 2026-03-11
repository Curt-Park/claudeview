---
title: "Stringutil Package (internal/stringutil)"
type: component
tags: [utilities, internals]
---

# Stringutil Package — `internal/stringutil`

Shared string utility functions used across packages to avoid duplication.

## Files & Functions

| File      | Purpose                                                                                  |
|-----------|------------------------------------------------------------------------------------------|
| `xml.go`  | `ExtractXMLTag(s, tag string) string` — returns trimmed content of the first `<tag>…</tag>` in s, or `""` if not found |

## Why This Exists

Extracted during a DRY pass to eliminate identical `extractXMLTag` / `extractXMLContent` implementations that had independently appeared in [[transcript-package]] (`parser.go`) and [[ui-package]] (`chat_item.go`).

## Related

- [[transcript-package]] — uses `ExtractXMLTag` in `extractTopic` for slash-command XML parsing
- [[ui-package]] — uses `ExtractXMLTag` in `cleanTextPreview` for command message extraction
- [[architecture]] — listed in the internal packages table
