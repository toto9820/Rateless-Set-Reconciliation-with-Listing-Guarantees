import csv
import os
os.environ['OPENBLAS_NUM_THREADS'] = '1'
import platform
import random
import math
import matplotlib.pyplot as plt
import numpy as np
import cProfile
import pstats
import io
import gc
import multiprocessing
from contextlib import contextmanager
from numba import jit
from typing import List, Set, Tuple
from functools import partial
from IBLTWithEGH import IBLTWithEGH
from IBLTWithCovArr import IBLTWithCovArr
from Method import Method
from IBLTWithRecursiveArr import IBLTWithRecursiveArr
from IBLTWithExtendedHamming import IBLTWithExtendedHamming
from IBLTWithBCH import IBLTWithBCH
from memory_profiler import profile

# def benchmark_set_reconciliation(symmetric_difference_size: int, 
#                                  method: Method,
#                                  num_trials: int, 
#                                  export_to_csv: bool = False, 
#                                  csv_filename: str = "results.csv",
#                                  set_inside_set: bool = True):
    
#     if set_inside_set:
#         print(f"Receiver set is a super set of sender set for symmetric_difference_size {symmetric_difference_size}")
#     else:
#         print(f"Receiver set is not a super set of sender set symmetric_difference_size {symmetric_difference_size}")
    
#     results = []
#     universe_size_trial_cnt = 1

#     for universe_size in [10**i for i in range(4, 5)]:
#     # for universe_size in [10**i for i in range(2, 6)]:
#         universe_list = list(range(1, universe_size+1)) 
    
#         total_cells_transmitted = 0

#         for trial in range(1, num_trials+1):
#             if set_inside_set:
#                 receiver_list = universe_list
#                 sender_list = random.sample(universe_list, universe_size - symmetric_difference_size)
#             else:
#                 receiver_size = random.randint(1, universe_size - symmetric_difference_size)
#                 receiver_size = max(symmetric_difference_size, receiver_size)
#                 receiver_list = random.sample(universe_list, receiver_size)
                
#                 # print("receiver_list: ", receiver_list)

#                 universe_without_receiver_set = set(universe_list) - set(receiver_list)
                
#                 sender_list = []

#                 sender_list = list(universe_without_receiver_set)[:symmetric_difference_size-1]

#                 sender_list.extend(receiver_list[:(receiver_size-1)])
                
#                 # print("sender_list: ", sender_list)

#             if method == Method.EGH:
#                 sender = IBLTWithEGH(set(sender_list), universe_size)
#                 receiver = IBLTWithEGH(set(receiver_list), universe_size)
#                 receiver.other_set_for_debug = set(sender_list)

#             elif method == Method.BINARY_COVERING_ARRAY:
#                 sender = IBLTWithCovArr(set(sender_list), universe_size)
#                 receiver = IBLTWithCovArr(set(receiver_list), universe_size)

#             elif method == Method.RECURSIVE_ARRAY:
#                 sender = IBLTWithRecursiveArr(set(sender_list), universe_size)
#                 receiver = IBLTWithRecursiveArr(set(receiver_list), universe_size)

#             elif method == Method.EXTENDED_HAMMING_CODE:
#                 sender = IBLTWithExtendedHamming(set(sender_list), universe_size)
#                 receiver = IBLTWithExtendedHamming(set(receiver_list), universe_size)
#                 receiver.other_set_for_debug = set(sender_list)

#             elif method == Method.BCH:
#                 sender = IBLTWithBCH(set(sender_list), universe_size)
#                 receiver = IBLTWithBCH(set(receiver_list), universe_size)
#                 receiver.other_set_for_debug = set(sender_list)
                

#             symmetric_difference = []

#             while True:
#                 sender_cells = []

#                 sender.transmit()

#                 while not sender.cells_queue.empty():
#                     cell = sender.cells_queue.get()

#                     # End of IBLT's cells transmitting.
#                     if cell == "end":
#                         break

#                     sender_cells.append(cell)

#                 symmetric_difference = receiver.receive(sender_cells)

#                 if symmetric_difference:
#                     sender.ack_queue.put("stop")
#                     break

