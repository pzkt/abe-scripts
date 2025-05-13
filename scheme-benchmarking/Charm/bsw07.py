from charm.toolbox.pairinggroup import PairingGroup,ZR,G1,G2,GT,pair
from charm.schemes.abenc.abenc_bsw07 import CPabe_BSW07

import timeit

def main():
    group = PairingGroup('SS512')
    cpabe = CPabe_BSW07(group)
    (master_public_key, master_key) = cpabe.setup()
    msg = group.random(GT)

    repeats = 1
    attribute_counts = [1, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50]

    def note(op, attr_nr, time):
        print(op, attr_nr, "Attributes: ", time)

    def setup():
        group = PairingGroup('SS512')
        cpabe = CPabe_BSW07(group)
        (master_public_key, master_key) = cpabe.setup()

    def keygen():
        secret_key = cpabe.keygen(master_public_key, master_key, attr)
    
    def encrypt():
        cipher_text = cpabe.encrypt(master_public_key, msg, access_policy)

    def decrypt():
        decrypted_msg = cpabe.decrypt(master_public_key, secret_key, cipher_text)

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