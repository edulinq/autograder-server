import random

SOME_CONSTANT = 1

some_int = random.randint(1, 2 ** 10)

def some_function():
    global some_int
    some_int = -1

some_function()
