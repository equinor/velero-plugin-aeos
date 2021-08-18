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

	keyb64 = str(base64.b64encode(key_enc), 'utf-8')

	md5key = hashlib.md5(key_enc).digest()
	md5keyb64 = str(base64.b64encode(md5key), 'utf-8')

	print(f'Key:\t\t{key}')
	print(f'Key Base64:\t{keyb64}')
	print(f'Hash Base64:\t{md5keyb64}')
	

if __name__ == '__main__':
	main(setup())
