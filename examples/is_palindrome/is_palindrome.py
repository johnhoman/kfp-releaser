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


if __name__ == "__main__":
    import tempfile
    import os
    import pathlib

    from kfp.compiler import Compiler
    from yaml import safe_load, safe_dump

    def manifest(pipeline_func):
        ref = os.environ.get("GITHUB_REF_NAME", "latest")
        name = pipeline_func.__qualname__
        with tempfile.NamedTemporaryFile(mode="w+t", suffix=".yaml") as open_file:
            Compiler().compile(pipeline_func=pipeline_func, package_path=open_file.name)
            with open(open_file.name, mode="rt") as open_file2:
                workflow = open_file2.read()

        name = name.replace("_", "-").lower()
        version_name = name + "-" + ref

        version = {
            "apiVersion": "aws.jackhoman.com/v1alpha1",
            "kind": "PipelineVersion",
            "metadata": {"name": version_name},
            "spec": {
                "pipeline": name,
                "description": f"pipeline release {ref}",
                "workflow": safe_load(workflow),
            }
        }
        base = pathlib.Path(__file__).parent.joinpath("pipeline", f"{name}.yaml")
        path = pathlib.Path(__file__).parent.joinpath("versions", f"{version_name}.yaml")

        with path.open(mode="wt") as open_file:
            safe_dump(version, stream=open_file, default_flow_style=False)

        if not base.exists():
            pipeline = {
                "apiVersion": "aws.jackhoman.com/v1alpha1",
                "kind": "Pipeline",
                "metadata": {"name": name},
                "spec": {
                    "description": "pipeline declaration",
                }
            }
            base.parent.mkdir(parents=True, exist_ok=True)
            with base.open(mode="wt") as open_file:
                safe_dump(pipeline, stream=open_file, default_flow_style=False)

    manifest(pipeline_func=is_palindrome)
