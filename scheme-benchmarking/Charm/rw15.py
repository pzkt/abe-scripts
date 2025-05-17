from charm.toolbox.pairinggroup import PairingGroup,GT
from charm.schemes.abenc.abenc_maabe_rw15 import MaabeRW15,merge_dicts

import timeit

def main():
    group = PairingGroup('SS512')
    maabe = MaabeRW15(group)
    public_parameters = maabe.setup()
    msg = group.random(GT)
    (public_key, secret_key) = maabe.authsetup(public_parameters, "AUTH0")
    gid = "user"

    repeats = 10
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
        decrypted_msg = maabe.decrypt(public_parameters, combined_keys, cipher_text)

    def keymerge():
        merged_dicts = merge_dicts(keys1, keys2)

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
        combined_keys = {'GID': gid, 'keys': user_keys}
        timer = (timeit.timeit(setup = "gc.enable()", stmt = decrypt, number = repeats))/repeats
        note("AND decrypt", a, timer)

    print("OR decrypt benchmark")
    for a in attribute_counts:
        access_policy = f"({' or '.join(f'ATTRIBUTE{i}@AUTH0' for i in range(a))})"
        attr = [f"ATTRIBUTE{a-1}@AUTH0"]
        (public_key, secret_key) = maabe.authsetup(public_parameters, "AUTH0")
        public_keys = {'AUTH0': public_key}
        cipher_text = maabe.encrypt(public_parameters, public_keys, msg, access_policy)
        user_keys = maabe.multiple_attributes_keygen(public_parameters, secret_key, gid, attr)
        combined_keys = {'GID': gid, 'keys': user_keys}
        timer = (timeit.timeit(setup = "gc.enable()", stmt = decrypt, number = repeats))/repeats
        note("OR decrypt", a, timer)

    print("Key Merging")
    for a in attribute_counts:
        attr0 = [f"ATTRIBUTE{i}@AUTH0" for i in range(a)]
        attr1 = [f"ATTRIBUTE{i}@AUTH1" for i in range(a)]

        (public_key0, secret_key0) = maabe.authsetup(public_parameters, "AUTH0")
        (public_key1, secret_key1) = maabe.authsetup(public_parameters, "AUTH1")

        keys1 = maabe.multiple_attributes_keygen(public_parameters, secret_key0, gid, attr0)
        keys2 = maabe.multiple_attributes_keygen(public_parameters, secret_key1, gid, attr1)

        timer = (timeit.timeit(setup = "gc.enable()", stmt = keymerge, number = repeats))/repeats
        note("Key Merging", f"{a} x {a}", timer)

    print("Complex AND encrypt benchmark")
    for a in attribute_counts:
        public_key_array = [None] * a
        secret_key_array = [None] * a
        for i in range(a):
            (public_key_array[i], secret_key_array[i]) = maabe.authsetup(public_parameters, f"AUTH{i}")

        public_keys = {f"AUTH{i}": public_key_array[i] for i in range(a)}
        access_policy = f"({' and '.join(f'ATTRIBUTE0@AUTH{i}' for i in range(a))})"
        timer = (timeit.timeit(setup = f"gc.enable()", stmt = encrypt, number = repeats))/repeats
        note("Complex AND encrypt", a, timer)

    print("Complex OR encrypt benchmark")
    for a in attribute_counts:
        public_key_array = [None] * a
        secret_key_array = [None] * a
        for i in range(a):
            (public_key_array[i], secret_key_array[i]) = maabe.authsetup(public_parameters, f"AUTH{i}")

        public_keys = {f"AUTH{i}": public_key_array[i] for i in range(a)}
        access_policy = f"({' or '.join(f'ATTRIBUTE0@AUTH{i}' for i in range(a))})"
        timer = (timeit.timeit(setup = f"gc.enable()", stmt = encrypt, number = repeats))/repeats
        note("Complex OR encrypt", a, timer)

    print("Complex AND decrypt benchmark")
    for a in attribute_counts:
        public_key_array = [None] * a
        secret_key_array = [None] * a
        user_keys = []
        for i in range(a):
            (public_key_array[i], secret_key_array[i]) = maabe.authsetup(public_parameters, f"AUTH{i}")
            user_keys = merge_dicts(user_keys, maabe.multiple_attributes_keygen(public_parameters, secret_key_array[i], gid, [f"ATTRIBUTE0@AUTH{i}"]))

        public_keys = {f"AUTH{i}": public_key_array[i] for i in range(a)}
        access_policy = f"({' and '.join(f'ATTRIBUTE0@AUTH{i}' for i in range(a))})"

        cipher_text = maabe.encrypt(public_parameters, public_keys, msg, access_policy)

        combined_keys = {'GID': gid, 'keys': user_keys}

        timer = (timeit.timeit(setup = f"gc.enable()", stmt = decrypt, number = repeats))/repeats
        note("Complex AND decrypt", a, timer)

    print("Complex OR decrypt benchmark")
    for a in attribute_counts:
        public_key_array = [None] * a
        secret_key_array = [None] * a

        for i in range(a):
            (public_key_array[i], secret_key_array[i]) = maabe.authsetup(public_parameters, f"AUTH{i}")

        user_keys = maabe.multiple_attributes_keygen(public_parameters, secret_key_array[a-1], gid, [f"ATTRIBUTE0@AUTH{a-1}"])

        public_keys = {f"AUTH{i}": public_key_array[i] for i in range(a)}
        access_policy = f"({' or '.join(f'ATTRIBUTE0@AUTH{i}' for i in range(a))})"

        cipher_text = maabe.encrypt(public_parameters, public_keys, msg, access_policy)

        combined_keys = {'GID': gid, 'keys': user_keys}

        timer = (timeit.timeit(setup = f"gc.enable()", stmt = decrypt, number = repeats))/repeats
        note("Complex OR decrypt", a, timer)

if __name__ == "__main__":
    debug = True
    main()