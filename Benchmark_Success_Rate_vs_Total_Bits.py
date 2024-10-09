import platform
import matplotlib.pyplot as plt 
from Utils import *
import multiprocessing
from functools import partial

# Bits per IBLT cell (3 fields - count, xorSum, checkSum)
# Each field is 64 bit.
cellSizeInBits = 64 * 3

def run_trial_success_rate(trial_number: int,
                           symmetric_difference_size: int, 
                           universe_size: int, 
                           method: Method, 
                           set_inside_set: bool = True,
                           max_iterations: int = 50):
    
    method_str = str(method).lower().replace('method.', '')

    results = [(0,0,0.0)]
    total_cells = 0
    iteration_count = 0

    p1_iblt, p2_iblt = generate_participants_iblts(universe_size,
                                                   symmetric_difference_size,
                                                   method,
                                                   set_inside_set)
        
    while iteration_count < max_iterations:
        p1_counters, p1_sums, p1_checksums = p1_iblt.encode()

        # Increment total cells by the number of new 
        # cells in this iteration
        total_cells += len(p1_counters)

        symmetric_difference = p2_iblt.decode(p1_counters, p1_sums, p1_checksums)
        
        success_rate = len(symmetric_difference) / symmetric_difference_size
        iteration_count += 1
        results.append((iteration_count, total_cells, success_rate))

        if int(success_rate) == 1:
            break   

    # Clean up
    del p1_iblt
    del p2_iblt
    gc.collect()

    print(f"Trial {trial_number} for method {method_str} completed for symmetric difference size {symmetric_difference_size}")
    # print(f"Trial {trial_number} completed for symmetric difference  {symmetric_difference}")
    return results

def calc_decode_success_rate(symmetric_difference_size: int, 
                             universe_size: int, 
                             method: Method, 
                             num_trials: int = 100,
                             set_inside_set: bool = True,
                             max_iterations: int = 50):
    
    partial_run_trial = partial(run_trial_success_rate, 
                                symmetric_difference_size=symmetric_difference_size,
                                universe_size=universe_size,
                                method=method,
                                set_inside_set=set_inside_set,
                                max_iterations= max_iterations)

    processes_num = int(multiprocessing.cpu_count() * 0.25)

    with get_pool(processes_num) as pool:
        results = list(pool.imap_unordered(partial_run_trial, range(1, num_trials + 1)))

    print("Aggregate results")
    # Aggregate results
    max_cells = max(len(trial) for trial in results)
    aggregated_results = np.zeros(max_cells)
    count_results = np.zeros(max_cells)

    for trial_results in results:
        for idx, cells, success_rate in trial_results: 
            if success_rate == 0.0:
                continue

            aggregated_results[idx] += success_rate
            count_results[idx] = cells

            # Pad the end with ones.
            if int(success_rate) == 1:
                aggregated_results[idx+1:] += success_rate
                break

    # Calculate average cells vs success rates
    avg_success_rates = [(0, 0.0)]

    for idx in range(max_cells):
        if count_results[idx] > 0:
            avg_success_rate = aggregated_results[idx] / num_trials
            avg_success_rates.append((count_results[idx], avg_success_rate))

    return avg_success_rates 

def benchmark_success_rates_vs_total_bits(universe_size, 
                      symmetric_difference_sizes, 
                      num_trials=1000,
                      export_to_csv: bool = False, 
                      set_inside_set=True,
                      max_iterations: int = 50):
    
    # Define a list of marker styles and colors
    # circle, square, diamond, triangle up, triangle down, plus, star, cross
    markers = ['o', 's', 'D', '^', 'v', 'P', '*', 'X']  
    # blue, green, red, cyan, magenta, yellow, black, orange
    colors = ['b', 'g', 'r', 'c', 'm', 'y', 'k', 'orange'] 

    for method in methods: 
        # Clear the current figure and create a new one
        plt.clf()

        for idx, symmetric_difference_size in enumerate(symmetric_difference_sizes):
            results = calc_decode_success_rate(symmetric_difference_size, 
                                                    universe_size, 
                                                    method, 
                                                    num_trials,
                                                    set_inside_set,
                                                    max_iterations)
            
            cells_counts, success_rates = zip(*results)
            
            if export_to_csv:
                if set_inside_set:
                    csv_filename = f"success_rate_vs_total_bits_benchmark/{str(method).lower().replace('method.', '')}_success_rate_vs_total_bits_diff_size_{symmetric_difference_size}_set_inside_set.csv"
                else:
                    csv_filename = f"success_rate_vs_total_bits_benchmark/{str(method).lower().replace('method.', '')}_success_rate_vs_total_bits_diff_size_{symmetric_difference_size}_set_not_inside_set.csv"
                
                export_results_to_csv(["Total Bits Transmitted", "Success Probability"],
                                    results, csv_filename)

            # Plot the results with different markers and colors
            plt.plot(cells_counts, success_rates, 
                    marker=markers[idx % len(markers)],    
                    linestyle='--',                        
                    color=colors[idx % len(colors)],      
                    label=f'|∆|={symmetric_difference_size}')

            plt.xlabel('Number of IBLT cells')
            plt.ylabel('Success Probability')
            plt.title(f'Success Probability vs. Number of cells for {method}')
            plt.legend()

            system = platform.system()

            if system == 'Linux':
                # Save the plot to a file
                plt.savefig(f'./data/success_rate_vs_total_bits_benchmark/{str(method).lower().replace('method.', '')}_success_rate_plot')
            
            elif system == 'Windows':
                # Not working on remote server.
                plt.show(block=True)
            
        # Close the plot to free up memory
        plt.close()  

if __name__ == "__main__":
    # Check the system platform
    system = platform.system()

    if system == 'Linux':
        trials = 50 
        # trials = 5 

    elif system == 'Windows':
        trials = 25 

    universe_size = 10**6

    export_to_csv = True
    set_inside_set = True

    benchmark_success_rates_vs_total_bits(universe_size=universe_size, 
                      symmetric_difference_sizes=[1,3,30,100,300,1000], 
                      num_trials=trials, 
                      export_to_csv=export_to_csv, 
                      set_inside_set=set_inside_set)

    
