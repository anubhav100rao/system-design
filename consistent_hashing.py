import matplotlib.pyplot as plt
import hashlib
import bisect
import pandas as pd


class ConsistentHashRing:
    def __init__(self, nodes=None, replicas=100):
        self.replicas = replicas
        self.ring = {}  # map <database_id -> hash_value>
        self.sorted_hashes = []
        if nodes:
            for node in nodes:
                self.add_node(node)

    def _hash(self, key):
        h = hashlib.md5(key.encode('utf-8')).hexdigest()
        return int(h, 16)

    def add_node(self, node):
        for i in range(self.replicas):
            replica_key = f"{node}:{i}"
            h = self._hash(replica_key)
            self.ring[h] = node
            bisect.insort(self.sorted_hashes, h)

    def remove_node(self, node):
        for i in range(self.replicas):
            replica_key = f"{node}:{i}"
            h = self._hash(replica_key)
            if h in self.ring:
                del self.ring[h]
            idx = bisect.bisect_left(self.sorted_hashes, h)
            if idx < len(self.sorted_hashes) and self.sorted_hashes[idx] == h:
                self.sorted_hashes.pop(idx)

    def get_node(self, key):
        if not self.ring:
            return None
        h = self._hash(key)
        idx = bisect.bisect(self.sorted_hashes, h)
        if idx == len(self.sorted_hashes):
            idx = 0  # circular ring, connecting first and last element
        return self.ring[self.sorted_hashes[idx]]


nodes_initial = ["A", "B", "C"]
ring = ConsistentHashRing(nodes_initial, replicas=10)

keys = [f"key{i}" for i in range(1, 51)]  # product_id = [1, 2, 3, 4 ... 50]
mapping_initial = {key: ring.get_node(key) for key in keys}
dist_initial = pd.Series(mapping_initial).value_counts().reset_index()
dist_initial.columns = ['Node', 'Initial Count']

ring.add_node("D")  # Adding a new database D
mapping_after_add = {key: ring.get_node(key) for key in keys}
dist_after_add = pd.Series(mapping_after_add).value_counts().reset_index()
dist_after_add.columns = ['Node', 'After Add D Count']

moved_after_add = [
    key for key in keys if mapping_initial[key] != mapping_after_add[key]]

# Remove node B and re-map
ring.remove_node("B")  # Removing the database B
mapping_after_remove = {key: ring.get_node(key) for key in keys}
dist_after_remove = pd.Series(
    mapping_after_remove).value_counts().reset_index()
dist_after_remove.columns = ['Node', 'After Remove B Count']

# Keys that moved after removing B
moved_after_remove = [
    key for key in keys if mapping_after_add[key] != mapping_after_remove[key]]

# Create summary of a few keys
example_keys = ["key1", "key10", "key20", "key30", "key40", "key50"]
summary_table = pd.DataFrame({
    "Key": example_keys,
    "Initial": [mapping_initial[k] for k in example_keys],
    "After Add D": [mapping_after_add[k] for k in example_keys],
    "After Remove B": [mapping_after_remove[k] for k in example_keys],
})

(dist_initial, dist_after_add, moved_after_add,
 dist_after_remove, moved_after_remove, summary_table)

print(dist_initial)
print(dist_after_add)
print(moved_after_add)
print(dist_after_remove)
print(moved_after_remove)
print(summary_table)


# Plotting function

# Combined plot for all distributions
fig, axs = plt.subplots(1, 3, figsize=(15, 4), sharey=True)

# Plot Initial Distribution
axs[0].bar(dist_initial['Node'],
           dist_initial['Initial Count'], color='skyblue')
axs[0].set_title("Initial (A, B, C)")
axs[0].set_xlabel("Node")
axs[0].set_ylabel("Key Count")
axs[0].grid(axis='y', linestyle='--', alpha=0.7)

# Plot After Adding Node D
axs[1].bar(dist_after_add['Node'],
           dist_after_add['After Add D Count'], color='lightgreen')
axs[1].set_title("After Adding Node D")
axs[1].set_xlabel("Node")
axs[1].grid(axis='y', linestyle='--', alpha=0.7)

# Plot After Removing Node B
axs[2].bar(dist_after_remove['Node'],
           dist_after_remove['After Remove B Count'], color='salmon')
axs[2].set_title("After Removing Node B")
axs[2].set_xlabel("Node")
axs[2].grid(axis='y', linestyle='--', alpha=0.7)

plt.suptitle("Consistent Hashing Key Distribution", fontsize=14)
plt.tight_layout(rect=[0, 0, 1, 0.95])
plt.show()
