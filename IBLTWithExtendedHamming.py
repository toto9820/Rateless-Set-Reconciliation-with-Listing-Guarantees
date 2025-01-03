import os
import numpy as np
from typing import List, Set
from IBLT import IBLT

class IBLTWithExtendedHamming(IBLT):
    def __init__(self, symbols: List[int], n: int, set_inside_set: bool = True):
        """
        Initializes an Invertible Bloom Lookup Table with
        extended hamming method.

        Parameters:
        - symbols (List[int]): set of source symbols.
        - n (int) - universe size.
        - set_inside_set (bool) - flag indicating whether a superset assumption holds, i.e. one participant's set
        includes the other.
        """
        super().__init__(symbols, n)
        self.mapping_matrix_file = 'extended_hamming_mapping_matrix.dat'
        self.mapping_matrix_used_rows = 0

        # Check if file exists, open in r+ mode if so, otherwise create it in w+ mode
        # For now number of rows 10^6 until solution to resizing
        # in case offset bigger than n is found.
        if os.path.exists(self.mapping_matrix_file):
            self.mapping_matrix = np.memmap(self.mapping_matrix_file, dtype=np.int8, mode='r+', shape=(10**6, n))
        else:
            self.mapping_matrix = np.memmap(self.mapping_matrix_file, dtype=np.int8, mode='w+', shape=(10**6, n))

    def generate_mapping(self, iteration: int) -> None:
        """
        Generates part of the mapping matrix for Extended Hamming where the number
        of rows depends on the iteration number. 
        The maximum symmetric difference size is 3.

        Parameters:
        - iteration (int): The iteration number for trasmit/receive.
        """
        if iteration == 1:
            # First row (all 1s)
            self.partial_mapping_matrix = np.ones((1, self.n), dtype=np.int8)
            self.mapping_matrix[:1] = self.partial_mapping_matrix
            self.mapping_matrix_used_rows = 1
            return

        period = 2 ** (iteration - 2)

        if period > self.n:
            return
        
        # Create the two alternating blocks
        block1 = np.concatenate((np.zeros(period, dtype=np.int8), np.ones(period, dtype=np.int8)))
        block2 = np.concatenate((np.ones(period, dtype=np.int8), np.zeros(period, dtype=np.int8)))

        # Calculate the number of repetitions needed
        num_repetitions = int(np.ceil(self.n / period))

        # Create the partial mapping matrix
        self.partial_mapping_matrix = np.vstack([
            np.tile(block1, num_repetitions)[:self.n],
            np.tile(block2, num_repetitions)[:self.n]
        ])

        # Update the mapping matrix
        start = self.mapping_matrix_used_rows
        end = start + self.partial_mapping_matrix.shape[0]
        self.mapping_matrix[start:end] = self.partial_mapping_matrix

        self.mapping_matrix_used_rows = end


    def get_current_mapping_rows(self):
        """
        Gets number of current rows for mapping matrix that are
        used for listing.
        """
        return self.mapping_matrix_used_rows