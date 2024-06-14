from typing import List, Set, Tuple
from Cell import Cell
from queue import Queue
from functools import reduce

class IBLTWithEGH:
    def __init__(self, symbols: Set[int], n: int):
        """
        Initializes the Invertible Bloom Lookup Table with
        combinatorial method EGH.

        Parameters:
        - symbols (Set[int]): set of source symbols.
        - n (int) - universe size.
        """
        # The sender/receiver set.
        self.symbols = symbols 
        # Universe size
        self.n = n 
        # Finite array of primes.
        self.primes = []
        # The mapping of each symbol to IBLT cells.
        self.mapping_matrix = []
        # The link to pass IBLT cells from sender to receiver (simulation
        # of a real communication link)
        self.cells_queue = Queue()
        # The link to pass Acknowledgement (ACK) to stop sending cells 
        # or Negative Acknowledgement (NACK) to send more cells
        #  from receiver to sender (simulation
        # of a real communication link)
        self.ack_queue = Queue()
        # Number of iterations the sender transmit cells to the receiver
        # (Sender side)
        self.trasmit_iterations = 0
        # Number of iterations the receiver gets cells from the sender
        # (Receiver side)
        self.receive_iterations = 0
        # All IBLT cells of sender that the receiver holds.
        self.iblt_sender_cells = []
        # All IBLT cells the receiver holds.
        self.iblt_receiver_cells = []
        # IBLT cells of the symmetric difference.
        self.iblt_diff_cells = []

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
    
    def generate_egh_mapping(self, iteration: int) -> None:
        """
        Generates part of the mapping matrix for EGH where the number
        of rows depends on the iteration number. 
        """
        prime = None

        if self.primes == []:
            prime = 2
        else:
            prime = self.get_next_prime(self.primes[-1])

        self.primes.append(prime)

        self.mapping_matrix = [[0 for _ in range(self.n)] 
                                for _ in range(prime)] 

        for symbol in range(1, self.n+1):

            res = symbol % prime
        
            for i in range(prime):
                if i == res:
                    self.mapping_matrix[i][symbol-1] = 1
                else:
                    self.mapping_matrix[i][symbol-1] = 0
            
    def transmit(self) -> None:
        """
        Sends each iteration amount of IBLT cells of the sender to the receiver.
        """
        if not self.ack_queue.empty():
            ack = self.ack_queue.get()

            # Receiver tells to stop sending cells.
            if ack == "stop":
                self.cells_queue.put("terminated")
                return
            
        self.trasmit_iterations += 1

        self.generate_egh_mapping(self.trasmit_iterations)

        cells = []

        for i in range(len(self.mapping_matrix)):
            cells.append(Cell())

        for i in range(len(self.mapping_matrix)):
            for symbol in self.symbols:
                mapping_value = self.mapping_matrix[i][symbol-1]

                if mapping_value == 1:
                    cells[i].add(symbol)

        for c in cells:
            self.cells_queue.put(c)

        self.cells_queue.put("end")

    def receive(self, iblt_sender_cells: List[Cell]) -> List[int]:
        """
        Receives transmitted cells and performs decoding to retrieve 
        the symmetric difference.

        Parameters:
        - iblt_sender_cells (List[int]): List of IBLT cells from the sender.

        Returns:
        - Set[int]: List of integers representing the symmetric difference.
        """
        if not iblt_sender_cells:
            raise ValueError("No cells received from sender - increase prime array.")

        self.receive_iterations += 1
        self.iblt_sender_cells.extend(iblt_sender_cells)

        if self.receive_iterations == 1:
            # Calculate number of elements received from sender.
            self.sender_set_size = sum([c.counter for c in iblt_sender_cells])

        self.generate_egh_mapping(self.receive_iterations)

        prev_rows_cnt = len(self.iblt_receiver_cells)
        
        # Create IBLT cells for the receiver.        
        for i in range(len(self.mapping_matrix)):
            self.iblt_receiver_cells.append(Cell())   

        for row in range(len(self.mapping_matrix)):
            for symbol in self.symbols:
                mapping_value = self.mapping_matrix[row][symbol-1]

                if mapping_value == 1:
                    self.iblt_receiver_cells[prev_rows_cnt+row].add(symbol)
        
        self.iblt_diff_cells = []
             
        for cell_idx in range(len(self.iblt_receiver_cells)):
            self.iblt_diff_cells.append(Cell())  

            self.iblt_diff_cells[cell_idx].sum =  self.iblt_receiver_cells[cell_idx].sum ^ self.iblt_sender_cells[cell_idx].sum

            if self.iblt_receiver_cells[cell_idx].checksum == 0:
                self.iblt_diff_cells[cell_idx].checksum = self.iblt_sender_cells[cell_idx].checksum
            
            elif self.iblt_sender_cells[cell_idx].checksum == 0:
                self.iblt_diff_cells[cell_idx].checksum = self.iblt_receiver_cells[cell_idx].checksum
            
            else:
                self.iblt_diff_cells[cell_idx].checksum =  bytes(a ^ b for a, b in zip(self.iblt_receiver_cells[cell_idx].checksum, self.iblt_sender_cells[cell_idx].checksum))
            
            self.iblt_diff_cells[cell_idx].counter -= self.iblt_receiver_cells[cell_idx].counter - self.iblt_sender_cells[cell_idx].counter

        # Check if free zone is guaranteed for BILT of symmetric difference.
        # If not, more IBLTWithEGH cells are needed.
        primes_mul = reduce(lambda x, y: x*y, self.primes[:self.receive_iterations])
        symmetric_difference_size = sum([abs(c.counter) for c in self.iblt_diff_cells[:2]])
        lower_bound = self.n**symmetric_difference_size

        # Free zone is not guaranteed.
        if lower_bound > primes_mul:
            return []

        symmetric_difference = self.listing(self.iblt_diff_cells)
        
        # Empty symmetric difference.
        if not symmetric_difference:
            return ["empty set"]
    
        return symmetric_difference
    
    def is_symbol_inside_IBLTWithEGH(self, cells: List[Cell], symbol: int) -> bool:
        """
        Checks for 1 for each chunk of cells for specific symbol.
        
        Parameters:
        - cells (List[Cell]): List of cells to perform the check on.
        - symbol (int): source symbol to check if might be in IBLT's cells.

        Returns:
        - bool: The query answer if symbol might be in IBLT.
        """
        cell_idx = 0
        for p in self.primes[:self.receive_iterations]:
            cell_idx += symbol % p

            if cells[cell_idx].counter == 0:
                return False

            cell_idx += p - symbol % p
        
        return True


    def peeling_decoder(self, cells: List[Cell]) -> int:
        """
        Extracts a soruce symbol from IBLT.

        Parameters:
        - cells (List[Cell]): List of cells to perform the peeling on.

        Returns:
        - int: The source symbol value.
        """
        symbol = None

        for cell in cells:
            if cell.is_pure_cell():
                symbol = cell.sum 
                
                return symbol

        return symbol     
    
    def listing(self, cells) -> List[int]:
        """
        Performs listing to the IBLT.

        Parameters:
        - cells (List[cells]): List of cells to perform the listing on.

        Returns:
        - List[int]: List of integers (type of source symbols) in the IBLT.
        """
        symbols = []

        while True:
            symbol = self.peeling_decoder(cells)

            if symbol == None:
                break

            symbols.append(symbol)

            cell_idx = 0
            for p in self.primes[:self.receive_iterations]:
                cell_idx += symbol % p
                cells[cell_idx].remove(symbol)

                cell_idx += p - symbol % p

        return sorted(symbols)

