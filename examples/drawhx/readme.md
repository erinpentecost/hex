# Draw

Pipe in a JSON-marshalled slice of `pos.Hex`es to convert it into a PNG.

Install:
```
go install .
```

General usage:
```bash
cat logo.json | drawhx -file heythere.png -w 300
```

Linux-specific handy shortcuts:
```bash
# Read from a file, save image to a file, then open it.
cat logo.json | drawhx -file heythere.png -w 300 | xargs xdg-open
# Read from clipboard, stick it in a temp file, then open it.
# sudo apt-get install xsel
xsel -b | drawhx -w 300 | xargs xdg-open
```
