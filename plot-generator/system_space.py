import matplotlib.pyplot as plt
import pandas as pd
import random

path = "../abe-scheme/results/system-space/variable_policy_space_OR.csv"
path2 = "../abe-scheme/results/system-space/variable_policy_space_AND.csv"
path3 = "../abe-scheme/results/system-space/variable_policy_space_BOTH.csv"
markers = ["o","^","s"];

colors1 = ["#41b6c4", "#225ea8" , "#081d58"]
colors2 = ["#fd8d3c", "#e31a1c" , "#800026"]
colors3 = ["#78c679", "#238443" , "#004529"]

data = pd.read_csv(path)
data2 = pd.read_csv(path2)
data3 = pd.read_csv(path3)

plt.figure(figsize=(7, 5))

rate = 1_000_000

data['small'] /= rate
data['medium'] /= rate
data['large'] /= rate

data2['small'] /= rate
data2['medium'] /= rate
data2['large'] /= rate

data3['small'] /= rate
data3['medium'] /= rate
data3['large'] /= rate

#plt.axhline(y=0.443, color='r', linestyle='--', label='plaintext size (small)')

plt.axhline(y=39.830, color='b', linestyle='--', label='plaintext size (medium)')


#plt.plot(data['attributes'], data['small'], label="small (OR)" , marker=markers[0], linestyle="-", color=colors1[0])
#plt.plot(data['attributes'], data['medium'], label="medium (OR)" , marker=markers[1], linestyle="-", color=colors1[1])
plt.plot(data['attributes'], data['large'], label="large (OR)" , marker=markers[0], linestyle="-", color=colors1[0])

#plt.plot(data2['attributes'], data2['small'], label="small (AND)" , marker=markers[0], linestyle="-", color=colors2[0])
#plt.plot(data2['attributes'], data2['medium'], label="medium (AND)" , marker=markers[1], linestyle="-", color=colors2[1])
plt.plot(data2['attributes'], data2['large'], label="large (AND)" , marker=markers[1], linestyle="-", color=colors2[0])

#plt.plot(data3['attributes'], data3['small'], label="small (BOTH)" , marker=markers[0], linestyle="-", color=colors3[0])
#plt.plot(data3['attributes'], data3['medium'], label="medium (BOTH)" , marker=markers[1], linestyle="-", color=colors3[1])
plt.plot(data3['attributes'], data3['large'], label="large (BOTH)" , marker=markers[2], linestyle="-", color=colors3[0])

plt.legend(fontsize=10)
plt.grid(True, linestyle='--', alpha=0.7)

#plt.title(view)
plt.xlabel('Number of Attributes')
plt.ylabel('Entry Size [MB]')

plt.grid(True)

plt.tight_layout()
plt.show()