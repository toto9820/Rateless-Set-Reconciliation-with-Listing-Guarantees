import platform
import math
import gc
import multiprocessing
from typing import List
from Utils import * 
from functools import partial

def run_trial_additional_cells_vs_diff_size(trial_number: int,
                            universe_size: int,
                            symmetric_difference_sizes: List[int],
                            method: Method,
                            set_inside_set: bool = True):
    
    """
    Run a trial for specific method and symmetric difference size.
    """
    method_str = str(method).lower().replace('method.', '')
    symmetric_difference_sizes = sorted(symmetric_difference_sizes)
    max_symmetric_difference_size = symmetric_difference_sizes[-1]

    p1_iblt, p2_iblt = generate_participants_iblts(universe_size,
                                                   max_symmetric_difference_size,
                                                   method,
                                                   set_inside_set)
    
    cur_symmetric_difference_size = 0
    target_symmetric_difference_size = 0
    cur_additional_cells_per_diff_increase = 0
    results = [(target_symmetric_difference_size, cur_additional_cells_per_diff_increase)]

    for target_symmetric_difference_size in symmetric_difference_sizes:
        prev_additional_cells_per_diff_increase = cur_additional_cells_per_diff_increase
        cur_additional_cells_per_diff_increase = 0.0

        if cur_symmetric_difference_size >= target_symmetric_difference_size:
                results.append((target_symmetric_difference_size, 
                                cur_additional_cells_per_diff_increase))
                continue

        while True:
            p1_counters, p1_sums, p1_checksums = p1_iblt.encode()
            
            symmetric_difference = p2_iblt.decode(p1_counters, 
                                                  p1_sums, 
                                                  p1_checksums)
            
            cur_symmetric_difference_size = len(symmetric_difference)
            
            cur_additional_cells_per_diff_increase = len(p2_iblt.diff_counters) - prev_additional_cells_per_diff_increase
            
            print(f"Trial {trial_number} for method {method_str}, Required Symmetric Difference " 
                  f"Size {target_symmetric_difference_size}, Current Symmetric Difference "
                  f"len: {cur_symmetric_difference_size} with "
                  f"{cur_additional_cells_per_diff_increase} additional cells.")

            if cur_symmetric_difference_size >= target_symmetric_difference_size:
                results.append((target_symmetric_difference_size, 
                                cur_additional_cells_per_diff_increase))
                break

    # Clean up
    del p1_iblt
    del p2_iblt
    gc.collect()

    return results

def benchmark_additional_cells_vs_diff_size(universe_size: int,
                            symmetric_diff_trials: int,
                            trials_per_symmetric_diff: int,
                            export_to_csv: bool = True,
                            set_inside_set: bool = True):
    
    """
    Benchmark the memory required (IBLT cells) based on symmetric 
    difference size.

    Parameters:
        universe_size (int): The size of the universe.
        symmetric_diff_trials (int): The number of symmetric difference trials 
        to run per specific method.
        trials_per_symmetric_diff (int): The number of trials to run per
        symmetric difference size.
        export_to_csv (bool): Flag to export results to a CSV file.
        set_inside_set (bool): Flag indicating whether the receiver list should be a subset of the universe list.
    """
    
    for method in methods:
        results = []
        
        if method == Method.EXTENDED_HAMMING:
            symmetric_difference_sizes = [1,2,3]
        else:
            symmetric_difference_sizes = [10**i for i in range(0, symmetric_diff_trials)]

        # Create a partial function with fixed arguments
        partial_run_trial = partial(run_trial_additional_cells_vs_diff_size, 
                                    universe_size=universe_size,
                                    symmetric_difference_sizes=symmetric_difference_sizes, 
                                    method=method, 
                                    set_inside_set=set_inside_set)
        
        processes_num = int(multiprocessing.cpu_count() * 0.5)

        # Use Pool to run trials in parallel
        with get_pool(processes_num) as pool:
            results = list(pool.imap_unordered(partial_run_trial, range(1, trials_per_symmetric_diff + 1)))

        aggregated_additional_cells = np.zeros(len(symmetric_difference_sizes)+1)
        
        for trial_results in results:
            for idx, (symmetric_diff_size, additional_cells) in enumerate(trial_results): 
                aggregated_additional_cells[idx] += additional_cells

        additional_cells_results = [(0,0)]

        for idx, symmetric_diff_size in enumerate(symmetric_difference_sizes):
            avg_additional_cells = int(math.ceil(aggregated_additional_cells[idx+1] // trials_per_symmetric_diff))
            additional_cells_results.append((symmetric_diff_size, avg_additional_cells))

        if export_to_csv:
            if set_inside_set:
                csv_filename = f"additional_cells_vs_diff_size_benchmark/{str(method).lower().replace('method.', '')}_additional_cells_vs_diff_size_set_inside_set.csv"
            else:
                csv_filename = f"additional_cells_vs_diff_size_benchmark/{str(method).lower().replace('method.', '')}_additional_cells_vs_diff_size_set_not_inside_set.csv"
            
            export_results_to_csv(["Symmetric Diff Size", "Additional Cells Transmitted"],
                                additional_cells_results, csv_filename)

if __name__ == "__main__":
    universe_size = 10**6
    # symmetric_diff_trials = int(math.log10(universe_size)) 
    symmetric_diff_trials = 4

    # Check the system platform (OS type)
    system = platform.system()

    if system == 'Linux':
        trials_per_symmetric_diff = 100
        # trials_per_symmetric_diff = 5

    elif system == 'Windows':
        trials_per_symmetric_diff = 10 

    export_to_csv = True
    set_inside_set = True

    benchmark_additional_cells_vs_diff_size(universe_size,
                            symmetric_diff_trials,
                            trials_per_symmetric_diff,
                            export_to_csv,
                            set_inside_set)

    # profile_function(benchmark_additional_cells_vs_diff_size, universe_size,
    #                         symmetric_diff_trials,
    #                         trials_per_symmetric_diff,
    #                         export_to_csv,
    #                         set_inside_set)