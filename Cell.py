import numpy as np
from hashlib import sha256
# A faster hashing algorithm
from xxhash import xxh64

class Cell:
    def __init__(self):
        """
        Represents a cell of an IBLTWithEGH
        """
        # Represents the sum (xor of source symbols)
        self.sum = 0
        # Represents the checksum (xor of hashes of source symbols)
        self.checksum = 0
        # Represents the counter array - how many soruce symbols
        # are mapped to the cell.
        self.counter = 0
        # Cache for storing precomputed hashes
        self.hash_cache = {}

    def add(self, symbol: int) -> None:
        """
        Add source symbol to the cell.
        """
        if symbol not in self.hash_cache:
            # Cache the hash if not already cached
            # self.hash_cache[symbol] = sha256(bytes(symbol)).digest()
            self.hash_cache[symbol] = xxh64(bytes(symbol)).digest()

        if self.counter == 0:
            self.sum = symbol
            # digest - get the byte representations of the hash.
            self.checksum = self.hash_cache[symbol]
        else:
            self.sum ^= symbol
            # Perform XOR operation between the hash digests
            xor_result = np.frombuffer(self.checksum, dtype=np.uint64) ^ np.frombuffer(self.hash_cache[symbol], dtype=np.uint64)
            self.checksum = xor_result.tobytes()

        self.counter += 1

    def add_multiple(self, symbols: list[int]) -> None:
        """
        Add multiple source symbols to the cell.
        """
        if len(symbols) == 0:
            return 
        
        self.sum ^= np.bitwise_xor.reduce(symbols)
        self.counter += len(symbols)

        for i, symbol in enumerate(symbols):
            if symbol not in self.hash_cache:
                # Cache the hash if not already cached
                # self.hash_cache[symbol] = sha256(bytes(symbol)).digest()
                self.hash_cache[symbol] = xxh64(bytes(symbol)).digest()

        checksum_arrays = [np.frombuffer(self.hash_cache[symbol], dtype=np.uint64) for symbol in symbols]
        self.checksum = np.bitwise_xor.reduce(checksum_arrays).tobytes()

    def remove(self, symbol: int) -> None:
        """
        Remove source symbol from the cell.
        """
        if symbol not in self.hash_cache:
            # Cache the hash if not already cached.
            # digest - get the byte representations of the hash.
            # self.hash_cache[symbol] = sha256(bytes(symbol)).digest()
            self.hash_cache[symbol] = xxh64(bytes(symbol)).digest()

        self.sum ^= symbol

        # Perform XOR operation between the hash digests        
        xor_result = np.frombuffer(self.checksum, dtype=np.uint64) ^ np.frombuffer(self.hash_cache[symbol], dtype=np.uint64)
        self.checksum = xor_result.tobytes()

        if self.counter > 0:
            self.counter -= 1
        else:
            self.counter += 1

    def is_pure_cell(self) -> bool:
        """
        Check if the cell is pure - containing one element.
        """
        if self.sum not in self.hash_cache:
            # self.hash_cache[self.sum] = sha256(bytes(self.sum)).digest()
            self.hash_cache[self.sum] = xxh64(bytes(self.sum)).digest()

        return (self.counter == 1 or self.counter == -1) and (self.checksum == self.hash_cache[self.sum])


    def is_empty_cell(self) -> bool:
        """
        Check if the cell is empty - containing no elements.
        """
        return self.counter == 0 and self.sum == 0
