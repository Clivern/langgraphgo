
path = "rag_falkordb_graph/temp_falkordb.go"
with open(path, "rb") as f:
    content = f.read()

# Pattern: "ESCAPED_QUOTE"
pattern = b'"ESCAPED_QUOTE"'

# Target: "\\' -> 22 5C 5C 27 22
target = bytes([0x22, 0x5c, 0x5c, 0x27, 0x22])

if pattern not in content:
    print("Pattern not found!")
    exit(1)

new_content = content.replace(pattern, target)

with open(path, "wb") as f:
    f.write(new_content)

