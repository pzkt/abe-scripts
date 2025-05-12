from charm.toolbox.pairinggroup import PairingGroup,ZR,G1,G2,GT,pair
from charm.schemes.abenc.ac17 import AC17CPABE

group = PairingGroup('SS512')
cpabe = AC17CPABE(group, assump_size=3)
msg = group.random(GT)
attributes = ['ONE', 'TWO', 'THREE']
access_policy = '((four or three) and (three or one))'
(master_public_key, master_key) = cpabe.setup()
secret_key = cpabe.keygen(master_public_key, master_key, attributes)
cipher_text = cpabe.encrypt(master_public_key, msg, access_policy)
decrypted_msg = cpabe.decrypt(master_public_key, cipher_text, secret_key)
print(msg == decrypted_msg)