#             total_cells_transmitted += len(receiver.iblt_diff_cells)
#             print(f"Symmetric difference in trail {trial}: {symmetric_difference}")

#         # Optimize memory usage by deleting large temporary objects
#         del universe_list
#         gc.collect()

#         avg_total_cells_transmitted =  math.ceil(total_cells_transmitted / num_trials)
#         print(f"Avg. number of cells transmitted: {avg_total_cells_transmitted:.2f}")
#         results.append((universe_size_trial_cnt, universe_size, avg_total_cells_transmitted))
#         universe_size_trial_cnt += 1

#     if export_to_csv:
#         export_results_to_csv(["Trial", "Universe Size", "Cells Transmitted"],
#                               results, csv_filename)

def run_trial(trial_number: int, universe_size: int, symmetric_difference_size: int, method: Method, set_inside_set: bool) -> int:
    universe_list = list(range(1, universe_size+1)) 

    if set_inside_set:
        receiver_list = universe_list
        sender_list = random.sample(universe_list, universe_size - symmetric_difference_size)
    else:
        receiver_size = max(symmetric_difference_size, random.randint(1, universe_size - symmetric_difference_size))
        receiver_list = random.sample(universe_list, receiver_size)
        universe_without_receiver_set = set(universe_list) - set(receiver_list)
        sender_list = list(universe_without_receiver_set)[:symmetric_difference_size-1]
        sender_list.extend(receiver_list[:(receiver_size-1)])

        del universe_without_receiver_set

    del universe_list
    gc.collect()

    if method == Method.EGH:
        sender = IBLTWithEGH(sender_list, universe_size)
        receiver = IBLTWithEGH(receiver_list, universe_size)
    elif method == Method.BINARY_COVERING_ARRAY:
        sender = IBLTWithCovArr(sender_list, universe_size)
        receiver = IBLTWithCovArr(receiver_list, universe_size)
    elif method == Method.RECURSIVE_ARRAY:
        sender = IBLTWithRecursiveArr(sender_list, universe_size)
        receiver = IBLTWithRecursiveArr(receiver_list, universe_size)
    elif method == Method.EXTENDED_HAMMING_CODE:
        sender = IBLTWithExtendedHamming(sender_list, universe_size)
        receiver = IBLTWithExtendedHamming(receiver_list, universe_size)
    elif method == Method.BCH:
        sender = IBLTWithBCH(sender_list, universe_size)
        receiver = IBLTWithBCH(receiver_list, universe_size)

    receiver.other_list_for_debug = sender_list

    while True:
        sender_cells = []
        sender.transmit()

        while not sender.cells_queue.empty():
            cell = sender.cells_queue.get()
            if cell == "end":
                break
            sender_cells.append(cell)

        symmetric_difference = receiver.receive(sender_cells)

        if symmetric_difference:
            print(f"Trial {trial_number}: Universe size {universe_size}, Symmetric difference {symmetric_difference}, Cells transmitted {len(receiver.iblt_diff_cells)}")
            sender.ack_queue.put("stop")
            break

    return len(receiver.iblt_diff_cells)

@contextmanager
def get_pool(processes_num):
    with multiprocessing.Pool(processes=processes_num, maxtasksperchild=1) as pool:
        yield pool

