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
    import tempfile
    import os
    import pathlib

    from kfp.compiler import Compiler
    from yaml import safe_load, safe_dump

    with tempfile.NamedTemporaryFile(mode="w+t", suffix=".yaml") as open_file:
        Compiler().compile(pipeline_func=pipeline, package_path=open_file.name)
        with open(open_file.name, mode="rt") as open_file2:
            workflow = open_file2.read()

    ref = os.environ.get("GITHUB_REF_NAME", "latest")

    manifest = {
        "apiVersion": "aws.jackhoman.com/v1alpha1",
        "kind": "PipelineVersion",
        "metadata": {
            "name": f"is-palindrome-{ref}",
        },
        "spec": {
            "pipeline": "is-palindrome",
            "description": f"pipeline release {ref}",
            "workflow": safe_load(workflow),
        }
    }
    path = pathlib.Path(__file__).parent.joinpath("versions", f"is-palindrome-{ref}.yaml")

    with path.open(mode="wt") as open_file:
        safe_dump(manifest, stream=open_file, default_flow_style=False)