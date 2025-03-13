# Some comment for the first function.
def function1(count):
    for i in range(count):
        if (i % 5 == 0):
            print("Bar")

        if (i % 3 == 0):
            print("Foo")

    return True

def function2(val):
    """
    A docstring-style comment inside of a function.
    """

    return val / 2
