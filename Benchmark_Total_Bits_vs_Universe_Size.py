import math
import platform
from functools import partial
from typing import List
from Utils import *

# Bits per IBLT cell (3 fields - count, xorSum, checkSum)
# Each field is 64 bit.
cellSizeInBits = 64 * 3

def run_trial_total_cells_vs_universe_size(trial_number: int, 
                                      universe_size: int, 
                                      symmetric_difference_size: int, 
                                      method: Method, 
                                      set_inside_set: bool) -> int:
    """
    Run a single trial of the set reconciliation.

    Parameters:
        trial_number (int): The trial number.
        universe_size (int): The size of the universe.
        symmetric_difference_size (int): The size of the symmetric difference.
        method (Method): The method to use for creating IBLTs.
        set_inside_set (bool): Flag indicating whether the receiver list should be a subset of the universe list.
    
    Returns:
        int: The number of cells transmitted.
    """
    method_str = str(method).lower().replace('method.', '')

    p1_iblt, p2_iblt = generate_participants_iblts(universe_size,
                                                   symmetric_difference_size,
                                                   method,
                                                   set_inside_set)
            
    while True:
        p1_counters, p1_sums, p1_checksums = p1_iblt.encode()

        symmetric_difference = p2_iblt.decode(p1_counters, p1_sums, p1_checksums)

        print(f"Trial {trial_number} for method {method_str}: "
              f"Universe size 10^{int(math.log10(universe_size))}, "
              f"Required Symmetric Difference " 
              f"Size {symmetric_difference_size} ,"
              f"Symmetric difference Size {len(symmetric_difference)}, "
              f"Cells transmitted {len(p2_iblt.diff_counters)}.")

        if len(symmetric_difference) == symmetric_difference_size:
            break

    iblt_diff_cells_size = len(p2_iblt.diff_counters)

    # Clean up
    del p1_iblt
    del p2_iblt
    gc.collect()

    return iblt_diff_cells_size

def benchmark_total_bits_vs_universe_size(universe_size: int,
                                         universe_trials: int, 
                                         trials_per_universe_size: int,
                                         export_to_csv: bool = True, 
                                         set_inside_set: bool = True):
    """
    Benchmark the set reconciliation using IBLT in parallel.

    Parameters:
        symmetric_difference_size (int): The size of the symmetric difference.
        num_trials (int): The number of trials to run.
        export_to_csv (bool): Flag to export results to a CSV file.
        set_inside_set (bool): Flag indicating whether the receiver list should be a subset of the universe list.
    """    
    results = []

    for method in methods:
        method_str = str(method).lower().replace('method.', '')

        symmetric_difference_sizes=[1,3,30,90]
        universe_trials_start = 3

        if method == Method.EXTENDED_HAMMING:
            symmetric_difference_sizes=[1,2,3]
        elif method == Method.OLS:
            universe_trials_start = 4
            # universe_trials -= 1
        else:
            universe_trials_start = 3

        for symmetric_difference_size in symmetric_difference_sizes:
            results = []

            # Iterate over universe sizes
            for universe_size in (10**i for i in range(universe_trials_start, universe_trials)):            
                total_cells_transmitted = 0

                # Create a partial function with fixed arguments
                partial_run_trial = partial(run_trial_total_cells_vs_universe_size, 
                                            universe_size=universe_size, 
                                            symmetric_difference_size=symmetric_difference_size, 
                                            method=method, 
                                            set_inside_set=set_inside_set)
                
                processes_num = int(multiprocessing.cpu_count() * 0.25)
                # Use Pool to run trials in parallel
                with get_pool(1) as pool:
                    # Use imap_unordered for better performance with large number of items
                    for i, cells_transmitted in enumerate(pool.imap_unordered(partial_run_trial, range(1, trials_per_universe_size + 1))):
                        total_cells_transmitted += cells_transmitted
                        print(f"Trial {i+1} for method {method_str}, Universe size 10^{int(math.log10(universe_size))} completed: {cells_transmitted} cells transmitted")

                avg_total_cells_transmitted = math.ceil(total_cells_transmitted / trials_per_universe_size)
                print("###############################################################################")
                print(f"Avg. number of cells transmitted for method {method_str}: {avg_total_cells_transmitted:.2f}")
                print("###############################################################################")
                avg_total_bits_transmitted = avg_total_cells_transmitted * cellSizeInBits
                results.append((universe_size, avg_total_bits_transmitted))

            if export_to_csv:
                if set_inside_set:
                    csv_filename = f"total_bits_vs_universe_size_benchmark/{str(method).lower().replace('method.', '')}_total_bits_vs_universe_size_for_diff_size_{symmetric_difference_size}_set_inside_set.csv"
                else:
                    csv_filename = f"total_bits_vs_universe_size_benchmark/{str(method).lower().replace('method.', '')}_total_bits_vs_universe_size_for_diff_size_{symmetric_difference_size}_set_not_inside_set.csv"
                    
                export_results_to_csv(["Universe Size", "Total Bits Transmitted"],
                                    results, csv_filename)
        
if __name__ == "__main__":
    # Check the system platform
    system = platform.system()

    if system == 'Linux':
        trials_per_universe_size = 50 

    elif system == 'Windows':
        trials_per_universe_size = 10 

    universe_size = 10**7
    universe_trials = int(math.log10(universe_size)) + 1

    export_to_csv = True
    set_inside_set = True

    benchmark_total_bits_vs_universe_size(universe_size,
                            universe_trials,
                            trials_per_universe_size,
                            export_to_csv=export_to_csv,
                            set_inside_set=set_inside_set)
    

    # profile_function(benchmark_total_bits_vs_universe_size, universe_size,
    #                         universe_trials,
    #                         trials_per_universe_size,
    #                         export_to_csv=export_to_csv,
    #                         set_inside_set=set_inside_set)
