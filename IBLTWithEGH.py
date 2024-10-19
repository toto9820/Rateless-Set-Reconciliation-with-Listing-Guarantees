import os
import numba as nb
import numpy as np
from typing import List
from IBLT import IBLT
from sympy import primerange

@nb.njit(parallel=True, cache=True)
def fast_generate_mapping(symbols, prime, n):
    partial_mapping_matrix = np.zeros((prime, n), dtype=np.int8)

    for i in nb.prange(len(symbols)):
        row = symbols[i] % prime
        col = symbols[i] - 1
        partial_mapping_matrix[row, col] = 1
      
    return partial_mapping_matrix

@nb.njit(cache=True)
def get_updated_mapping_matrix(mapping_matrix, partial_mapping_matrix):
    mapping_matrix = np.hstack((mapping_matrix, partial_mapping_matrix.T))
    return mapping_matrix

class IBLTWithEGH(IBLT):
    def __init__(self, symbols: List[int], n: int, set_inside_set: bool = True):
        super().__init__(symbols, n, set_inside_set)
        self.primes = [2]
        self.symbols = np.arange(1, self.n + 1)
        # self.mapping_matrix = np.array([], dtype=np.int8)
        # Create mapping every chunk size iterations.
        # Other times, just lookups.
        self.chunk_size = 1000
        self.offsets = []
        self.mapping_matrix_used_rows = 0
        self.mapping_matrix_file = 'egh_mapping_matrix.dat'

        # Check if file exists, open in r+ mode if so, otherwise create it in w+ mode
        # For now number of rows 10^6 until solution to resizing
        # in case offset bigger than n is found.
        if os.path.exists(self.mapping_matrix_file):
            self.mapping_matrix = np.memmap(self.mapping_matrix_file, dtype=np.int8, mode='r+', shape=(10**6, n))
        else:
            self.mapping_matrix = np.memmap(self.mapping_matrix_file, dtype=np.int8, mode='w+', shape=(10**6, n))
         
    def create_mapping_matrix(self):
        primes = primerange(10**6)
    
    
    def generate_mapping(self, iteration: int) -> None:
        if iteration == len(self.primes):
            start = self.primes[-1] + 1
            end = start + self.chunk_size
            new_primes = list(primerange(start, end))
            self.primes.extend(new_primes)
            # self.offsets = np.cumsum(self.primes)

        prime = self.primes[iteration - 1]
        self.partial_mapping_matrix = fast_generate_mapping(self.symbols, prime, self.n)

        if iteration == 1:
            # self.mapping_matrix = [(self.partial_mapping_matrix, 0)]
            self.mapping_matrix[:prime] = self.partial_mapping_matrix
            self.mapping_matrix_used_rows = prime
        else:
            # offset = self.offsets[iteration - 2]
            offset = self.mapping_matrix_used_rows
            # self.mapping_matrix.append((self.partial_mapping_matrix, offset))

            self.mapping_matrix[offset:offset+prime] = self.partial_mapping_matrix
            self.mapping_matrix_used_rows += prime
    
    def get_current_mapping_rows(self):
        """
        Gets number of current rows for mapping matrix that are
        used for listing.
        """
        return self.mapping_matrix_used_rows