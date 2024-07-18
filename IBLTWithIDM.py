import math
import numpy as np
from typing import List, Set
from IBLT import IBLT
from itertools import combinations, product
from scipy.sparse import csr_matrix, vstack

class IBLTWithIDM(IBLT):
    def __init__(self, symbols: List[int], n: int):
        """
        Initializes the Invertible Bloom Lookup Table with
        iterative decodability matrix method.

        Parameters:
        - symbols (List[int]): set of source symbols.
        - n (int) - universe size.
        """
        super().__init__(symbols, n)
        self.k = int(np.ceil(np.log2(n))) 

    def xor_rows(self, matrix, row_indices):
        return np.bitwise_xor.reduce(matrix[row_indices, :])

    def generate_mapping(self, iteration: int) -> None:
        """
        Generates part of the mapping matrix for Iterative Decodability where the number
        of rows depends on the iteration number. 

        Parameters:
        - iteration (int): The iteration number for trasmit/receive.
        """
        partial_mapping_matrix = []

        if iteration == 1:
            partial_mapping_matrix = np.array([list(map(int, format(i, f'0{self.k}b'))) for i in range(1, self.n+1)]).T

            self.partial_mapping_matrix = csr_matrix(partial_mapping_matrix)
            self.mapping_matrix = self.partial_mapping_matrix
            return

        elif iteration == 2:
            identity = np.eye(self.k, dtype=int)
            partial_mapping_matrix = np.tile(identity, (1, (self.n // self.k) + 1))[:self.k, :self.n]

            self.partial_mapping_matrix = csr_matrix(partial_mapping_matrix)
            self.mapping_matrix = self.partial_mapping_matrix
            return
        
        else:
            for combo in combinations(range(2*self.k), iteration):
                new_row = self.xor_rows(self.mapping_matrix.toarray(), combo)
                partial_mapping_matrix.append(new_row)

            self.partial_mapping_matrix = csr_matrix(partial_mapping_matrix)
            self.mapping_matrix = self.partial_mapping_matrix

        
        
        
        