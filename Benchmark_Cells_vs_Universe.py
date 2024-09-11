import math
import platform
from functools import partial
from Utils import *

def benchmark_universe_vs_cells_serial(symmetric_difference_size: int, 
                                       method: Method,
                                       num_trials: int, 
                                       export_to_csv: bool = False, 
                                       csv_filename: str = "results.csv",
                                       set_inside_set: bool = True):
    """
    Benchmark the set reconciliation using IBLT serially.

    Parameters:
        symmetric_difference_size (int): The size of the symmetric difference.
        method (Method): The method to use for creating IBLTs.
        num_trials (int): The number of trials to run.
        export_to_csv (bool): Flag to export results to a CSV file.
        csv_filename (str): The name of the CSV file.
        set_inside_set (bool): Flag indicating whether the receiver list should be a subset of the universe list.
    """
    
    if set_inside_set:
        print(f"Receiver set is a super set of sender set for symmetric_difference_size {symmetric_difference_size}")
    else:
        print(f"Receiver set is not a super set of sender set for symmetric_difference_size {symmetric_difference_size}")
    
    results = []
    universe_size_trial_cnt = 1

    # Iterate over universe sizes
    # for universe_size in [10**i for i in range(2, 8)]:
    for universe_size in [10**i for i in range(6, 7)]:
        total_cells_transmitted = 0

        for trial in range(1, num_trials+1):
            # Declaring that we're using the global universe list variable.
            global universe_list 
            universe_list = range(1, universe_size + 1)

            p1_iblt, p2_iblt = generate_participants_iblts(universe_size,
                                                   symmetric_difference_size,
                                                   method,
                                                   set_inside_set)
                  
            symmetric_difference = []

            # Reconcile the sets
            while True:
                p1_cells = p1_iblt.encode()


                symmetric_difference = p2_iblt.decode(p1_cells)


                if symmetric_difference:
                    break

            total_cells_transmitted += len(p2_iblt.diff_cells)
            # print(f"Symmetric difference in trial {trial}: {symmetric_difference}")

        # Optimize memory usage by deleting large temporary objects
        del universe_list
        gc.collect()

        avg_total_cells_transmitted =  math.ceil(total_cells_transmitted / num_trials)
        print(f"Avg. number of cells transmitted: {avg_total_cells_transmitted:.2f}")
        results.append((universe_size_trial_cnt, universe_size, avg_total_cells_transmitted))
        universe_size_trial_cnt += 1

    if export_to_csv:
        export_results_to_csv(["Trial", "Universe Size", "Cells Transmitted"],
                              results, csv_filename)

def run_trial_cells_vs_universe(trial_number: int, universe_size: int, symmetric_difference_size: int, method: Method, set_inside_set: bool) -> int:
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
    # Declaring that we're using the global universe list variable.
    global universe_list 

    sender_iblt, receiver_iblt = generate_participants_iblts(symmetric_difference_size,
                                                                method,
                                                                set_inside_set)
            
    while True:
        sender_cells = []
        sender_iblt.encode()

        while not sender_iblt.cells_queue.empty():
            cell = sender_iblt.cells_queue.get()
            if cell == "end":
                break
            sender_cells.append(cell)

        symmetric_difference = receiver_iblt.receive(sender_cells)

        if symmetric_difference:
            print(f"Trial {trial_number}: Universe size 10^{int(math.log10(universe_size))}, Symmetric difference {symmetric_difference}, Cells transmitted {len(receiver_iblt.iblt_diff_cells)}")
            sender_iblt.ack_queue.put("stop")
            break

    iblt_diff_cells_size = len(receiver_iblt.iblt_diff_cells)

    # Clean up
    del sender_iblt
    del receiver_iblt
    gc.collect()

    return iblt_diff_cells_size

def benchmark_universe_vs_cells_parallel(symmetric_difference_size: int, 
                                         method: Method,
                                         num_trials: int, 
                                         export_to_csv: bool = False, 
                                         csv_filename: str = "results.csv",
                                         set_inside_set: bool = True):
    """
    Benchmark the set reconciliation using IBLT in parallel.

    Parameters:
        symmetric_difference_size (int): The size of the symmetric difference.
        method (Method): The method to use for creating IBLTs.
        num_trials (int): The number of trials to run.
        export_to_csv (bool): Flag to export results to a CSV file.
        csv_filename (str): The name of the CSV file.
        set_inside_set (bool): Flag indicating whether the receiver list should be a subset of the universe list.
    """
    
    print(f"{'Receiver set is' if set_inside_set else 'Receiver set is not'} a super set of sender set for symmetric_difference_size {symmetric_difference_size}")
    
    results = []
    universe_size_trial_cnt = 1

    # Iterate over universe sizes
    for universe_size in (10**i for i in range(2, 8)):
        # Declaring that we're using the global universe list variable.
        global universe_list 
        universe_list = range(1, universe_size + 1)
        
        total_cells_transmitted = 0

        # Create a partial function with fixed arguments
        partial_run_trial = partial(run_trial_cells_vs_universe, 
                                    universe_size=universe_size, 
                                    symmetric_difference_size=symmetric_difference_size, 
                                    method=method, 
                                    set_inside_set=set_inside_set)
        
        processes_num = int(multiprocessing.cpu_count() * 0.5)
        # Use Pool to run trials in parallel
        with get_pool(processes_num) as pool:
            # Use imap_unordered for better performance with large number of items
            for i, cells_transmitted in enumerate(pool.imap_unordered(partial_run_trial, range(1, num_trials + 1))):
                total_cells_transmitted += cells_transmitted
                print(f"Trial {i+1}, Universe size 10^{int(math.log10(universe_size))} completed: {cells_transmitted} cells transmitted")

        avg_total_cells_transmitted = math.ceil(total_cells_transmitted / num_trials)
        print("###############################################################################")
        print(f"Avg. number of cells transmitted: {avg_total_cells_transmitted:.2f}")
        print("###############################################################################")
        results.append((universe_size_trial_cnt, universe_size, avg_total_cells_transmitted))
        universe_size_trial_cnt += 1

    if export_to_csv:
        export_results_to_csv(["Trial", "Universe Size", "Cells Transmitted"],
                              results, csv_filename)
        
