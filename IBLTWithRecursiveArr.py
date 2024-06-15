import math
from typing import List, Set, Tuple
from Cell import Cell
from queue import Queue
from functools import reduce
from itertools import combinations, product
from IBLT import IBLT

class IBLTWithRecursiveArr(IBLT):
    def __init__(self, symbols: Set[int], n: int):
        """
        Initializes the Invertible Bloom Lookup Table with
        recursive array method.

        Parameters:
        - symbols (Set[int]): set of source symbols.
        - n (int) - universe size.
        """
        super().__init__(symbols, n)
        self.stopping_condition_exists = False

    def generate_mapping(self, iteration: int) -> None:
        """
        Generates part of the mapping matrix for Binary Covering Array
        where the number of rows depends on the iteration number. 
        """
        if iteration > self.n:
            raise ValueError("iteration cannot be greater than universe size.")
        
        # Base case
        if iteration == 1:
            self.partial_mapping_matrix = [[1] * self.n]
            self.mapping_matrix = [[1] * self.n]
            return

        self.partial_mapping_matrix = [[0 for _ in range(self.n)] 
                                for _ in range(iteration)] 
        
        prev_rows_cnt = len(self.mapping_matrix)
        
        for i in range(iteration):
            self.mapping_matrix.append([0 for _ in range(self.n)])

            for symbol in range(1, self.n+1):
                res = symbol % iteration

                if res == i:
                    self.partial_mapping_matrix[i][symbol-1] = 1
                    self.mapping_matrix[prev_rows_cnt+i][symbol-1] = 1
                else:
                    self.partial_mapping_matrix[i][symbol-1] = 0
                    self.mapping_matrix[prev_rows_cnt+i][symbol-1] = 0