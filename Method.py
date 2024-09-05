
from enum import Enum
 
class Method(Enum):
    """
    The method to use for mapping symbols to IBLT cells.
    """
    EGH = 1
    OLS = 2,
    ID = 3,
    EXTENDED_HAMMING_CODE = 4
