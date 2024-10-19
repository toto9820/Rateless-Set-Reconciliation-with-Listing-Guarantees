from concurrent.futures import ThreadPoolExecutor, as_completed
import numpy as np
# To utilize GPU 
from typing import List, Set, Tuple
from Cell import Cell
from queue import Queue
from functools import reduce
import numba as nb
from scipy.sparse import csr_matrix
from functools import partial
from multiprocessing import Lock
from multiprocessing.pool import ThreadPool as Pool
from hashlib import sha256
# A faster hashing algorithm
from xxhash import xxh32_intdigest, xxh64_intdigest, xxh3_64_intdigest, xxh64_hexdigest

@nb.njit(cache=True)
def fast_encode(partial_mapping_matrix, histogram, rows):
    row_indices, col_indices = partial_mapping_matrix.nonzero()
    mask = histogram[col_indices] > 0
    filtered_row_indices = row_indices[mask]
    mapped_symbols = col_indices[mask] + 1

    counters = np.bincount(filtered_row_indices, minlength=rows)
    sums = np.zeros(rows, dtype=np.int64)
    checksums = np.zeros(rows, dtype=np.uint64)

    for i, row in enumerate(filtered_row_indices):
        sums[row] ^= mapped_symbols[i]
        # checksums[row] ^= hash(mapped_symbols[i])

    return counters, sums, checksums
        
