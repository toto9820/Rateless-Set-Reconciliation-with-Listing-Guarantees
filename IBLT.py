from typing import List, Set, Tuple
from Cell import Cell
from queue import Queue
from functools import reduce

class IBLT:
    def __init__(self, symbols: Set[int], n: int):
        """
        Initializes the Rateless Invertible Bloom Lookup Table.

        Parameters:
        - symbols (Set[int]): set of source symbols.
        - n (int) - universe size.
        """
        # The sender/receiver set.
        self.symbols = symbols 
        # Universe size
        self.n = n 
        # Partial mapping matrix of each symbol to IBLT cells.
        self.partial_mapping_matrix = []
        # The whole mapping matrix of each symbol to IBLT cells.
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
        # Indicator if stopping condition for sender exists.
        self.stopping_condition_exists = False
        # The size of the symmetric difference.
        self.symmetric_difference_size = 0

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

        cells = []

        for i in range(len(self.partial_mapping_matrix)):
            cells.append(Cell())

        for i in range(len(self.partial_mapping_matrix)):
            for symbol in self.symbols:
                mapping_value = self.partial_mapping_matrix[i][symbol-1]

                if mapping_value == 1:
                    cells[i].add(symbol)

        for c in cells:
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

        # The number of IBLT cells before the current iteration received.
        prev_rows_cnt = len(self.iblt_receiver_cells)
        
        self.generate_mapping(self.receive_iterations)

        # Create IBLT cells for the receiver.        
        for i in range(len(self.partial_mapping_matrix)):
            self.iblt_receiver_cells.append(Cell())   

        for row in range(len(self.partial_mapping_matrix)):
            for symbol in self.symbols:
                mapping_value = self.partial_mapping_matrix[row][symbol-1]

                if mapping_value == 1:
                    self.iblt_receiver_cells[prev_rows_cnt+row].add(symbol)
        
        self.iblt_diff_cells = self.calc_iblt_diff(self.iblt_sender_cells,
                                                   self.iblt_receiver_cells)      

        # Check if free zone is guaranteed for IBLT of symmetric difference.
        # If not, more IBLT cells are needed.
        # if self.stopping_condition_exists:
        #     if self.sender_should_halt_check() == False:
        #         return []

        symmetric_difference = self.listing(self.iblt_diff_cells)
        
        # Failure to decode.
        if symmetric_difference == ["Decode Failure"]:
            return []
                
        return symmetric_difference
    

    def calc_iblt_diff(self, iblt_sender: List[int], iblt_receiver: List[int]):
        """
        Calculates the IBLT of symmetric difference.

        Parameters:
        - iblt_sender (List[cells]): IBLT cells of the sender.
        - iblt_receiver (List[cells]): IBLT cells of the receiver.

        Returns:
        - List[int]: IBLT cells of the symmetric difference.
        """
        iblt_diff = []
               
        for cell_idx in range(len(iblt_receiver)):
            iblt_diff.append(Cell())  

            iblt_diff[cell_idx].sum =  iblt_receiver[cell_idx].sum ^ iblt_sender[cell_idx].sum

            if iblt_receiver[cell_idx].checksum == 0:
                iblt_diff[cell_idx].checksum = iblt_sender[cell_idx].checksum
            
            elif iblt_sender[cell_idx].checksum == 0:
                iblt_diff[cell_idx].checksum = iblt_receiver[cell_idx].checksum
            
            else:
                iblt_diff[cell_idx].checksum =  bytes(a ^ b for a, b in zip(iblt_receiver[cell_idx].checksum, iblt_sender[cell_idx].checksum))
            
            iblt_diff[cell_idx].counter -= iblt_receiver[cell_idx].counter - iblt_sender[cell_idx].counter
        
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

            for row in range(len(self.mapping_matrix)):
                mapping_value = self.mapping_matrix[row][symbol-1]

                if mapping_value == 1 and row < len(cells):
                    cells[row].remove(symbol)

        # Empty symmetric difference
        if symbols == []:
            return ["empty set"]

        return sorted(symbols)     
        
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
        for cell_idx in range(len(iblt_cells)):
            if iblt_cells[cell_idx].is_empty_cell() == False:
                return False
            
        return True

    
  

