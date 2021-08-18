import argparse as ap
import string
import random
import base64
import hashlib
from enum import Enum

class EncryptionKeyLength(Enum):
	bits128 = int(128 / 8)
	bits192 = int(192 / 8)
	bits256 = int(256 / 8)
	bits384 = int(384 / 8)
	bits512 = int(512 / 8)

	def __str__(self):
		return self.name
	
	@staticmethod
	def from_string(s):
		try:
				return EncryptionKeyLength[s]
		except KeyError:
				raise ValueError()

def setup() -> ap.Namespace:
	parser = ap.ArgumentParser('Generate a valid encryption key')
	parser.add_argument('length', type=EncryptionKeyLength.from_string, choices=list(EncryptionKeyLength))
	return parser.parse_args()

def main(ns: ap.Namespace):
	key = ''.join(random.choice(string.ascii_letters) for i in range(ns.length.value))
	key_enc = key.encode('utf-8')

	key_b64 = base64.b64encode(key_enc)
	key_b64_str = str(key_b64, 'utf-8') # encryption-key

	hasher = hashlib.sha256()
	hasher.update(key_enc)
	hashed_key = hasher.digest()

	hashed_key_b64 = base64.b64encode(hashed_key) # encryption-key-sha256
	hashed_key_b64_str = str(hashed_key_b64, 'utf-8')

	print(f'Key    (ASCII)\t:\t{key}')
	print(f'Key     (B64)\t:\t{key_b64_str}')
	print(f'KeyHash (B64)\t:\t{hashed_key_b64_str}')
	

if __name__ == '__main__':
	main(setup())