def benchmark_set_reconciliation(symmetric_difference_size: int, 
                                 method: Method,
                                 num_trials: int, 
                                 export_to_csv: bool = False, 
                                 csv_filename: str = "results.csv",
                                 set_inside_set: bool = True):
    
    print(f"{'Receiver set is' if set_inside_set else 'Receiver set is not'} a super set of sender set for symmetric_difference_size {symmetric_difference_size}")
    
    results = []
    universe_size_trial_cnt = 1

    # for universe_size in [10**i for i in range(1, 7)]:
    for universe_size in [10**i for i in range(2, 8)]:
        total_cells_transmitted = 0
        # trials_to_futures = {}

        # Create a partial function with fixed arguments
        partial_run_trial = partial(run_trial, 
                                    universe_size=universe_size, 
                                    symmetric_difference_size=symmetric_difference_size, 
                                    method=method, 
                                    set_inside_set=set_inside_set)
        
        processes_num = multiprocessing.cpu_count() - 6
        # Use Pool to run trials in parallel
        with get_pool(processes_num) as pool:
            # Use imap_unordered for better performance with large number of items
            for i, cells_transmitted in enumerate(pool.imap_unordered(partial_run_trial, range(1, num_trials + 1))):
                total_cells_transmitted += cells_transmitted
                print(f"Trial {i+1}, Universe size 0^{int(math.log10(universe_size))} completed: {cells_transmitted} cells transmitted")

        avg_total_cells_transmitted = math.ceil(total_cells_transmitted / num_trials)
        print("###############################################################################")
        print(f"Avg. number of cells transmitted: {avg_total_cells_transmitted:.2f}")
        print("###############################################################################")
        results.append((universe_size_trial_cnt, universe_size, avg_total_cells_transmitted))
        universe_size_trial_cnt += 1

    if export_to_csv:
        export_results_to_csv(["Trial", "Universe Size", "Cells Transmitted"],
                              results, csv_filename)

def export_results_to_csv(header, results, csv_filename: str) -> None:
    with open(os.path.join("./data", csv_filename), mode='w', newline='') as file:
        writer = csv.writer(file)
        writer.writerow(header)
        writer.writerows(results)

def measure_decode_success_rate(symmetric_difference_size: int, 
                                max_symmetric_diff_size: int,
                                universe_size: int, 
                                method: Method, 
                                num_trials: int = 1000,
                                set_inside_set: bool = True):
    
    universe_list = list(range(1, universe_size + 1))  

    success_probabilities = [0] * 20 * max_symmetric_diff_size

    for _ in range(num_trials):
        end_trial = False

        if set_inside_set:
            receiver_list = universe_list
            sender_list = random.sample(universe_list, universe_size - symmetric_difference_size)        
        else:
            receiver_size = random.randint(1, universe_size - symmetric_difference_size)
            receiver_size = max(symmetric_difference_size, receiver_size)
            receiver_list = random.sample(universe_list, receiver_size)

            # print("receiver_list: ", receiver_list)

            universe_without_receiver_set = set(universe_list) - set(receiver_list)
            
            sender_list = []

            sender_list = list(universe_without_receiver_set)[:symmetric_difference_size-1]

            sender_list.extend(receiver_list[:(receiver_size-1)])
            
            # print("sender_list: ", sender_list)

        if method == Method.EGH:
            sender = IBLTWithEGH(set(sender_list), universe_size)
            receiver = IBLTWithEGH(set(receiver_list), universe_size)

        elif method == Method.BINARY_COVERING_ARRAY:
            sender = IBLTWithCovArr(set(sender_list), universe_size)
            receiver = IBLTWithCovArr(set(receiver_list), universe_size)

        elif method == Method.RECURSIVE_ARRAY:
            sender = IBLTWithRecursiveArr(set(sender_list), universe_size)
            receiver = IBLTWithRecursiveArr(set(receiver_list), universe_size)

        elif method == Method.EXTENDED_HAMMING_CODE:
            sender = IBLTWithExtendedHamming(set(sender_list), universe_size)
            receiver = IBLTWithExtendedHamming(set(receiver_list), universe_size)

        elif method == Method.BCH:
            sender = IBLTWithBCH(set(sender_list), universe_size)
            receiver = IBLTWithBCH(set(receiver_list), universe_size)

        sender_cells = []
        reciver_cells = []

        num_iterations = 0 

        while True:
            sender.transmit()

            while not sender.cells_queue.empty():
                cell = sender.cells_queue.get()

                # End of IBLT's cells transmitting.
                if cell == "end":
                    break

                sender_cells.append(cell)

            receiver.transmit()

            while not receiver.cells_queue.empty():
                cell = receiver.cells_queue.get()

                # End of IBLT's cells transmitting.
                if cell == "end":
                    break

                reciver_cells.append(cell)

            initial_prob_idx = num_iterations

            for i in range(len(sender_cells)):
                iblt_diffrence = receiver.calc_iblt_diff(sender_cells[:initial_prob_idx+i], reciver_cells[:initial_prob_idx+i])
                symmetric_difference = receiver.listing(iblt_diffrence, with_deocde_frac = True)
                
                if symmetric_difference[0] == "Decode Failure":
                    success_probabilities[initial_prob_idx+len(sender_cells[:i])] += symmetric_difference[1]
                else:
                    success_probabilities[initial_prob_idx+len(sender_cells[:i])] += 1

                num_iterations += 1

                if num_iterations >=  len(success_probabilities):
                    end_trial = True
                    break
            
            if end_trial:
                break

    success_probabilities = [prob / num_trials for prob in success_probabilities]
    return success_probabilities

