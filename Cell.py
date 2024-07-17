import numpy as np
import operator
import multiprocessing
from hashlib import sha256
# A faster hashing algorithm
from xxhash import xxh32_intdigest, xxh64_intdigest, xxh3_64_intdigest 
from functools import reduce

class Cell:
    def __init__(self, hash_func='xxh64'):
        """
        Represents a cell of an IBLTWithEGH

        Parameters:
        - hash_func (str): Specifying the hash function to use ('xxh32', 'xxh64', or 'sha256').
        """
        # Represents the sum (xor of source symbols)
        self.sum = 0
        # Represents the checksum (xor of hashes of source symbols)
        self.checksum = 0
        # Represents the counter array - how many soruce symbols
        # are mapped to the cell.
        self.counter = 0

        # TODO - hash of transactions is in string and not int (they are the symbols) form - 
        # should enable to define options to symbols type - int, str, etc.
        if hash_func == 'xxh32':
            self.hash_func = lambda x: xxh32_intdigest(x)
        elif hash_func == 'xxh64':
            self.hash_func = lambda x: xxh64_intdigest(x)
        elif hash_func == 'xxh3_64':
            self.hash_func = lambda x: xxh3_64_intdigest(x)
        # TODO - not supported - lack of intdigest so xor with ^= not useful - fix!
        # elif hash_func == 'sha256':
        #     self.hash_func = lambda x: sha256(bytes(x)).hexdigest()
        #     self.checksum = self.hash_func(self.sum)
        else:
            raise ValueError("Invalid hash function specified. Choose 'xxh32', 'xxh64', 'xxh3_64 or 'sha256'.")

    def add(self, symbol: int) -> None:
        """
        Add source symbol to the cell.
        """
        self.sum ^= symbol
        self.counter += 1

        # Perform XOR operation between the hash digests
        self.checksum ^= self.hash_func(symbol)

    def add_multiple(self, symbols: list[int]) -> None:
        """
        Add multiple source symbols to the cell.
        """
        if len(symbols) == 0:
            return 
        
        self.counter += len(symbols)

        hashes = [self.hash_func(symbol) for symbol in symbols]

        # hashes = []

        # cpu_cores_count = multiprocessing.cpu_count()
        # chunk_size = len(symbols) // cpu_cores_count + 1

        # with multiprocessing.Pool(processes=cpu_cores_count) as pool:
        #     hashes = pool.map(xxh64_intdigest, symbols, chunk_size)

        self.sum ^= reduce(operator.xor, symbols)
        self.checksum ^= reduce(operator.xor, hashes)

    def remove(self, symbol: int) -> None:
        """
        Remove source symbol from the cell.
        """
        self.sum ^= symbol

        self.checksum ^= self.hash_func(symbol)
        
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

        return (self.counter == 1 or self.counter == -1) and (self.checksum == self.hash_func(self.sum))


    def is_empty_cell(self) -> bool:
        """
        Check if the cell is empty - containing no elements.
        """
        return self.counter == 0 and self.sum == 0