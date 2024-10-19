import numpy as np
from hashlib import sha256
# A faster hashing algorithm
from xxhash import xxh32_intdigest, xxh64_intdigest, xxh3_64_intdigest, xxh64_hexdigest 

class Cell:
    def __init__(self, set_inside_set: bool = True, hash_func='xxh64'):
        """
        Represents a cell of an IBLTWithEGH

        Parameters:
        - set_inside_set (bool) : Specifiying if hash calculation is needed
        due to superset assumption.
        - hash_func (str): Specifying the hash function to use ('xxh32', 'xxh64', or 'sha256').
        """
        # Represents the sum (xor of source symbols)
        self.sum = 0
        # Represents the checksum (xor of hashes of source symbols)
        self.checksum = 0
        # Represents the counter array - how many soruce symbols
        # are mapped to the cell.
        self.counter = 0

        self.set_inside_set = set_inside_set

        # TODO - hash of transactions is in string and not int (they are the symbols) form - 
        # should enable to define options to symbols type - int, str, etc.
        if hash_func == 'xxh32':
            self.hash_func = xxh32_intdigest
        elif hash_func == 'xxh64':
            self.hash_func = xxh64_intdigest
        elif hash_func == 'xxh3_64':
            self.hash_func = xxh3_64_intdigest
        # TODO - not supported - lack of intdigest so xor with ^= not useful - fix!
        # elif hash_func == 'sha256':
        #     self.hash_func = lambda x: sha256(bytes(x)).hexdigest()
        #     self.checksum = self.hash_func(self.sum)
        else:
            raise ValueError("Invalid hash function specified. Choose 'xxh32', 'xxh64', 'xxh3_64 or 'sha256'.")
        
        self.vectorized_hash_func = np.vectorize(self.hash_func, otypes=[np.uint64])

    def add(self, symbol: int) -> None:
        """
        Add source symbol to the cell.
        """
        self.sum ^= symbol
        self.counter += 1

        if (self.set_inside_set == False):
            # Perform XOR operation between the hash digests
            self.checksum ^= self.hash_func(symbol)

    def add_multiple(self, symbols: list[int]) -> None:
        """
        Add multiple source symbols to the cell.
        """
        if len(symbols) == 0:
            return 
        
        self.counter += len(symbols)
        self.sum ^= np.bitwise_xor.reduce(symbols)

        if (self.set_inside_set == False):
            hashes = self.vectorized_hash_func(list(symbols))
            hashes_xor = np.bitwise_xor.reduce(hashes)
            self.checksum ^= int(hashes_xor)

    def remove(self, symbol: int) -> None:
        """
        Remove source symbol from the cell.
        """
        self.sum ^= symbol

        if (self.set_inside_set == False):
            self.checksum ^= self.hash_func(symbol)
        
        if self.counter > 0:
            self.counter -= 1
        else:
            self.counter += 1

    def is_pure_cell(self) -> bool:
        """
        Check if the cell is pure - containing one element.
        """
        if abs(self.counter) != 1 or self.sum == 0:
            return False

        if (self.set_inside_set == False):
            return (self.counter == 1 or self.counter == -1) and (self.checksum == self.hash_func(self.sum))
        else:
            return (self.counter == 1 or self.counter == -1)

    def is_empty_cell(self) -> bool:
        """
        Check if the cell is empty - containing no elements.
        """
        return self.counter == 0 and self.sum == 0