def plot_success_rate(method, universe_size, symmetric_diff_sizes, num_trials=1000,
                      export_to_csv: bool = False, 
                      csv_dir: str = "results/",
                      set_inside_set=True):

    for symmetric_diff_size in symmetric_diff_sizes:
        success_prob = measure_decode_success_rate(symmetric_diff_size, 
                                                    max(symmetric_diff_sizes),
                                                    universe_size, 
                                                    method, 
                                                    num_trials,
                                                    set_inside_set)
        
        if export_to_csv:
            if set_inside_set:
                csv_filename = f"{csv_dir}/success_probability/{csv_dir}_success_probability_receiver_includes_sender_symmetric_diff_size_{symmetric_diff_size}.csv"
            
            else:
                csv_filename = f"{csv_dir}/success_probability/{csv_dir}_success_probability_receiver_not_includes_sender_symmetric_diff_size_{symmetric_diff_size}.csv"
            export_results_to_csv(["Cells Transmitted", "Success Probability"],
                                list(enumerate(success_prob)), csv_filename)

        plt.plot(range(1, len(success_prob) + 1), success_prob, label=f'|∆|={symmetric_diff_size}')

    plt.xlabel('Number of IBLT cells')
    plt.ylabel('Success Probability')
    plt.title(f'Success Probability vs. Number of cells for {method}')
    plt.legend()
    plt.show(block=True)

def profile_function(func, *args, **kwargs):
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

