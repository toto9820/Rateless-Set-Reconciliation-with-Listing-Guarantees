import math
from typing import List, Set, Tuple
from Cell import Cell
from queue import Queue
from functools import reduce
from itertools import combinations, product
from IBLT import IBLT
import numpy as np

class IBLTWithCovArr(IBLT):
    def __init__(self, symbols: List[int], n: int):
        """
        Initializes the Invertible Bloom Lookup Table with
        binary covering array method.

        Parameters:
        - symbols (List[int]): set of source symbols.
        - n (int) - universe size.
        """
        super().__init__(symbols, n)
        self.stopping_condition_exists = False

    def binomial_coefficient(self, k, t):
        return list(combinations(range(k), t))

    # def generate_mapping(self, iteration: int) -> None:
    #     """
    #     Generates part of the mapping matrix for Binary Covering Array
    #     where the number of rows depends on the iteration number. 
    #     """
    #     k = self.n
    #     t = iteration

    #     if t > k:
    #         raise ValueError("t cannot be greater than k")
        
    #     # Base case
    #     if t == 1:
    #         self.partial_mapping_matrix = [[0] * k, [1] * k]
    #         self.mapping_matrix = [[0] * k, [1] * k]
    #         return

    #     can_t_minus_1 = self.mapping_matrix
    #     self.partial_mapping_matrix = []

    #     # Generate all possible combinations of t columns
    #     all_combinations = list(combinations(range(k), t))

    #     # Generate all possible binary rows of length t.
    #     all_binary_rows = list(product([0, 1], repeat=t))

    #     # Ensure each combination of t columns has all 2^t combinations of binary values
    #     for combo in all_combinations:
    #         existing_combinations = set(tuple(row[i] for i in combo) for row in can_t_minus_1)

    #         for binary_row in all_binary_rows:
    #             if binary_row not in existing_combinations:
    #                 # Construct a new row with the given binary_row in the specified columns
    #                 new_row = [0] * k
    #                 for idx, col in enumerate(combo):
    #                     new_row[col] = binary_row[idx]
    #                 self.partial_mapping_matrix.append(new_row)
    #                 self.mapping_matrix.append(new_row)

    #                 # Update the existing combinations set
    #                 existing_combinations.add(binary_row)


    def generate_mapping(self, iteration: int) -> None:
        """
        Generates part of the mapping matrix for Binary Covering Array
        where the number of rows depends on the iteration number. 
        """
        k = self.n
        t = iteration 

        print(f"Iteration num: {iteration}")

        if t > k:
            raise ValueError("t cannot be greater than k")
        
        subset_indices = []
        
        initial_class = (math.floor(t / 2)) % (k - t + 1)
        separation = k - t + 1
        num_classes = math.floor((k - initial_class) / (k - t + 1))
        
        for j in range(num_classes+1):
            kClass = initial_class + separation * j
            subset_indices.extend(self.binomial_coefficient(k, kClass))
            
        self.partial_mapping_matrix = []

        for indices in subset_indices:
            row = np.zeros(k, dtype=int)
            row[list(indices)] = 1

            # Code to find whether all elements in row are zero.
            if not np.any(row):
                continue

            if any(np.array_equal(row, matrix_row) for matrix_row in self.mapping_matrix):
                continue

            self.partial_mapping_matrix.append(row.tolist())
            self.mapping_matrix.append(row.tolist())
        
