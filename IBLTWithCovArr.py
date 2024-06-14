# from typing import List, Set, Tuple
# from Cell import Cell
# from queue import Queue
# from functools import reduce

# class IBLTWithCovArr:
#     def __init__(self, symbols: Set[int], n: int):
#         """
#         Initializes the Invertible Bloom Lookup Table with
#         combinatorial method Binary Covering Arrays.

#         Parameters:
#         - symbols (Set[int]): set of source symbols.
#         - n (int) - universe size.
#         """

#         self.symbols = symbols 
#         # universe size
#         self.n = n 
#         self.mapping_matrix = [[0 for _ in range(n)] for _ in range(sum(primes))]
#         self.cells_queue = Queue()
#         self.ack_queue = Queue()
#         self.trasmit_iterations = 0
#         self.receive_iterations = 0
#         self.sender_cells = []
#         self.receiver_cells = []

import math
from itertools import combinations, product

def construct_binary_can(k, t):
    if t > k:
        raise ValueError("t cannot be greater than k")
    
    s = k - t  

    # Base cases
    if t == 1:
        return [[0] * k, [1] * k]

    # Boundary condition: CAN(k, k-s) ≤ 2^k / (s + 1)
    boundary = math.ceil(2**k / (s + 1))

    # Construct CAN(k, t) from CAN(k, t-1)
    can_t_minus_1 = construct_binary_can(k, t - 1)
    # Start with the previous CAN
    result = can_t_minus_1[:]  

    # Generate all possible combinations of t columns
    all_combinations = list(combinations(range(k), t))

    # Generate all possible binary rows of length t
    all_binary_rows = list(product([0, 1], repeat=t))

    # Ensure each combination of t columns has all 2^t combinations of binary values
    for combo in all_combinations:
        existing_combinations = set(tuple(row[i] for i in combo) for row in result)

        for binary_row in all_binary_rows:
            if binary_row not in existing_combinations:
                # Construct a new row with the given binary_row in the specified columns
                new_row = [0] * k
                for idx, col in enumerate(combo):
                    new_row[col] = binary_row[idx]
                result.append(new_row)

                # Update the existing combinations set
                existing_combinations.add(binary_row)

    return result

if __name__ == "__main__":
    # Example usage
    k = 4  # Number of symbols
    t = 4  # Strength

    # Construct binary CAN(k, t)
    binary_can_t = construct_binary_can(k, t)

    print(f"Binary CAN(k={k}, t={t}):")
    for row in binary_can_t:
        print(row)

    print(f"Number of IBLT cells (number of rows) required: {len(binary_can_t)}")