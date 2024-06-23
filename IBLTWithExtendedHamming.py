import math
from typing import List, Set
from IBLT import IBLT
from itertools import combinations, product

class IBLTWithExtendedHamming(IBLT):
    def __init__(self, symbols: Set[int], n: int):
        """
        Initializes the Invertible Bloom Lookup Table with
        binary covering array method.

        Parameters:
        - symbols (Set[int]): set of source symbols.
        - n (int) - universe size.
        """
        super().__init__(symbols, n)
        self.stopping_condition_exists = False

    def generate_mapping(self, iteration: int) -> None:
        """
        Generates part of the mapping matrix for Extended Hamming where the number
        of rows depends on the iteration number. 
        The maximum symmetric difference size is 3.

        Parameters:
        - iteration (int): The iteration number for trasmit/receive.
        """
        self.partial_mapping_matrix = []
        row = []

        block_size = 2 ** (iteration - 1)

        for _ in range(self.n // block_size):
            row.extend([0] * block_size + [1] * block_size)
        
        # Trim to n elements
        if row != []:
            self.partial_mapping_matrix.append(row[:self.n]) 
            self.mapping_matrix.append(row[:self.n]) 
        
        if iteration == 1:
            # Last row (all 1s)
            self.partial_mapping_matrix.append([1] * self.n)
            self.mapping_matrix.append([1] * self.n)