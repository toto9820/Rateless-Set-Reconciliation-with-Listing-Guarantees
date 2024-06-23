import random
import hashlib
from Cell import Cell

class RegularIBLT:
    def __init__(self, n, hash_count):
        self.n = n
        self.hash_count = hash_count
        self.cells = [Cell() for _ in range(n)]
    
    def hashes(self, value):
        return [int(hashlib.sha256((str(value) + str(i)).encode()).hexdigest(), 16) % self.n for i in range(self.hash_count)]
    
    def insert(self, value):
        for h in self.hashes(value):
            self.cells[h].add(value)
    
    def delete(self, value):
        for h in self.hashes(value):
            self.cells[h].remove(value)
    
    def lisitng(self):
        elements = []

        for cell in self.cells:
            if cell.is_pure_cell():
                element = cell.sum
                elements.append(element)
                self.delete(element)

        return elements

if __name__ == "__main__":
    regular_iblt = RegularIBLT(10, 3)

    regular_iblt.insert(1)
    regular_iblt.insert(3)
    regular_iblt.insert(4)
    regular_iblt.insert(6)
    
    print(regular_iblt.lisitng())  