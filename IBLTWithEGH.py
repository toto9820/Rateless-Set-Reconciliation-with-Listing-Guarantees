import numpy as np
from typing import List, Set, Tuple
from Cell import Cell
from queue import Queue
from functools import reduce
from IBLT import IBLT
from scipy.sparse import csr_matrix, vstack
from sympy import nextprime
from numba import jit

class IBLTWithEGH(IBLT):
    def __init__(self, symbols: List[int], n: int):
        """
        Initializes the Invertible Bloom Lookup Table with
        combinatorial method EGH.

        Parameters:
        - symbols (List[int]): set of source symbols.
        - n (int) - universe size.
        """
        super().__init__(symbols, n)
        # Finite array of primes.
        self.primes = []

    def get_next_prime(self, prev_prime: int):
        """
        Returns the next prime number after the given start value.
        If start is not provided or is less than 2, it starts from 2.

        Parameters:
        - prev_prime (int): The previous prime number.

        Returns:
        - int: The next prime number.
        """
        return nextprime(prev_prime)
    
    def generate_mapping(self, iteration: int) -> None:
        """
        Generates part of the mapping matrix for EGH where the number
        of rows depends on the iteration number. 

        Parameters:
        - iteration (int): The iteration number for trasmit/receive.
        """
        prime = None

        if self.primes == []:
            prime = 2
        else:
            prime = self.get_next_prime(self.primes[-1])

        self.primes.append(prime)

        symbols = np.arange(1, self.n + 1)
        row_indices = symbols % prime
        col_indices = symbols - 1
        data = np.ones(self.n, dtype=int)

        self.partial_mapping_matrix = csr_matrix((data, (row_indices, col_indices)), shape=(prime, self.n))

        if iteration == 1:
            self.mapping_matrix = self.partial_mapping_matrix
        else:
            self.mapping_matrix = vstack([self.mapping_matrix, self.partial_mapping_matrix])
    
