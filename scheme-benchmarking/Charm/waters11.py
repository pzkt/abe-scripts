from charm.toolbox.pairinggroup import PairingGroup,GT
from charm.schemes.abenc.abenc_waters09 import CPabe09
from charm.adapters.abenc_adapt_hybrid import HybridABEnc
from charm.core.engine.util import bytesToObject, objectToBytes
from charm.toolbox.conversion import Conversion
from hashlib import sha256
import pickle
from shared import *

import timeit
import sys

def main():
    group = PairingGroup('SS512')
    cpabe = CPabe09(group)
    (master_key, master_public_key) = cpabe.setup()
    msg = group.random(GT)

    repeats = 10
    attribute_counts = [1, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50]

    def note(op, attr_nr, time):
        update_csv("charm_waters11.csv", str(attr_nr), str(op), str(time))
        print(op, attr_nr, "Attributes: ", time)

    def setup():
        group = PairingGroup('SS512')
        cpabe = CPabe09(group)
        (master_key, master_public_key) = cpabe.setup()

    def keygen():
        secret_key = cpabe.keygen(master_public_key, master_key, attr)
    
    def encrypt():
        cipher_text = cpabe.encrypt(master_public_key, msg, access_policy)

    def decrypt():
        decrypted_msg = cpabe.decrypt(master_public_key, secret_key, cipher_text)

    print("AND cipher size benchmark")
    for size in range(25):
        hyb_abe = HybridABEnc(cpabe, group)
        (mk, pk) = hyb_abe.setup()
        
        content = os.urandom(1<<size)
        for a in attribute_counts:
            access_policy = f"({' and '.join(f'ATTRIBUTE{i}' for i in range(a))})"
            ct = str(hyb_abe.encrypt(pk, content, access_policy))
            update_csv("charm_waters11_ct.csv", str(a), "hybrid " + str(1<<size), str(len(ct)))

    print("setup benchmark")
    timer = (timeit.timeit(setup = "gc.enable()", stmt = setup, number = repeats))/repeats
    note("setup", None, timer)

    print("keygen benchmark")
    for a in attribute_counts:
        attr = [f"ATTRIBUTE{i}" for i in range(a)]
        timer = (timeit.timeit(setup = "gc.enable()", stmt = keygen, number = repeats))/repeats
        note("keygen", a, timer)

    print("AND encrypt benchmark")
    for a in attribute_counts:
        access_policy = f"({' and '.join(f'ATTRIBUTE{i}' for i in range(a))})"
        timer = (timeit.timeit(setup = "gc.enable()", stmt = encrypt, number = repeats))/repeats
        note("AND encrypt", a, timer)

    print("OR encrypt benchmark")
    for a in attribute_counts:
        access_policy = f"({' or '.join(f'ATTRIBUTE{i}' for i in range(a))})"
        timer = (timeit.timeit(setup = f"gc.enable()", stmt = encrypt, number = repeats))/repeats
        note("OR encrypt", a, timer)

    print("AND decrypt benchmark")
    for a in attribute_counts:
        access_policy = f"({' and '.join(f'ATTRIBUTE{i}' for i in range(a))})"
        attr = [f"ATTRIBUTE{i}" for i in range(a)]
        cipher_text = cpabe.encrypt(master_public_key, msg, access_policy)
        secret_key = cpabe.keygen(master_public_key, master_key, attr)

        timer = (timeit.timeit(setup = "gc.enable()", stmt = decrypt, number = repeats))/repeats
        note("AND decrypt", a, timer)

    print("OR decrypt benchmark")
    for a in attribute_counts:
        access_policy = f"({' or '.join(f'ATTRIBUTE{i}' for i in range(a))})"
        attr = [f"ATTRIBUTE{a-1}"]
        cipher_text = cpabe.encrypt(master_public_key, msg, access_policy)
        secret_key = cpabe.keygen(master_public_key, master_key, attr)

        timer = (timeit.timeit(setup = "gc.enable()", stmt = decrypt, number = repeats))/repeats
        note("OR decrypt", a, timer)

if __name__ == "__main__":
    debug = True
    main()