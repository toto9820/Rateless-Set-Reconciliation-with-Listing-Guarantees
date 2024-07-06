import numpy as np
from hashlib import sha256
# A faster hashing algorithm
from xxhash import xxh64, xxh32

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
        self.sum ^= symbol
        self.counter += 1

        if symbol not in self.hash_cache:
            # Cache the hash if not already cached
            # self.hash_cache[symbol] = sha256(bytes(symbol)).digest()
            # self.hash_cache[symbol] = xxh64(symbol).intdigest()
            self.hash_cache[symbol] = xxh32(symbol).intdigest()

        # Perform XOR operation between the hash digests
        self.checksum ^= self.hash_cache[symbol]


    def add_multiple(self, symbols: list[int]) -> None:
        """
        Add multiple source symbols to the cell.
        """
        if len(symbols) == 0:
            return 
        
        symbols_with_sum = np.append(self.sum, symbols)
        self.sum = np.bitwise_xor.reduce(symbols_with_sum)
        self.counter += len(symbols)
        
        for symbol in symbols:
            if symbol not in self.hash_cache:
                # Cache the hash if not already cached
                # self.hash_cache[symbol] = sha256(bytes(symbol)).digest()
                # self.hash_cache[symbol] = xxh64(symbol).intdigest()
                self.hash_cache[symbol] = xxh32(symbol).intdigest()

                
        # For xxh64
        # checksum_values = np.array([self.hash_cache[symbol] for symbol in symbols], dtype=np.uint64)

        # For xxh32
        checksum_values = np.array([self.hash_cache[symbol] for symbol in symbols], dtype=np.uint32)
        
        checksum_values = np.append(checksum_values, self.checksum)
        self.checksum = np.bitwise_xor.reduce(checksum_values)

    def remove(self, symbol: int) -> None:
        """
        Remove source symbol from the cell.
        """
        if symbol not in self.hash_cache:
            # Cache the hash if not already cached.
            # digest - get the byte representations of the hash.
            # self.hash_cache[symbol] = sha256(bytes(symbol)).digest()
            # self.hash_cache[symbol] = xxh64(symbol).intdigest()
            self.hash_cache[symbol] = xxh32(symbol).intdigest()

        self.sum ^= symbol

        self.checksum ^= self.hash_cache[symbol]
        
        if self.counter > 0:
            self.counter -= 1
        else:
            self.counter += 1

    def is_pure_cell(self) -> bool:
        """
        Check if the cell is pure - containing one element.
        """
        if np.abs(self.counter) != 1 or self.sum == 0:
            return False

        if self.sum not in self.hash_cache:
            # self.hash_cache[self.sum] = sha256(bytes(self.sum)).digest()
            # self.hash_cache[self.sum] = xxh64(self.sum).intdigest()
            self.hash_cache[self.sum] = xxh32(self.sum).intdigest()

        return (self.counter == 1 or self.counter == -1) and (self.checksum == self.hash_cache[self.sum])


    def is_empty_cell(self) -> bool:
        """
        Check if the cell is empty - containing no elements.
        """
        return self.counter == 0 and self.sum == 0
