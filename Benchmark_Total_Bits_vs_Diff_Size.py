import platform
import math
import gc
import multiprocessing
from Utils import * 
from functools import partial

# Bits per IBLT cell (3 fields - count, xorSum, checkSum)
# Each field is 64 bit.
cellSizeInBits = 64 * 3

def run_trial_total_cells_vs_diff_size(trial_number: int,
                            universe_size: int,
                            symmetric_difference_size: int,
                            method: Method,
                            set_inside_set: bool = True):
    
    """
    Run a trial for specific method and symmetric difference size.
    """

    method_str = str(method).lower().replace('method.', '')


    p1_iblt, p2_iblt = generate_participants_iblts(universe_size,
                                                   symmetric_difference_size,
                                                   method,
                                                   set_inside_set)
    while True:
        p1_counters, p1_sums, p1_checksums = p1_iblt.encode()

        symmetric_difference = p2_iblt.decode(p1_counters, p1_sums, p1_checksums)
        
        print(F"Trial {trial_number} for method {method_str}, Symmetric Difference Tested Size {symmetric_difference_size}, Symmetric Difference len: {len(symmetric_difference)} with {len(p2_iblt.diff_counters)} cells")

        if len(symmetric_difference) == symmetric_difference_size:
            break
    
    iblt_diff_cells_size = len(p2_iblt.diff_counters)

    # Clean up
    del p1_iblt
    del p2_iblt
    gc.collect()

    return iblt_diff_cells_size

def benchmark_total_bits_vs_diff_size(universe_size: int,
                            symmetric_difference_size: int,
                            trials_per_symmetric_diff: int,
                            export_to_csv: bool = True,
                            set_inside_set: bool = True):
    
    """
    Benchmark the memory required (IBLT cells in bits) based on symmetric 
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
            
        for symmetric_difference_size in symmetric_difference_sizes:
        
            total_cells_transmitted = 0

            # Create a partial function with fixed arguments
            partial_run_trial = partial(run_trial_total_cells_vs_diff_size, 
                                        universe_size=universe_size,
                                        symmetric_difference_size=symmetric_difference_size, 
                                        method=method, 
                                        set_inside_set=set_inside_set)
            
            processes_num = int(multiprocessing.cpu_count() * 0.25)

            # Use Pool to run trials in parallel
            with get_pool(processes_num) as pool:
                # Use imap_unordered for better performance with large number of items
                for i, cells_transmitted in enumerate(pool.imap_unordered(partial_run_trial, range(trials_per_symmetric_diff))):
                    total_cells_transmitted += cells_transmitted
                    print(f"Trial {i+1}, Universe size 10^{int(math.log10(universe_size))} completed: {cells_transmitted} cells transmitted")

            avg_total_cells_transmitted = math.ceil(total_cells_transmitted / trials_per_symmetric_diff)
            print("###############################################################################")
            print(f"Avg. number of cells transmitted: {avg_total_cells_transmitted:.2f} for |∆|={symmetric_difference_size}")
            print("###############################################################################")
            avg_total_bits_transmitted = avg_total_cells_transmitted * cellSizeInBits
            results.append((symmetric_difference_size, avg_total_bits_transmitted))

        if export_to_csv:
            if set_inside_set:
                csv_filename = f"total_bits_vs_diff_size_benchmark/{str(method).lower().replace('method.', '')}_total_bits_vs_diff_size_set_inside_set.csv"
            else:
                csv_filename = f"total_bits_vs_diff_size_benchmark/{str(method).lower().replace('method.', '')}_total_bits_vs_diff_size_set_not_inside_set.csv"
        
            export_results_to_csv(["Symmetric Diff Size", "Total Bits Transmitted"],
                                results, csv_filename)

if __name__ == "__main__":
    universe_size = 10**6
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

    benchmark_total_bits_vs_diff_size(universe_size,
                            symmetric_diff_trials,
                            trials_per_symmetric_diff,
                            export_to_csv,
                            set_inside_set)

    # profile_function(benchmark_total_bits_vs_diff_size, universe_size,
    #                         symmetric_diff_trials,
    #                         trials_per_symmetric_diff,
    #                         export_to_csv,
    #                         set_inside_set)