import numpy as np
# To utilize GPU 
from typing import List, Set, Tuple
from Cell import Cell
from queue import Queue
from functools import reduce
from scipy.sparse import csr_matrix
from scipy.sparse import csr_matrix

class IBLT:
    def __init__(self, symbols: List[int], n: int):
        """
        Initializes the Rateless Invertible Bloom Lookup Table.

        Parameters:
        - symbols (List[int]): set of source symbols.
        - n (int) - universe size.
        """
        # The sender/receiver set.
        self.symbols = np.array(symbols)
        # Symbols indices in 0 indexing.
        self.symbols_indices = self.symbols - 1
        # Universe size
        self.n = n 
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
        self.iblt_diff_cells = []
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

        self.generate_mapping(self.trasmit_iterations)
        
        # Add IBLT cells for the sender.
        rows = self.partial_mapping_matrix.shape[0]

        iblt_sender_cells = [Cell() for _ in range(rows)]

        for row in range(rows):
            # Get the indices where the row has a value of 1.
            mask_symbols_indices = np.intersect1d(self.partial_mapping_matrix[row].indices, self.symbols_indices)
            # Get the symbols corresponding to these indices.
            mapped_symbols = mask_symbols_indices + 1

            # for symbol in mapped_symbols:
            #     iblt_sender_cells[row].add(symbol)

            iblt_sender_cells[row].add_multiple(mapped_symbols)

        for c in iblt_sender_cells:
            self.cells_queue.put(c)

        self.cells_queue.put("end")

    def sender_should_halt_check(self) -> bool:
        """
        Checks if a stopping condition specific to method is met and 
        accordingly tell the sender to stop transmitting cells.

        Returns:
        - bool: True - stop/ False - continue.
        """
        return False
    
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
            raise ValueError("No cells received from sender.")

        self.receive_iterations += 1
        self.iblt_sender_cells.extend(iblt_sender_cells)

        if self.receive_iterations == 1:
            # Calculate number of elements received from sender.
            self.sender_set_size = sum([c.counter for c in iblt_sender_cells])
        
        self.generate_mapping(self.receive_iterations)

        # Add IBLT cells for the receiver.
        rows = self.partial_mapping_matrix.shape[0]

        iblt_receiver_cells = [Cell() for _ in range(rows)]

        for row in range(rows):
            # Get the indices where the row has a value of 1.
            mask_symbols_indices = np.intersect1d(self.partial_mapping_matrix[row].indices, self.symbols_indices)
            # Get the symbols corresponding to these indices.
            mapped_symbols = mask_symbols_indices + 1

            # for symbol in mapped_symbols:
            #     iblt_receiver_cells[row].add(symbol)

            iblt_receiver_cells[row].add_multiple(mapped_symbols)

        self.iblt_receiver_cells.extend(iblt_receiver_cells)
   

        self.iblt_diff_cells = self.calc_iblt_diff(self.iblt_sender_cells,
                                                   self.iblt_receiver_cells)

        symmetric_difference = self.listing(self.iblt_diff_cells)
        
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
        iblt_diff = []
               
        for cell_idx in range(len(iblt_receiver_cells)):
            iblt_diff.append(Cell())  
            
            iblt_diff[cell_idx].counter = iblt_receiver_cells[cell_idx].counter - iblt_sender_cells[cell_idx].counter
            iblt_diff[cell_idx].sum =  iblt_receiver_cells[cell_idx].sum ^ iblt_sender_cells[cell_idx].sum
            iblt_diff[cell_idx].checksum = iblt_receiver_cells[cell_idx].checksum ^ iblt_sender_cells[cell_idx].checksum

        
        return iblt_diff
    
    
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
            symbols_cnt = sum([abs(c.counter) for c in cells])

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

            mapped_rows = np.where(self.mapping_matrix[:, symbol-1].toarray() == 1)[0]
            
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
    
  

