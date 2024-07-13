import numpy as np
from typing import List, Set, Tuple
from Cell import Cell
from queue import Queue
from functools import reduce
from IBLT import IBLT
from scipy.sparse import csr_matrix, vstack
from galois import GF, primitive_poly

class IBLTWithBCH(IBLT):
    def __init__(self, symbols: Set[int], n: int):
        """
        Initializes the Invertible Bloom Lookup Table with
        coding method BCH.

        Parameters:
        - symbols (Set[int]): set of source symbols.
        - n (int) - universe size.
        """
        super().__init__(symbols, n)
        # Finite array of primes.
        self.primes = []
        self.m = int(np.ceil(np.log2(self.n + 1)))
        self.gf2m = GF(2**self.m)
        self.alpha = self.gf2m.primitive_element
    
    def gf2m_pow(self, x, power):
        """Compute x^power in GF(2^m)."""
        result = x
        for _ in range(power - 1):
            result = (result << 1) & ((1 << self.m) - 1)
            if result & (1 << (self.m-1)):
                result ^= int(primitive_poly(2,self.m))  # Primitive polynomial for GF(2^m)
        return result

    def generate_mapping(self, iteration: int) -> None:
        """
        Generate part of the parity-check matrix for BCH code.
        
        Parameters:
        iteration : int
            The iteration number, determining the number of rows to add.
        """

        if iteration >= 2**(self.m - 1):
            raise ValueError(f"No more gurantee for decodablitiy for d = {iteration}")
        
        partial_mapping_matrix = np.zeros((self.m, self.n), dtype=int)
                
        # Set the left most column 
        partial_mapping_matrix[:self.m, 0] = np.array([int(bit) for bit in format(1, f'0{self.m}b')[::-1]])

        # Precompute powers of alpha
        powers = (2 * (iteration - 1) + 1) * np.arange(1, self.n)
        field_elements = np.power(self.alpha, powers)

        binary_matrix = np.array([list(map(int, format(element, f'0{self.m}b')[::-1])) for element in field_elements], dtype=int)
    
        partial_mapping_matrix[:self.m, 1:self.n] = binary_matrix.T

        self.partial_mapping_matrix = csr_matrix(partial_mapping_matrix)
        
        if iteration == 1:
            self.mapping_matrix = self.partial_mapping_matrix
        else:
            self.mapping_matrix = vstack([self.mapping_matrix, self.partial_mapping_matrix])
    
