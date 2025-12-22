path = "rag_falkordb_graph/temp_falkordb.go"
with open(path, "rb") as f:
    content = f.read()

# Pattern found: 22 27 5c 27 27 22 -> "'\''"
pattern = bytes([0x22, 0x27, 0x5c, 0x27, 0x27, 0x22])

# Target: 22 5c 5c 27 22 -> "\\"
target = bytes([0x22, 0x5c, 0x5c, 0x27, 0x22])

new_content = content.replace(pattern, target)

with open(path, "wb") as f:
    f.write(new_content)