from hashlib import sha256

class Cell:
    def __init__(self):
        """
        Represents a cell of an IBLTWithEGH
        """
        # Represents the sum (xor of source symbols)
        self.sum = 0
        # Represents the checksum (xor of hashes of source symbols)
        self.checksum = 0
        # Represents the counter array - how many soruce symbols 
        # are mapped to the cell.
        self.counter = 0

    def add(self, symbol: int) -> None:
        """
        Add source symbol to the cell.
        """
        if self.counter == 0:
            self.sum = symbol
            # digest - get the byte representations of the SHA-256 hash.
            self.checksum = sha256(bytes(symbol)).digest()
        else:
            self.sum ^= symbol
            
            # digest - get the byte representations of the SHA-256 hash. 
            symbol_digest = sha256(bytes(symbol)).digest()

            # Perform XOR operation between the hash digests
            xor_result = bytes(a ^ b for a, b in zip(self.checksum, symbol_digest))

            self.checksum = xor_result

        self.counter += 1

    def remove(self, symbol: int) -> None:
        """
        Remove source symbol from the cell.
        """
        self.sum ^= symbol

        # digest - get the byte representations of the SHA-256 hash. 
        symbol_digest = sha256(bytes(symbol)).digest()

        # Perform XOR operation between the hash digests
        xor_result = bytes(a ^ b for a, b in zip(self.checksum, symbol_digest))

        self.checksum = xor_result

        self.counter -= 1

    def is_pure_cell(self) -> bool:
        """
        Check if the cell is pure - containing one element.
        """
        return (self.counter == 1 or self.counter == -1) and (sha256(bytes(self.sum)).digest() == self.checksum)
    