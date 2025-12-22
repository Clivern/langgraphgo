

path = "rag_falkordb_graph/temp_falkordb.go"
with open(path, "r") as f:
    lines = f.readlines()

# We want to append: , "'", "\\'\'"
replacement_suffix = ", \"'\", \\\\\'\\\'\"")

new_lines = []
for line in lines:
    if "strings.ReplaceAll" in line and "escaped" in line:
        indent = "\t"
        if "escapedID :=" in line and "entity.ID" in line:
            new_lines.append(indent + "escapedID := strings.ReplaceAll(entity.ID" + replacement_suffix + "\n")
        elif "escapedSource :=" in line:
            new_lines.append(indent + "escapedSource := strings.ReplaceAll(rel.Source" + replacement_suffix + "\n")
        elif "escapedTarget :=" in line:
            new_lines.append(indent + "escapedTarget := strings.ReplaceAll(rel.Target" + replacement_suffix + "\n")
        elif "escapedID :=" in line and "rel.ID" in line:
            new_lines.append(indent + "escapedID := strings.ReplaceAll(rel.ID" + replacement_suffix + "\n")
        else:
            new_lines.append(line)
    else:
        new_lines.append(line)

with open(path, "w") as f:
    f.writelines(new_lines)

