# Draw

Pipe in a JSON-marshalled slice of `pos.Hex`es to convert it into a PNG.

Usage:

```bash
cat logo.json | go run main.go -file heythere.png -w 300 | xargs xdg-open
```