if __name__ == "__main__":
    # Check the system platform
    system = platform.system()

    if system == 'Linux':
        trials = 100 
        # trials = 20 

    elif system == 'Windows':
        trials = 10 

    universe_size = 10**6

    # print("IBLT + EGH:")

    # symmetric_difference_size is parameter d.

    # for symmetric_difference_size in [1, 3, 10, 20]:
    for symmetric_difference_size in [150]:

        # benchmark_universe_vs_cells_parallel(symmetric_difference_size, 
        #                                     Method.EGH,
        #                                     num_trials=trials, 
        #                                     export_to_csv=True, 
        #                                     csv_filename=f"egh_results/egh_results_receiver_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
        #                                     set_inside_set = True)

        profile_function(benchmark_universe_vs_cells_serial, symmetric_difference_size, 
                        Method.EGH,
                        num_trials=trials, 
                        export_to_csv=False, 
                        csv_filename=f"egh_results/egh_results_receiver_not_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
                        set_inside_set = True)

        # benchmark_universe_vs_cells_parallel(symmetric_difference_size, 
        #                                     Method.EGH,
        #                                     num_trials=trials, 
        #                                     export_to_csv=True, 
        #                                     csv_filename=f"egh_results/egh_results_receiver_not_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
        #                                     set_inside_set = False)

    # print("IBLT + Extended Hamming Code:")

    # for symmetric_difference_size in [1,2,3]:
    #     benchmark_universe_vs_cells_parallel(symmetric_difference_size,
    #                                         Method.EXTENDED_HAMMING_CODE, 
    #                                         num_trials=trials, 
    #                                         export_to_csv=True, 
    #                                         csv_filename=f"extended_hamming_results/extended_hamming_results_receiver_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
    #                                         set_inside_set = True)

    #     benchmark_universe_vs_cells_parallel(symmetric_difference_size, 
    #                                         Method.EXTENDED_HAMMING_CODE,
    #                                         num_trials=trials, 
    #                                         export_to_csv=True, 
    #                                         csv_filename=f"extended_hamming_results/extended_hamming_results_receiver_not_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
    #                                         set_inside_set = False)

    # print("IBLT + BCH:")

    # for symmetric_difference_size in [1, 3, 10, 20]:
    #     benchmark_universe_vs_cells_parallel(symmetric_difference_size, 
    #                                         Method.BCH,
    #                                         num_trials=trials, 
    #                                         export_to_csv=True, 
    #                                         csv_filename=f"bch_results/bch_results_receiver_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
    #                                         set_inside_set = True)

        # profile_function(benchmark_universe_vs_cells_serial, symmetric_difference_size, 
        #                 Method.BCH,
        #                 num_trials=trials, 
        #                 export_to_csv=False, 
        #                 csv_filename=f"bch_results/bch_results_receiver_not_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
        #                 set_inside_set = True)

        # benchmark_universe_vs_cells_parallel(symmetric_difference_size, 
        #                                     Method.BCH,
        #                                     num_trials=trials, 
        #                                     export_to_csv=True, 
        #                                     csv_filename=f"bch_results/bch_results_receiver_not_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
        #                                     set_inside_set = False)
        
    print("IBLT + ID:")
    
    # for symmetric_difference_size in [1, 3, 10, 20]:
    # for symmetric_difference_size in [3]:
        # benchmark_universe_vs_cells_parallel(symmetric_difference_size, 
        #                                     Method.IDM,
        #                                     num_trials=trials, 
        #                                     export_to_csv=True, 
        #                                     csv_filename=f"idm_results/idm_results_receiver_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
        #                                     set_inside_set = True)

        # profile_function(benchmark_universe_vs_cells_serial, symmetric_difference_size, 
        #                 Method.IDM,
        #                 num_trials=trials, 
        #                 export_to_csv=False, 
        #                 csv_filename=f"idm_results/idm_results_receiver_not_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
        #                 set_inside_set = True)

        # benchmark_universe_vs_cells_parallel(symmetric_difference_size, 
        #                                     Method.IDM,
        #                                     num_trials=trials, 
        #                                     export_to_csv=True, 
        #                                     csv_filename=f"idm_results/idm_results_receiver_not_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
        #                                     set_inside_set = False)
        
            
