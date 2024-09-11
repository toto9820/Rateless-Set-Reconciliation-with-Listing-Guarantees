import numpy as np
# To utilize GPU 
from typing import List, Set, Tuple
from Cell import Cell
from queue import Queue
from functools import reduce
import numba as nb
from scipy.sparse import csr_matrix

@nb.njit(parallel=True)
def is_in_set_pnb(a, b):
    n = len(a)
    result = np.zeros(n, dtype=np.bool_)
    set_b = set(b)

    for i in nb.prange(n):
        result[i] = a[i] in set_b

    return result

class IBLT:
    def __init__(self, symbols: List[int], n: int, set_inside_set: bool = True):
        """
        Initializes the Rateless Invertible Bloom Lookup Table.

        Parameters:
        - symbols (List[int]): set of source symbols.
        - n (int) - universe size.
        - set_inside_set (bool) - flag indicating whether a superset assumption holds, i.e. one participant's set
        includes the other.
        """
        # The sender/receiver set.
        self.symbols = np.array(symbols)
        # Symbols indices in 0 indexing.
        self.symbols_indices = self.symbols - 1
        # Universe size
        self.n = n 
        # Superset assumption
        self.set_inside_set = set_inside_set
        # Partial mapping matrix of each symbol to IBLT cells.
        self.partial_mapping_matrix = []
        # The whole mapping matrix of each symbol to IBLT cells (sparse - to save memory).
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
        # IBLT cells of the sender (in receiver side).
        self.iblt_sender_cells = []
        # IBLT cells of the receiver.
        self.iblt_receiver_cells = []
        # IBLT cells of the symmetric difference.
        self.diff_cells = []
        # The size of the symmetric difference.
        self.symmetric_difference_size = 0
        # Sender set for debugging.
        self.other_list_for_debug = set()

    def generate_mapping(self) -> None:
        """
        Generates part of the mapping matrix for specific method where the number
        of rows depends on the iteration number. 

        Parameters:
        - iteration (int): The iteration number for trasmit/receive.
        """
        raise NotImplementedError("Please Implement this method")

    def encode(self) -> list[Cell]:
        """
        Encodes specific amount of IBLT cells.
        """
        if not self.ack_queue.empty():
            ack = self.ack_queue.get()

            # Receiver tells to stop sending cells.
            if ack == "stop":
                self.cells_queue.put("terminated")
                return
            
        self.trasmit_iterations += 1

        self.generate_mapping(self.trasmit_iterations)
        
        # Add IBLT cells for the sender.
        rows = self.partial_mapping_matrix.shape[0]

        iblt_sender_cells = [Cell(self.set_inside_set) for _ in range(rows)]

        for row in range(rows):
            # # Get the indices where the row has a value of 1.
            # mask_symbols_indices = np.intersect1d(self.partial_mapping_matrix[row].indices, self.symbols_indices)
            # # Get the symbols corresponding to these indices.
            # mapped_symbols = mask_symbols_indices + 1

            # # Get the boolean mask where the row has a value of 1.
            # mask_symbols = np.isin(self.partial_mapping_matrix[row].indices, self.symbols_indices)
            # # Get the corresponding symbols where a row has a value of 1.
            # mapped_symbols = self.partial_mapping_matrix[row].indices[mask_symbols] + 1

            # For faster execution
            if self.set_inside_set:
                # # Convert the sparse row to a dense array
                # row_array = self.partial_mapping_matrix[row].toarray().flatten()

                # # Create a boolean mask for the symbols
                # symbols_mask_1 = np.zeros(row_array.shape, dtype=bool)
                # symbols_mask_1[self.symbols_indices] = True

                # # Use boolean indexing to get the mapped symbols
                # mapped_symbols_1 = np.where((row_array == 1) & symbols_mask_1)[0] + 1

                symbols_mask_1 = is_in_set_pnb(self.partial_mapping_matrix[row].indices, self.symbols_indices)
                mapped_symbols_1 = self.partial_mapping_matrix[row].indices[symbols_mask_1] + 1
            
            iblt_sender_cells[row].add_multiple(mapped_symbols_1)

        return iblt_sender_cells
    
    def decode(self, iblt_sender_cells: List[Cell]) -> List[int]:
        """
        Decodes cells to retrieve the symmetric difference.

        Parameters:
        - iblt_sender_cells (List[int]): List of IBLT cells from the sender.

        Returns:
        - Set[int]: List of integers representing the symmetric difference.
        """
        if not iblt_sender_cells:
            raise ValueError("No cells received from sender.")

        self.receive_iterations += 1
        self.iblt_sender_cells.extend(iblt_sender_cells)

        if self.receive_iterations == 1:
            # Calculate number of elements received from sender.
            self.sender_set_size = sum([c.counter for c in iblt_sender_cells])
        
        iblt_receiver_cells = self.encode()
        self.iblt_receiver_cells.extend(iblt_receiver_cells)
   
        self.diff_cells = self.calc_iblt_diff(self.iblt_sender_cells,
                                              self.iblt_receiver_cells)
                
        symmetric_difference = self.listing(self.diff_cells)
        
        # Failure to decode.
        if symmetric_difference == ["Decode Failure"]:
            return []

        elif symmetric_difference == ["empty set"]:
            return "empty set"
                
        return [int(symbol) for symbol in symmetric_difference]
    

    def calc_iblt_diff(self, iblt_sender_cells: List[int], iblt_receiver_cells: List[int]):
        """
        Calculates the IBLT of symmetric difference.

        Parameters:
        - iblt_sender (List[cells]): IBLT cells of the sender.
        - iblt_receiver (List[cells]): IBLT cells of the receiver.

        Returns:
        - List[int]: IBLT cells of the symmetric difference.
        """
        diff_cells = [Cell(self.set_inside_set) for _ in range(len(iblt_receiver_cells))]
            
        for cell_idx in range(len(iblt_receiver_cells)):
            diff_cells[cell_idx].counter = iblt_receiver_cells[cell_idx].counter - iblt_sender_cells[cell_idx].counter
            diff_cells[cell_idx].sum =  iblt_receiver_cells[cell_idx].sum ^ iblt_sender_cells[cell_idx].sum
            
            if (self.set_inside_set == False):
                diff_cells[cell_idx].checksum = iblt_receiver_cells[cell_idx].checksum ^ iblt_sender_cells[cell_idx].checksum

        return diff_cells
    
    
    def listing(self, cells: List[Cell], with_deocde_frac: bool = False) -> List[int]:
        """
        Performs listing to the IBLT.

        Parameters:
        - cells (List[cells]): List of cells to perform the listing on.
        - with_deocde_frac (bool): Fraction of recovered symbols of 
        IBLT.

        Returns:
        - List[int]: List of integers (type of source symbols) in the IBLT.
        """
        symbols = []
        symbols_cnt = None

        if with_deocde_frac:
            symbols_cnt = sum(abs(c.counter) for c in cells)

            if symbols_cnt == 0:
                return ["Decode Failure", 0]

        while True:
            symbol = self.peeling_decoder(cells)

            if symbol == None:
                # Check if IBLT is empty for symmetric difference 
                # empty set or decoding failure.

                # Decode Failure
                if self.is_iblt_empty(cells) == False:
                    if with_deocde_frac == False:
                        return ["Decode Failure"]
                    else:
                        return ["Decode Failure", len(symbols)/symbols_cnt]
                else:
                    break
            
            symbols.append(symbol)

            mapped_rows = self.mapping_matrix[:, symbol-1].nonzero()[0]
            
            for row in mapped_rows:
                if row < len(cells):
                    cells[row].remove(symbol)

        # Empty symmetric difference
        if symbols == []:
            return ["empty set"]

        return symbols
        
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

    def is_iblt_empty(self, iblt_cells: List[Cell]) -> bool:
        """
        Check if an IBLT is empty - all its cells are with counter = 0.

        Parameters:
        - cells (List[Cell]): List of cells of the IBLT.

        Returns:
        - bool: IBLT is empty (True) or not (False).
        """
        return all(cell.is_empty_cell() for cell in iblt_cells)
    
  