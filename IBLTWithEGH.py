from typing import List, Set, Tuple
from Cell import Cell
from queue import Queue
from functools import reduce

class IBLTWithEGH:
    def __init__(self, symbols: Set[int], n: int, primes: List[int]):
        """
        Initializes the Invertible Bloom Lookup Table with
        combinatorial method EGH.

        Parameters:
        - symbols (Set[int]): set of source symbols.
        - n (int) - universe size.
        - primes (List[int]): list of primes for mapping matrix.
        """

        self.symbols = symbols 
        # universe size
        self.n = n 
        # set size 
        self.primes = primes 
        self.mapping_matrix = [[0 for _ in range(n)] for _ in range(sum(primes))]
        self.cells_queue = Queue()
        self.ack_queue = Queue()
        self.trasmit_iterations = 0
        self.receive_iterations = 0
        self.d = 0
        self.sender_cells = []
        self.receiver_cells = []
    
    def generate_egh_mapping(self) -> None:
        """
        Generates a mapping matrix for EGH.
        """
        row = 0

        for symbol in range(1, self.n+1):
            row = 0

            for prime in self.primes:
                res = symbol % prime
                
                for i in range(prime):
                    if i == res:
                        self.mapping_matrix[row][symbol-1] = 1
                    else:
                        self.mapping_matrix[row][symbol-1] = 0

                    row += 1


    def transmit(self) -> None:
        """
        Sends each iteration amount of cells of the IBLT to the receiver.
        """

        if self.trasmit_iterations >= len(self.primes):
            self.cells_queue.put("end")
            return

        if not self.ack_queue.empty():
            ack = self.ack_queue.get()

            if ack == "stop":
                self.cells_queue.put("terminated")
                return

        # Current row to start in mapping matrix.
        row = sum(self.primes[:self.trasmit_iterations])

        cells = []

        for i in range(self.primes[self.trasmit_iterations]):
            cells.append(Cell())

        for i in range(self.primes[self.trasmit_iterations]):
            for symbol in self.symbols:
                mapping_value = self.mapping_matrix[row][symbol-1]

                if mapping_value == 1:
                    cells[i].add(symbol)

            row += 1

        self.trasmit_iterations += 1

        for c in cells:
            self.cells_queue.put(c)

        self.cells_queue.put("end")

    def receive(self, sender_cells: List[Cell]) -> List[int]:
        """
        Receives transmitted cells and performs decoding to retrieve 
        the symmetric difference.

        Parameters:
        - received_cells (List[int]): List of sender cells.

        Returns:
        - Set[int]: List of integers representing the symmetric difference.
        """
        if not sender_cells:
            raise ValueError("No cells received from sender - increase prime array.")

        self.receive_iterations += 1
        self.sender_cells.extend(sender_cells)

        if self.receive_iterations == 1:
            # Calculate d - number of elements received from sender.
            self.d = sum([c.counter for c in sender_cells])

        # Check if free zone is guaranteed.
        # If not, more IBLTWithEGH cells are needed.
        primes_mul = reduce(lambda x, y: x*y, self.primes[:self.receive_iterations])
        lower_bound = self.n**self.d

        # Free zone is not guaranteed.
        if lower_bound > primes_mul:
            return []
        
        # Create IBLT for the receiver.
        self.receiver_cells = []
        for i in range(sum(self.primes[:self.receive_iterations])):
            self.receiver_cells.append(Cell())

        for row in range(len(self.receiver_cells)):
            for symbol in self.symbols:
                mapping_value = self.mapping_matrix[row][symbol-1]

                if mapping_value == 1:
                    self.receiver_cells[row].add(symbol)

        while True:
            symbol = self.peeling_decoder(self.sender_cells)

            if symbol == None:
                break

            # Extreme case - symbol is bigger than current max prime - so more
            # cells are required for listing to success.
            if self.primes[self.receive_iterations-1] < symbol:
                return []

            # Remove from sender IBLTWithEGH
            cell_idx = 0
            for p in self.primes[:self.receive_iterations]:
                cell_idx += symbol % p
                self.sender_cells[cell_idx].remove(symbol)

                cell_idx += p - symbol % p
            
            # Remove from receiver IBLTWithEGH if exists in sender IBLTWithEGH.
            # Otherwise, add the symbol to the receiver IBLTWithEGH.
            # In the end the goal is to get IBLTWithEGH which represents 
            # the symmetric difference and by listing getting the elements
            # of the symmetric difference.

            flag = self.is_symbol_inside_IBLTWithEGH(self.receiver_cells, symbol)
            
            cell_idx = 0
            for p in self.primes[:self.receive_iterations]:
                cell_idx += symbol % p

                if flag == True:
                    self.receiver_cells[cell_idx].remove(symbol)

                else:
                    self.receiver_cells[cell_idx].add(symbol)

                cell_idx += p - symbol % p

        symmetric_difference = self.listing(self.receiver_cells)

        if len(symmetric_difference) < (len(self.symbols) - self.d):
            return []
        
        # Empty symmetric difference.
        elif not symmetric_difference:
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
