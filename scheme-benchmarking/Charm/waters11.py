from charm.toolbox.pairinggroup import PairingGroup,GT
from charm.schemes.abenc.abenc_waters09 import CPabe09

group = PairingGroup('SS512')
cpabe = CPabe09(group)
msg = group.random(GT)
(master_secret_key, master_public_key) = cpabe.setup()
policy = '((ONE or THREE) and (TWO or FOUR))'
attr_list = ['THREE', 'ONE', 'TWO']
secret_key = cpabe.keygen(master_public_key, master_secret_key, attr_list)
cipher_text = cpabe.encrypt(master_public_key, msg, policy)
decrypted_msg = cpabe.decrypt(master_public_key, secret_key, cipher_text)
print(decrypted_msg == msg)