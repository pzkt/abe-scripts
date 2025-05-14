from charm.toolbox.pairinggroup import PairingGroup,GT
from charm.schemes.abenc.abenc_maabe_rw15 import MaabeRW15

import timeit

def main():
    group = PairingGroup('SS512')
    maabe = MaabeRW15(group)
    public_parameters = maabe.setup()
    msg = group.random(GT)
    (public_key, secret_key) = maabe.authsetup(public_parameters, "AUTH0")
    gid = "user"

    repeats = 1
    attribute_counts = [1, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50]

    def note(op, attr_nr, time):
        print(op, attr_nr, "Attributes: ", time)

    def setup():
        group = PairingGroup('SS512')
        maabe = MaabeRW15(group)
        public_parameters = maabe.setup()
    
    def authsetup():
        (public_key, secret_key) = maabe.authsetup(public_parameters, "AUTH0")

    def keygen():
        user_keys = maabe.multiple_attributes_keygen(public_parameters, secret_key, gid, attr)
    
    def encrypt():
        cipher_text = maabe.encrypt(public_parameters, public_keys, msg, access_policy)

    def decrypt():
        decrypted_msg = maabe.decrypt(public_parameters, user_keys, cipher_text)

    print("setup benchmark")
    timer = (timeit.timeit(setup = "gc.enable()", stmt = setup, number = repeats))/repeats
    note("setup", None, timer)

    print("auth setup benchmark")
    timer = (timeit.timeit(setup = "gc.enable()", stmt = authsetup, number = repeats))/repeats
    note("auth setup", None, timer)


    print("keygen benchmark")
    for a in attribute_counts:
        attr = [f"ATTRIBUTE{i}@AUTH0" for i in range(a)]
        (public_key, secret_key) = maabe.authsetup(public_parameters, "AUTH0")
        timer = (timeit.timeit(setup = "gc.enable()", stmt = keygen, number = repeats))/repeats
        note("keygen", a, timer)

    print("AND encrypt benchmark")
    for a in attribute_counts:
        (public_key, secret_key) = maabe.authsetup(public_parameters, "AUTH0")
        public_keys = {'AUTH0': public_key}
        access_policy = f"({' and '.join(f'ATTRIBUTE{i}@AUTH0' for i in range(a))})"
        timer = (timeit.timeit(setup = "gc.enable()", stmt = encrypt, number = repeats))/repeats
        note("AND encrypt", a, timer)

    print("OR encrypt benchmark")
    for a in attribute_counts:
        (public_key, secret_key) = maabe.authsetup(public_parameters, "AUTH0")
        public_keys = {'AUTH0': public_key}
        access_policy = f"({' or '.join(f'ATTRIBUTE{i}@AUTH0' for i in range(a))})"
        timer = (timeit.timeit(setup = f"gc.enable()", stmt = encrypt, number = repeats))/repeats
        note("OR encrypt", a, timer)

    print("AND decrypt benchmark")
    for a in attribute_counts:
        access_policy = f"({' and '.join(f'ATTRIBUTE{i}@AUTH0' for i in range(a))})"
        attr = [f"ATTRIBUTE{i}@AUTH0" for i in range(a)]
        (public_key, secret_key) = maabe.authsetup(public_parameters, "AUTH0")
        public_keys = {'AUTH0': public_key}
        cipher_text = maabe.encrypt(public_parameters, public_keys, msg, access_policy)
        user_keys = maabe.multiple_attributes_keygen(public_parameters, secret_key, gid, attr)
        
        timer = (timeit.timeit(setup = "gc.enable()", stmt = decrypt, number = repeats))/repeats
        note("AND decrypt", a, timer)

    print("OR decrypt benchmark")
    for a in attribute_counts:
        access_policy = f"({' or '.join(f'ATTRIBUTE{i}@AUTH0' for i in range(a))})"
        attr = [f"ATTRIBUTE{a-1}"]
        cipher_text = cpabe.encrypt(master_public_key, msg, access_policy)
        secret_key = cpabe.keygen(master_public_key, master_key, attr)

        timer = (timeit.timeit(setup = "gc.enable()", stmt = decrypt, number = repeats))/repeats
        note("OR decrypt", a, timer)

if __name__ == "__main__":
    debug = True
    main()