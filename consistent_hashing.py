import bisect
import hashlib


class ConsistentHashing:
    def __init__(self, num_replicas=3):
        """
        Initialize consistent hashing.
        :param num_replicas: Number of virtual replicas per node to ensure even key distribution.
        """
        self.num_replicas = num_replicas
        self.ring = {}  # Mapping of hash value to node
        self.sorted_hashes = []  # Sorted list of hash values (ring positions)

    def _hash(self, key):
        """
        Generate a hash for a given key using SHA-256.
        """
        return int(hashlib.sha256(key.encode('utf-8')).hexdigest(), 16)

    def add_node(self, node):
        """
        Add a new node to the hash ring.
        :param node: Node identifier (e.g., IP address or node name).
        """
        for i in range(self.num_replicas):
            virtual_node = f"{node}:{i}"  # Virtual node name
            hash_val = self._hash(virtual_node)
            self.ring[hash_val] = node
            bisect.insort(self.sorted_hashes, hash_val)

    def remove_node(self, node):
        """
        Remove a node and its replicas from the hash ring.
        :param node: Node identifier.
        """
        for i in range(self.num_replicas):
            virtual_node = f"{node}:{i}"
            hash_val = self._hash(virtual_node)
            self.ring.pop(hash_val, None)
            self.sorted_hashes.remove(hash_val)

    def get_node(self, key):
        """
        Get the node responsible for a given key.
        :param key: The key to look up.
        :return: The node responsible for the key.
        """
        if not self.ring:
            return None
        hash_val = self._hash(key)
        idx = bisect.bisect(self.sorted_hashes, hash_val)
        if idx == len(self.sorted_hashes):
            idx = 0  # Wrap around to the first node
        return self.ring[self.sorted_hashes[idx]]

    def display_ring(self):
        """
        Display the current hash ring for debugging purposes.
        """
        for hash_val in self.sorted_hashes:
            print(f"Hash: {hash_val}, Node: {self.ring[hash_val]}")


# Create a consistent hashing instance
ch = ConsistentHashing(num_replicas=3)

# Add nodes to the hash ring
ch.add_node("NodeA")
ch.add_node("NodeB")
ch.add_node("NodeC")

# Display the hash ring
print("Initial hash ring:")
ch.display_ring()

# Lookup which node a key belongs to
keys = ["key1", "key2", "key3", "key4", "key5"]
for key in keys:
    print(f"Key '{key}' is assigned to node: {ch.get_node(key)}")

# Remove a node
ch.remove_node("NodeB")
print("\nHash ring after removing NodeB:")
ch.display_ring()

# Re-check which node keys belong to
for key in keys:
    print(f"Key '{key}' is assigned to node: {ch.get_node(key)}")

