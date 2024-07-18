import csv
import os
os.environ['OPENBLAS_NUM_THREADS'] = '1'  # Set the number of OpenBLAS threads to 1 to avoid conflicts in multiprocessing
import random
import cProfile
import pstats
import io
import gc
import multiprocessing
from contextlib import contextmanager
from Method import Method
from IBLTWithEGH import IBLTWithEGH
from IBLTWithExtendedHamming import IBLTWithExtendedHamming
from IBLTWithBCH import IBLTWithBCH

# Global universe list shared between pool of processes to save memory.
universe_list = range(1, 100)

def generate_sender_receiver_iblts(symmetric_difference_size, method, set_inside_set):
    """
    Generate sender and receiver lists based on the given parameters.
    
    Parameters:
        symmetric_difference_size (int): The size of the symmetric difference.
        method (Method): The method to use for creating IBLTs.
        set_inside_set (bool): Flag indicating whether the receiver list should be a subset of the universe list.
    
    Returns:
        tuple: A tuple containing the sender IBLT and the receiver IBLT.
    """
    # Declaring that we're using the global universe list variable.
    global universe_list

    if set_inside_set:
        # Create receiver list from the universe list.
        receiver_list = universe_list
        # Randomly select elements to remove from the universe list to create the sender list.
        elements_to_remove = random.sample(universe_list, symmetric_difference_size)
        remove_set = set(elements_to_remove)
        sender_list = [elem for elem in universe_list if elem not in remove_set]
        
        del elements_to_remove
        del remove_set
        
    else:
        # Randomly determine the receiver list size and create it.
        receiver_size = max(symmetric_difference_size, random.randint(1, len(universe_list) - symmetric_difference_size))
        receiver_list = random.sample(universe_list, receiver_size)
        universe_without_receiver_set = set(universe_list) - set(receiver_list)
        sender_list = list(universe_without_receiver_set)[:symmetric_difference_size - 1]
        sender_list.extend(receiver_list[:receiver_size - 1])
        
        del universe_without_receiver_set

    # Clean up memory.
    gc.collect()

    # Dictionary mapping methods to their corresponding IBLT classes.
    iblt_classes = {
        Method.EGH: IBLTWithEGH,
        Method.EXTENDED_HAMMING_CODE: IBLTWithExtendedHamming,
        Method.BCH: IBLTWithBCH
    }

    # Create sender and receiver IBLTs using the selected method.
    sender_iblt = iblt_classes[method](sender_list, len(universe_list))
    receiver_iblt = iblt_classes[method](receiver_list, len(universe_list))

    # For debugging: store sender list in receiver IBLT.
    receiver_iblt.other_list_for_debug = sender_list.copy()

    del sender_list
    del receiver_list
    gc.collect()
        
    return sender_iblt, receiver_iblt

@contextmanager
def get_pool(processes_num):
    """
    Context manager for creating a multiprocessing pool.

    Parameters:
        processes_num (int): The number of processes in the pool.
    
    Yields:
        multiprocessing.Pool: The created pool.
    """
    with multiprocessing.Pool(processes=processes_num, maxtasksperchild=1) as pool:
        yield pool

def export_results_to_csv(header, results, csv_filename: str) -> None:
    """
    Export results to a CSV file.
    
    Parameters:
        header (list): The header row for the CSV file.
        results (list of lists): The results to write to the CSV file.
        csv_filename (str): The name of the CSV file.
    """
    with open(os.path.join("./data", csv_filename), mode='w', newline='') as file:
        writer = csv.writer(file)
        writer.writerow(header)
        writer.writerows(results)

def profile_function(func, *args, **kwargs):
    """
    Profile a function and print its statistics.
    
    Parameters:
        func (callable): The function to profile.
        *args: Positional arguments for the function.
        **kwargs: Keyword arguments for the function.
    
    Returns:
        The result of the function call.
    """
    pr = cProfile.Profile()
    pr.enable()
    result = func(*args, **kwargs)
    pr.disable()
    s = io.StringIO()
    sortby = 'cumulative'
    ps = pstats.Stats(pr, stream=s).sort_stats(sortby)
    ps.print_stats()
    print(s.getvalue())
    return result