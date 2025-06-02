from charm.toolbox.symcrypto import AuthenticatedCryptoAbstraction
from charm.core.engine.util import bytesToObject, objectToBytes
from charm.toolbox.conversion import Conversion
from charm.core.engine.util import objectToBytes

import pickle
import json
import csv
import os

def update_csv(file_name: str, index: str, column: str, value: str):
    records: List[List[str]] = []
    headers: List[str] = []
    file_exists = True

    try:
        with open(file_name, 'r', newline='') as file:
            reader = csv.reader(file)
            records = list(reader)
            if records:
                headers = records[0]
    except FileNotFoundError:
        file_exists = False
    except Exception as e:
        return f"error opening file: {e}"

    if not records:
        headers = ["attributes"]
        records = [headers]

    try:
        column_index = headers.index(column)
    except ValueError:
        headers.append(column)
        column_index = len(headers) - 1
        records[0] = headers

        for row in records[1:]:
            while len(row) <= column_index:
                row.append("")

    row_index = -1
    for i, record in enumerate(records):
        if i == 0:
            continue
        if record and record[0] == index:
            row_index = i
            break

    if row_index == -1:
        new_row = [index] + [""] * (len(headers) - 1)
        records.append(new_row)
        row_index = len(records) - 1

    while len(records[row_index]) <= column_index:
        records[row_index].append("")

    records[row_index][column_index] = value

    try:
        with open(file_name, 'w', newline='') as file:
            writer = csv.writer(file)
            writer.writerows(records)
    except Exception as e:
        return f"error writing CSV: {e}"

    return None

def encrypt_aes(key: bytes, plaintext: str) -> bytes:

    if len(key) not in {16, 24, 32}:
        raise ValueError("Key must be 16, 24, or 32 bytes long")

    auth_enc = AuthenticatedCryptoAbstraction(key)
    ciphertext = auth_enc.encrypt(plaintext.encode('utf-8'))

    return ciphertext

def serialize_ct(ct, group):
    serializable_ct = {}
    for key, value in ct.items():
        if isinstance(value, dict):
            serializable_ct[key] = {
                k: group.serialize(v) for k, v in value.items()
            }
        elif isinstance(value, list):
            serializable_ct[key] = [group.serialize(v) for v in value]
        else:
            try:
                serializable_ct[key] = group.serialize(value)
            except:
                serializable_ct[key] = value
    return serializable_ct