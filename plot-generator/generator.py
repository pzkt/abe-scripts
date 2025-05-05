import matplotlib.pyplot as plt
import pandas as pd

data = pd.read_csv('FAME-encryption-AND.csv')
data2 = pd.read_csv('FAME-encryption-OR.csv')

plt.figure(figsize=(10, 6))  # Set figure size
plt.plot(data['attributes'], data['ms'], label='AND gates' , marker='o', linestyle='-', color='b')
plt.plot(data2['attributes'], data2['ms'], label='OR gates' , marker='o', linestyle='-', color='r')

plt.legend(fontsize=10)
plt.grid(True, linestyle='--', alpha=0.7)

plt.title('GoFE, FAME, Encryption')
plt.xlabel('Number of Attributes')
plt.ylabel('Time [ms]')

plt.grid(True)
plt.show()