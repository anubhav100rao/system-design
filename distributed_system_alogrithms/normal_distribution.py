import hashlib
import pandas as pd
import matplotlib.pyplot as plt


def hash_int(key: str) -> int:
    """Return an integer hash for a string key (MD5 -> int)."""
    h = hashlib.md5(key.encode('utf-8')).hexdigest()
    return int(h, 16)


def get_node_mod(key: str, nodes: list[str]) -> str:
    """Assign key to a node by simple mod of hash over number of nodes."""
    if not nodes:
        return None
    return nodes[hash_int(key) % len(nodes)]


# Initial nodes and keys
nodes_initial = ["A", "B", "C"]
keys = [f"key{i}" for i in range(1, 51)]  # 50 keys



# Mapping with simple modulo assignment
mapping_initial = {k: get_node_mod(k, nodes_initial) for k in keys}
dist_initial = pd.Series(mapping_initial).value_counts().reset_index()
dist_initial.columns = ['Node', 'Initial Count']

# Add a new node D (append to node list)
nodes_after_add = nodes_initial + ["D"]
mapping_after_add = {k: get_node_mod(k, nodes_after_add) for k in keys}
dist_after_add = pd.Series(mapping_after_add).value_counts().reset_index()
dist_after_add.columns = ['Node', 'After Add D Count']

# Keys that moved after adding D
moved_after_add = [
    k for k in keys if mapping_initial[k] != mapping_after_add[k]]

# Remove node B from the nodes set (simulate removal after add)
nodes_after_remove = [n for n in nodes_after_add if n != "B"]
mapping_after_remove = {k: get_node_mod(k, nodes_after_remove) for k in keys}
dist_after_remove = pd.Series(
    mapping_after_remove).value_counts().reset_index()
dist_after_remove.columns = ['Node', 'After Remove B Count']

# Keys that moved after removing B (compare add -> remove)
moved_after_remove = [
    k for k in keys if mapping_after_add[k] != mapping_after_remove[k]]

# Example summary (a few keys)
example_keys = ["key1", "key10", "key20", "key30", "key40", "key50"]
summary_table = pd.DataFrame({
    "Key": example_keys,
    "Initial": [mapping_initial[k] for k in example_keys],
    "After Add D": [mapping_after_add[k] for k in example_keys],
    "After Remove B": [mapping_after_remove[k] for k in example_keys],
})

# Print summaries
print("Initial distribution:\n", dist_initial, "\n")
print("After adding D distribution:\n", dist_after_add, "\n")
print("Keys moved after adding D (count):", len(moved_after_add))
print(moved_after_add, "\n")
print("After removing B distribution:\n", dist_after_remove, "\n")
print("Keys moved after removing B (count):", len(moved_after_remove))
print(moved_after_remove, "\n")
print("Example key mapping changes:\n", summary_table, "\n")

# Plotting (same layout as your original)
fig, axs = plt.subplots(1, 3, figsize=(15, 4), sharey=True)

# Ensure consistent order of node labels on x-axis for nicer comparison:
all_nodes_initial = sorted(dist_initial['Node'].tolist())
all_nodes_add = sorted(dist_after_add['Node'].tolist())
all_nodes_remove = sorted(dist_after_remove['Node'].tolist())

# Plot Initial
axs[0].bar(dist_initial['Node'], dist_initial['Initial Count'])
axs[0].set_title("Initial (A, B, C)")
axs[0].set_xlabel("Node")
axs[0].set_ylabel("Key Count")
axs[0].grid(axis='y', linestyle='--', alpha=0.7)

# Plot After Add D
axs[1].bar(dist_after_add['Node'], dist_after_add['After Add D Count'])
axs[1].set_title("After Adding Node D")
axs[1].set_xlabel("Node")
axs[1].grid(axis='y', linestyle='--', alpha=0.7)

# Plot After Remove B
axs[2].bar(dist_after_remove['Node'],
           dist_after_remove['After Remove B Count'])
axs[2].set_title("After Removing Node B")
axs[2].set_xlabel("Node")
axs[2].grid(axis='y', linestyle='--', alpha=0.7)

plt.suptitle(
    "Modulo Hashing Key Distribution (no consistent hashing)", fontsize=14)
plt.tight_layout(rect=[0, 0, 1, 0.95])
plt.show()
