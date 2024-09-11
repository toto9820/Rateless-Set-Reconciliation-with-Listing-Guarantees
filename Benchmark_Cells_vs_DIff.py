import platform
import math
import gc
import multiprocessing
from Utils import * 
from functools import partial


def run_trial_cells_vs_diff(trial_number: int,
                            universe_size: int,
                            symmetric_diff_size: int,
                            method: Method,
                            set_inside_set: bool = True):
    
    """
    Run a trial for specific method and symmetric difference size.
    """

    p1_iblt, p2_iblt = generate_participants_iblts(universe_size,
                                                   symmetric_diff_size,
                                                   method,
                                                   set_inside_set)
    while True:
        p1_cells = p1_iblt.encode()
        #p1_counters, p1_sums, p1_checksums = p1_iblt.encode()

        # symmetric_difference = p2_iblt.decode(p1_counters, p1_sums, p1_checksums)
        symmetric_difference = p2_iblt.decode(p1_cells)

        if symmetric_difference == ["Decode Failure"]:
            continue

        if len(symmetric_difference) == symmetric_diff_size:
            # print(f"Trial {trial_number}: Universe size 10^{int(math.log10(universe_size))}, Symmetric difference {symmetric_difference}, Cells transmitted {len(p2_iblt.diff_sums)}")
            break
    
    # iblt_diff_cells_size = len(p2_iblt.diff_sums)
    iblt_diff_cells_size = len(p2_iblt.diff_cells)

    # Clean up
    del p1_iblt
    del p2_iblt
    gc.collect()

    return iblt_diff_cells_size

def benchmark_cells_vs_diff(universe_size: int,
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
            
        # for symmetric_diff_size in [10**i for i in 
        #                             range(0, symmetric_diff_trials+1)]:
        for symmetric_diff_size in [100]:
        
            total_cells_transmitted = 0

            # Create a partial function with fixed arguments
            partial_run_trial = partial(run_trial_cells_vs_diff, 
                                        universe_size=universe_size,
                                        symmetric_diff_size=symmetric_diff_size, 
                                        method=method, 
                                        set_inside_set=set_inside_set)
            
            processes_num = int(multiprocessing.cpu_count() * 0.75)
            # processes_num = 1

            # Use Pool to run trials in parallel
            with get_pool(processes_num) as pool:
                # Use imap_unordered for better performance with large number of items
                for i, cells_transmitted in enumerate(pool.imap_unordered(partial_run_trial, range(trials_per_symmetric_diff))):
                    total_cells_transmitted += cells_transmitted
                    print(f"Trial {i+1}, Universe size 10^{int(math.log10(universe_size))} completed: {cells_transmitted} cells transmitted")

            avg_total_cells_transmitted = math.ceil(total_cells_transmitted / trials_per_symmetric_diff)
            print("###############################################################################")
            print(f"Avg. number of cells transmitted: {avg_total_cells_transmitted:.2f} for |∆|={symmetric_diff_size}")
            print("###############################################################################")
            results.append((symmetric_diff_size, avg_total_cells_transmitted))

        if export_to_csv:
            csv_filename = f"{str(method).lower().replace('method.', '')}_results/{str(method).lower().replace('method.', '')}_cells_vs_diff.csv"
            export_results_to_csv(["Symmetric Diff Size", "Cells Transmitted"],
                                results, csv_filename)

if __name__ == "__main__":
    universe_size = 10**6
    symmetric_diff_trials = int(math.log10(universe_size)) 

    # Check the system platform (OS type)
    system = platform.system()

    if system == 'Linux':
        trials_per_symmetric_diff = 100 
    elif system == 'Windows':
        trials_per_symmetric_diff = 10 

    export_to_csv = False
    set_inside_set = True

    # benchmark_cells_vs_diff(universe_size,
    #                         symmetric_diff_trials,
    #                         trials_per_symmetric_diff,
    #                         export_to_csv,
    #                         set_inside_set)

    profile_function(benchmark_cells_vs_diff, universe_size,
                            symmetric_diff_trials,
                            trials_per_symmetric_diff,
                            export_to_csv,
                            set_inside_set)