if __name__ == "__main__":
    # Check the system platform
    system = platform.system()

    if system == 'Linux':
        trials = 100 

    elif system == 'Windows':
        trials = 25 

    universe_size = 1000

    print("IBLT + EGH:")

    # symmetric_difference_size is parameter d.
    # for symmetric_difference_size in [1, 3, 10, 20]:
    for symmetric_difference_size in [3]:

        # benchmark_set_reconciliation(symmetric_difference_size, 
        #                              Method.EGH,
        #                              num_trials=trials, 
        #                              export_to_csv=True, 
        #                              csv_filename=f"egh_results/egh_results_receiver_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
        #                              set_inside_set = True)

        benchmark_set_reconciliation(symmetric_difference_size, 
                                     Method.EGH,
                                     num_trials=trials, 
                                     export_to_csv=False, 
                                     csv_filename=f"egh_results/egh_results_receiver_not_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
                                     set_inside_set = False)
        
        # benchmark_set_reconciliation(symmetric_difference_size, 
        #                              Method.EGH,
        #                              num_trials=trials, 
        #                              export_to_csv=True, 
        #                              csv_filename=f"egh_results/egh_results_receiver_not_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
        #                              set_inside_set = False)
    
    # plot_success_rate(Method.EGH, 
    #                   universe_size=universe_size, symmetric_diff_sizes=[1,3,10,20], 
    #                   num_trials=trials, export_to_csv=True, 
    #                   csv_dir=f"egh_results",
    #                   set_inside_set=True)
        

    # print("IBLT + Binary Covering Arrays:")

    # #for symmetric_difference_size in [1, 3, 10, 20]:
    # for symmetric_difference_size in [10, 20]:
    #     benchmark_set_reconciliation(symmetric_difference_size,
    #                                  Method.BINARY_COVERING_ARRAY, 
    #                                  num_trials=10, 
    #                                  export_to_csv=True, 
    #                                  csv_filename=f"covering_arr_results/covering_arr_results_receiver_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
    #                                  set_inside_set = True)

    #     benchmark_set_reconciliation(symmetric_difference_size, 
    #                                  Method.BINARY_COVERING_ARRAY,
    #                                  num_trials=10, 
    #                                  export_to_csv=True, 
    #                                  csv_filename=f"covering_arr_results/covering_arr_results_receiver_not_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
    #                                  set_inside_set = False)
        
    # print("IBLT + Recurisve Array:")

    # for symmetric_difference_size in [1, 3, 10, 20]:
    #     benchmark_set_reconciliation(symmetric_difference_size,
    #                                  Method.RECURSIVE_ARRAY, 
    #                                  num_trials=10, 
    #                                  export_to_csv=True, 
    #                                  csv_filename=f"recursive_arr_results/recursive_arr_results_receiver_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
    #                                  set_inside_set = True)

    #     benchmark_set_reconciliation(symmetric_difference_size, 
    #                                  Method.RECURSIVE_ARRAY,
    #                                  num_trials=10, 
    #                                  export_to_csv=True, 
    #                                  csv_filename=f"recursive_arr_results/recursive_arr_results_receiver_not_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
    #                                  set_inside_set = False)

    # print("IBLT + Extended Hamming Code:")

    # for symmetric_difference_size in [1,2,3]:
    #     benchmark_set_reconciliation(symmetric_difference_size,
    #                                  Method.EXTENDED_HAMMING_CODE, 
    #                                  num_trials=trials, 
    #                                  export_to_csv=True, 
    #                                  csv_filename=f"extended_hamming_results/extended_hamming_results_receiver_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
    #                                  set_inside_set = True)

    #     benchmark_set_reconciliation(symmetric_difference_size, 
    #                                  Method.EXTENDED_HAMMING_CODE,
    #                                  num_trials=trials, 
    #                                  export_to_csv=True, 
    #                                  csv_filename=f"extended_hamming_results/extended_hamming_results_receiver_not_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
    #                                  set_inside_set = False)

    # plot_success_rate(Method.EXTENDED_HAMMING_CODE, 
    #                 universe_size=universe_size, symmetric_diff_sizes=[1,2,3,4,6,8], 
    #                 num_trials=trials, export_to_csv=True,
    #                 csv_dir=f"extended_hamming_results",
    #                 set_inside_set=False)

    # For now - I can't use with multithreading due to issue with SQL - the 
    # creator thread and using thread are different.
    # print("IBLT + BCH:")

    # symmetric_difference_size is parameter d.
    # for symmetric_difference_size in [1, 3, 10, 20]:
    # for symmetric_difference_size in [3]:

        # benchmark_set_reconciliation(symmetric_difference_size, 
        #                              Method.BCH,
        #                              num_trials=trials, 
        #                              export_to_csv=True, 
        #                              csv_filename=f"bch_results/bch_results_receiver_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
        #                              set_inside_set = True)

        # profile_function(benchmark_set_reconciliation,symmetric_difference_size, 
        #                              Method.BCH,
        #                              num_trials=trials, 
        #                              export_to_csv=True, 
        #                              csv_filename=f"bch_results/bch_results_receiver_not_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
        #                              set_inside_set = False)
        
        # benchmark_set_reconciliation(symmetric_difference_size, 
        #                              Method.BCH,
        #                              num_trials=trials, 
        #                              export_to_csv=False, 
        #                              csv_filename=f"bch_results/bch_results_receiver_not_includes_sender_symmetric_diff_size_{symmetric_difference_size}.csv", 
        #                              set_inside_set = False)
    
    # plot_success_rate(Method.BCH, 
    #                   universe_size=universe_size, symmetric_diff_sizes=[1,5,7,10,15,20], 
    #                   num_trials=trials, export_to_csv=True, 
    #                   csv_dir=f"bch_results",
    #                   set_inside_set=True)

    # snapshot = tracemalloc.take_snapshot()
    # top_stats = snapshot.statistics('lineno')

    # print("[ Top 10 ]")
    # for stat in top_stats[:10]:
    #     print(stat) 