class IBLT:
    def __init__(self, symbols: np.ndarray, n: int, set_inside_set: bool = True, hash_func='xxh64'):
        """
        Initializes the Rateless Invertible Bloom Lookup Table.

        Parameters:
        - symbols (List[int]): set of source symbols.
        - n (int) - universe size.
        - set_inside_set (bool) - flag indicating whether a superset assumption holds, i.e. one participant's set
        includes the other.
        """
        # The sender/receiver set.
        # self.symbols = np.array(symbols)
        self.symbols = symbols
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
        self.other_list_for_debug = np.array([])

        # TODO - hash of transactions is in string and not int (they are the symbols) form - 
        # should enable to define options to symbols type - int, str, etc.
        if hash_func == 'xxh32':
            self.hash_func = xxh32_intdigest
        elif hash_func == 'xxh64':
            self.hash_func = xxh64_intdigest
        elif hash_func == 'xxh3_64':
            self.hash_func = xxh3_64_intdigest

        self.sender_counters = np.array([], dtype=np.int64)
        self.sender_sums = np.array([], dtype=np.int64)
        self.sender_checksums = np.array([], dtype=np.uint64)
        
        self.receiver_counters = np.array([], dtype=np.int64)
        self.receiver_sums = np.array([], dtype=np.int64)
        self.receiver_checksums = np.array([], dtype=np.uint64)
        
        self.diff_counters = np.array([], dtype=np.int64)
        self.diff_sums = np.array([], dtype=np.int64)
        self.diff_checksums = np.array([], dtype=np.uint64)

        self.histogram = np.bincount(self.symbols_indices, minlength=self.n)

        self.vectorized_hash_func = np.vectorize(self.hash_func, otypes=[np.uint64])

    def generate_mapping(self) -> None:
        """
        Generates part of the mapping matrix for specific method where the number
        of rows depends on the iteration number. 

        Parameters:
        - iteration (int): The iteration number for trasmit/receive.
        """
        raise NotImplementedError("Please Implement this method")
    
    def get_current_mapping_rows(self):
        """
        Gets number of current rows for mapping matrix that are
        used for listing.
        """
        raise NotImplementedError("Please Implement this method")
    
    def encode(self):
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

        # iblt_sender_cells = [Cell(self.set_inside_set) for _ in range(rows)]
        # counters = np.zeros(rows, dtype=np.int64)
        # sums = np.zeros(rows, dtype=np.int64)
        # checksums = np.zeros(rows, dtype=np.uint64)  

        if self.set_inside_set:
            counters, sums, checksums = fast_encode(
                self.partial_mapping_matrix, self.histogram, rows)
                        
            # TODO: Implement checksum calculation if needed

        # TBD with isin or intersect1d or like if case - need to check.
        else:
            print("TBD")

        # for row in range(rows):
        #     # # Get the indices where the row has a value of 1.
        #     # mask_symbols_indices = np.intersect1d(self.partial_mapping_matrix[row].indices, self.symbols_indices)
        #     # # Get the symbols corresponding to these indices.
        #     # mapped_symbols = mask_symbols_indices + 1

        #     # # Get the boolean mask where the row has a value of 1.
        #     # mask_symbols = np.isin(self.partial_mapping_matrix[row].indices, self.symbols_indices)
        #     # # Get the corresponding symbols where a row has a value of 1.
        #     # mapped_symbols = self.partial_mapping_matrix[row].indices[mask_symbols] + 1

        #     # For faster execution
        #     if self.set_inside_set:
        #         # # Convert the sparse row to a dense array
        #         # row_array = self.partial_mapping_matrix[row].toarray().flatten()

        #         # # Create a boolean mask for the symbols
        #         # symbols_mask_1 = np.zeros(row_array.shape, dtype=bool)
        #         # symbols_mask_1[self.symbols_indices] = True

        #         # # Use boolean indexing to get the mapped symbols
        #         # mapped_symbols_1 = np.where((row_array == 1) & symbols_mask_1)[0] + 1

        #         # symbols_mask_1 = is_in_set_pnb(self.partial_mapping_matrix[row].indices, self.symbols_indices)
                
        #         # histogram = np.zeros(self.n, dtype=np.uint64)

        #         # for index in self.symbols_indices:
        #         #     # Increment the count for each index
        #         #     histogram[index] += 1  

        #         # Step 2: Check membership using the histogram
        #         # symbols_mask_1 = np.array([index for index in self.partial_mapping_matrix[row].indices if histogram[index] > 0])
        #         symbols_mask_1 = filter_indices(self.partial_mapping_matrix.getrow(row).indices, histogram)
        #         mapped_symbols_1 = symbols_mask_1 + 1
        #         # mapped_symbols_1 = self.partial_mapping_matrix[row].indices[symbols_mask_1] + 1
            
        #     iblt_sender_cells[row].add_multiple(mapped_symbols_1)

        return counters, sums, checksums
    
    # def decode(self, iblt_sender_cells: List[Cell]) -> List[int]:
    def decode(self, sender_counters: np.ndarray, sender_sums: np.ndarray, sender_checksums: np.ndarray) -> List[int]:
        """
        Decodes cells to retrieve the symmetric difference.

        Parameters:
        - iblt_sender_cells (List[int]): List of IBLT cells from the sender.

        Returns:
        - Set[int]: List of integers representing the symmetric difference.
        """
        # if not iblt_sender_cells:
        #     raise ValueError("No cells received from sender.")

        if sender_counters.size == 0:
            raise ValueError("No cells received from sender.")

        self.receive_iterations += 1
        # self.iblt_sender_cells.extend(iblt_sender_cells)

        self.sender_counters = np.append(self.sender_counters, sender_counters)
        self.sender_sums = np.append(self.sender_sums, sender_sums)
        self.sender_checksums = np.append(self.sender_checksums, sender_checksums)

        if self.receive_iterations == 1:
            self.sender_set_size = np.sum(sender_counters)

        receiver_counters, receiver_sums, receiver_checksums = self.encode()

        self.receiver_counters = np.append(self.receiver_counters, receiver_counters)
        self.receiver_sums = np.append(self.receiver_sums, receiver_sums)
        self.receiver_checksums = np.append(self.receiver_checksums, receiver_checksums)

        self.diff_counters = self.receiver_counters - self.sender_counters
        self.diff_sums = np.bitwise_xor(self.receiver_sums, self.sender_sums)
        # TOOD - fix later
        self.diff_checksums = np.bitwise_xor(self.receiver_checksums, self.sender_checksums) 
        
        symmetric_difference = self.listing()

        return symmetric_difference

    def listing(self) -> List[int]:
        symbols = []

        while True:
            pure_cells = np.where(np.abs(self.diff_counters) == 1)[0]
            
            if len(pure_cells) == 0:
                break

            for cell_num in pure_cells:
                if self.diff_counters[cell_num] == 0:
                    continue
                
                symbol = self.diff_sums[cell_num]

                if (symbol > self.n):
                    continue

                symbols.append(symbol)

                col_index = symbol - 1

                # mapped_rows = [
                # (np.argmax(partial_mapping_matrix[:, col_index]) + offset)
                # for partial_mapping_matrix, offset in self.mapping_matrix]

                cur_mapping_rows = self.get_current_mapping_rows()

                mapped_rows = self.mapping_matrix[:cur_mapping_rows, col_index].nonzero()
                
                self.diff_counters[mapped_rows] -= np.sign(self.diff_counters[mapped_rows], dtype=int)
            
                self.diff_sums[mapped_rows] ^= self.diff_sums[cell_num]
        
        return symbols
        
    
    
  