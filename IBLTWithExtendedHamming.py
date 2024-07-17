import math
import numpy as np
from typing import List, Set
from IBLT import IBLT
from itertools import combinations, product
from scipy.sparse import csr_matrix, vstack

class IBLTWithExtendedHamming(IBLT):
    def __init__(self, symbols: List[int], n: int):
        """
        Initializes the Invertible Bloom Lookup Table with
        binary covering array method.

        Parameters:
        - symbols (List[int]): set of source symbols.
        - n (int) - universe size.
        """
        super().__init__(symbols, n)

    def generate_mapping(self, iteration: int) -> None:
        """
        Generates part of the mapping matrix for Extended Hamming where the number
        of rows depends on the iteration number. 
        The maximum symmetric difference size is 3.

        Parameters:
        - iteration (int): The iteration number for trasmit/receive.
        """
        partial_mapping_matrix = []

        if iteration == 1:
            # First row (all 1s)
            partial_mapping_matrix.append([1] * self.n)

            self.partial_mapping_matrix = csr_matrix(partial_mapping_matrix)
            self.mapping_matrix = self.partial_mapping_matrix
            return

        period = 2 ** (iteration - 2)

        # Create the two alternating blocks
        block1 = [0] * period + [1] * period
        block2 = [1] * period + [0] * period

        num_blocks = self.n // len(block1) + 1
        
        partial_mapping_matrix = [
        (block1 * num_blocks)[:self.n],
        (block2 * num_blocks)[:self.n]
        ]

        # Trim to n elements
        if partial_mapping_matrix != []:
            self.partial_mapping_matrix = csr_matrix(partial_mapping_matrix)
            self.mapping_matrix = csr_matrix(vstack([self.mapping_matrix, self.partial_mapping_matrix]))
        
        