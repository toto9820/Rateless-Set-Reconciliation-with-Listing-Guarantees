import numpy as np
from typing import List, Set, Tuple
from Cell import Cell
from queue import Queue
from functools import reduce
from IBLT import IBLT
from scipy.sparse import csr_matrix, vstack

class IBLTWithEGH(IBLT):
    def __init__(self, symbols: Set[int], n: int):
        """
        Initializes the Invertible Bloom Lookup Table with
        combinatorial method EGH.

        Parameters:
        - symbols (Set[int]): set of source symbols.
        - n (int) - universe size.
        """
        super().__init__(symbols, n)
        self.stopping_condition_exists = True
        # Finite array of primes.
        self.primes = []

    def is_prime(self, num: int):
        """
        Checks if the current num is a prime number.

        Parameters:
        - num (int): The candidate to be a prime number.

        Returns:
        - bool: Whether num is a prime number or not.
        """
        if num < 2:
            return False
        for i in range(2, int(num ** 0.5) + 1):
            if num % i == 0:
                return False
        return True

    def get_next_prime(self, prev_prime: int):
        """
        Returns the next prime number after the given start value.
        If start is not provided or is less than 2, it starts from 2.

        Parameters:
        - prev_prime (int): The previous prime number.

        Returns:
        - int: The next prime number.
        """
        next_num = prev_prime + 1

        while not self.is_prime(next_num):
            next_num += 1

        return next_num
    
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

        partial_mapping_matrix = [[0 for _ in range(self.n)] 
                                for _ in range(prime)] 

        for symbol in range(1, self.n+1):

            res = symbol % prime
        
            for i in range(prime):
                if i == res:
                    partial_mapping_matrix[i][symbol-1] = 1
                else:
                    partial_mapping_matrix[i][symbol-1] = 0
        
        self.partial_mapping_matrix = csr_matrix(partial_mapping_matrix)

        if iteration == 1:
            self.mapping_matrix = self.partial_mapping_matrix
        else:
            self.mapping_matrix = csr_matrix(vstack([self.mapping_matrix, self.partial_mapping_matrix]))

    def sender_should_halt_check(self) -> bool:
        """
        Checks if a stopping condition specific to EGH is met and 
        accordingly tell the sender to stop transmitting cells.

        Returns:
        - bool: True - stop/ False - continue.
        """
        primes_mul = reduce(lambda x, y: x*y, self.primes[:self.receive_iterations])
        self.symmetric_difference_size = sum([abs(c.counter) for c in self.iblt_diff_cells[:2]])
        lower_bound = self.n**self.symmetric_difference_size

        # Free zone is not guaranteed.
        if lower_bound > primes_mul:
            return False
        else:
            return True
        
    
