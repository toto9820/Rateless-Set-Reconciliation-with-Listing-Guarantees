import os
import numpy as np
from typing import List, Set
from IBLT import IBLT

class IBLTWithOLS(IBLT):
    def __init__(self, symbols: List[int], n: int, set_inside_set: bool = True):
        """
        Initializes an Invertible Bloom Lookup Table with
        ortogonal latin squares method.

        Parameters:
        - symbols (List[int]): set of source symbols.
        - n (int) - universe size.
        - set_inside_set (bool) - flag indicating whether a superset assumption holds, i.e. one participant's set
        includes the other.
        """
        super().__init__(symbols, n)
        self.mapping_matrix_file = 'ols_mapping_matrix.dat'
        self.mapping_matrix_used_rows = 0
        # Calculate the size of the Latin square
        self.s = int(np.ceil(np.sqrt(n))) 

        # Check if file exists, open in r+ mode if so, otherwise create it in w+ mode
        # For now number of rows 10^6 until solution to resizing
        # in case offset bigger than n is found.
        if os.path.exists(self.mapping_matrix_file):
            self.mapping_matrix = np.memmap(self.mapping_matrix_file, dtype=np.int8, mode='r+', shape=(10**6, n))
        else:
            self.mapping_matrix = np.memmap(self.mapping_matrix_file, dtype=np.int8, mode='w+', shape=(10**6, n))

    def generate_mapping(self, iteration: int) -> None:
        """
        Generates part of the mapping matrix for OLS where the number
        of rows depends on the iteration number. 

        Parameters:
        - iteration (int): The iteration number for trasmit/receive.
        """
        # Create a Latin square
        self.partial_mapping_matrix = np.zeros((self.s, self.n), dtype=np.int8)
        latin_square_num = iteration - 1

        latin_square = (np.arange(self.s) + latin_square_num * np.arange(self.s).reshape(self.s, 1)) % self.s

        # Compute the row and column indices for each element
        rows = np.arange(self.n) // self.s
        cols = np.arange(self.n) % self.s

        ones_indices = latin_square[rows, cols]
        self.partial_mapping_matrix[ones_indices, np.arange(self.n)] = 1

        start = (iteration - 1) * self.s
        end = iteration * self.s
        self.mapping_matrix[start:end] = self.partial_mapping_matrix

        self.mapping_matrix_used_rows += self.s

    def get_current_mapping_rows(self):
        """
        Gets number of current rows for mapping matrix that are
        used for listing.
        """
        return self.mapping_matrix_used_rows