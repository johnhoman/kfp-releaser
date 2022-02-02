from kfp import dsl
from kfp.components import create_component_from_func


@create_component_from_func
def length(s: str) -> int:
    return len(s)


@create_component_from_func
def trim(s: str) -> str:
    return s[1:-1]


@create_component_from_func
def compare(s: str) -> str:
    return "true" if s[0] == s[-1] else "false"


@create_component_from_func
def print_op(s):
    print(s)


@dsl.graph_component
def is_palindrome(s: str):

    l = length(s)
    with dsl.Condition(l.output == 1):
        print_op("true")

    with dsl.Condition(l.output == 0):
        print_op("true")

    with dsl.Condition(l.output > 1):

        comp = compare(s)
        comp.after(l)
        with dsl.Condition(comp.output == "false"):
            print_op("false")

        with dsl.Condition(comp.output == "true"):
            substring = trim(s)
            substring.after(comp)
            is_palindrome(substring.output)


@dsl.pipeline(name="is_palindrome")
def pipeline(s: str = ""):
    is_palindrome(s)


if __name__ == "__main__":
    from kfp.compiler import Compiler
    Compiler().compile(pipeline, package_path=__file__.replace(".py", ".yaml"))
