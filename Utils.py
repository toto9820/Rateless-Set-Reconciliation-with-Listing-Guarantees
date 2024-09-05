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

# method (Method): The method to use for creating IBLTs.
# methods = [Method.EGH, 
#            Method.OLS, 
#            Method.ID, 
#            Method.EXTENDED_HAMMING_CODE]

methods = [Method.EGH]

# Global universe list shared between pool of processes to save memory.

def generate_participants_iblts(universe_size, symmetric_difference_size, method, set_inside_set):
    """
    Generate participants lists based on the given parameters.
    
    Parameters:
        symmetric_difference_size (int): The size of the symmetric difference.
        method (Method): The method to use for creating IBLTs.
        set_inside_set (bool): Flag indicating whether a superset assumption holds, i.e. one participant's set
        includes the other.
    
    Returns:
        tuple: A tuple containing the participants' IBLTs.
    """
    # Declaring that we're using the global universe list variable.
    universe_list = range(1, universe_size + 1)

    if set_inside_set:
        # Create list of participant 2 from the universe list.
        p2_list = universe_list
        # Randomly select elements to remove from the universe list 
        # to create the list of participant 1.
        elements_to_remove = random.sample(universe_list, symmetric_difference_size)
        remove_set = set(elements_to_remove)
        p1_list = [elem for elem in universe_list if elem not in remove_set]
        
        del elements_to_remove
        del remove_set
        
    else:
        # Randomly determine the size of participant 2 list and create it.
        p2_size = max(symmetric_difference_size, random.randint(1, len(universe_list) - symmetric_difference_size))
        p2_list = random.sample(universe_list, p2_size)
        universe_without_p2_set = set(universe_list) - set(p2_list)
        p1_list = list(universe_without_p2_set)[:symmetric_difference_size - 1]
        p1_list.extend(p2_list[:p2_size - 1])
        
        del universe_without_p2_set

    # Clean up memory.
    gc.collect()

    # Dictionary mapping methods to their corresponding IBLT classes.
    iblt_classes = {
        Method.EGH: IBLTWithEGH,
        # Method.OLS: IBLTWithOLS,
        # Method.ID: IBLTWithID,
        Method.EXTENDED_HAMMING_CODE: IBLTWithExtendedHamming
    }

    # Create sender and receiver IBLTs using the selected method.
    p1_iblt = iblt_classes[method](p1_list, len(universe_list))
    p2_iblt = iblt_classes[method](p2_list, len(universe_list))

    # For debugging: store particpant 1's list in 
    # participant 2's IBLT object.
    p2_iblt.other_list_for_debug = p1_iblt

    del p1_list
    del p2_list
    gc.collect()
        
    return p1_iblt, p2_iblt

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