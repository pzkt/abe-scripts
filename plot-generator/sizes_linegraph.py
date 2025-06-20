import matplotlib.pyplot as plt
import matplotlib.ticker as ticker
import pandas as pd
import numpy as np
import random

def human_readable_bytes(x, pos):
    """Convert bytes to human-readable units."""
    for unit in ['B', 'KiB', 'MiB', 'GiB', 'TiB']:
        if x < 1024:
            return f"{int(x)} {unit}"
        x /= 1024
    return f"{x:.1f} PiB"

path = "../scheme-benchmarking/Results/storage/"
files = ["charm_bsw07_ct","charm_fame_ct","charm_rw15_ct","charm_waters11_ct","charm_yahk14_ct", "circl_tkn20_ct","gofe_fame_ct","openabe_waters11_ct","rabe_bsw07_ct","rabe_fame_ct","rabe_ghw11_ct"]
markers = ["o","s","D","P","^","o","s","D","P","o","s","o","o","s","D"];
lib_colors = ["#003a7d"] * 5 + ["#008dff", "#d83034" , "#c701ff"] + ["#4ecb8d"] * 3

data = []
for file in files:
    try:
        csv_file = pd.read_csv(path + file + ".csv")
        selectRow = csv_file.iloc[2]
        selectRow = selectRow.drop(["index", "attributes"], errors="ignore")
        if (file not in ["rabe_bsw07_ct", "rabe_fame_ct", "rabe_ghw11_ct"]):
            selectRow = selectRow[~selectRow.index.str.startswith("single")]
        selectRow.index = selectRow.index.str.replace(r'^single ', '', regex=True)
        selectRow.index = selectRow.index.str.replace(r'^hybrid ', '', regex=True)
        data.append(selectRow)
    except:
        continue
plt.figure(figsize=(9, 7))

colors = ['#e41a1c','#377eb8','#4daf4a','#984ea3','#ff7f00']

x_positions = np.arange(len(data[0]))
reference_line = 2 ** x_positions

# Plot the reference line
plt.loglog(reference_line, reference_line, 'r--', label=f'plaintext size')

for i in range(len(data)):
    if (i in []): #cull certain entries
        continue
    adjusted_values = data[i].values - reference_line
    adjusted_values = np.where(adjusted_values <= 0, 1e-6, adjusted_values)
    plt.loglog(reference_line, data[i].values, label=files[i] , marker=markers[i], linestyle="-", color=lib_colors[i])

ax = plt.gca()
ax.xaxis.set_major_formatter(ticker.FuncFormatter(human_readable_bytes))
ax.yaxis.set_major_formatter(ticker.FuncFormatter(human_readable_bytes))

plt.legend(fontsize=10)
plt.grid(True, linestyle='--', alpha=0.7)

#plt.title(view)
plt.xlabel('Plaintext Size')
plt.ylabel('Ciphertext Size')

plt.grid(True)

plt.tight_layout()
plt.show()