"""
Copied from https://www.kubeflow.org/docs/components/pipelines/sdk/output-viewer/
"""

from kfp.v2.dsl import component
from kfp.v2.dsl import Output, ClassificationMetrics, Model
from kfp.v2 import dsl


@component(packages_to_install=["sklearn", "joblib"], base_image="python:3.9")
def iris_sgdclassifier(
    test_samples_fraction: float,
    metrics: Output[ClassificationMetrics],
    model: Output[Model],
):
    from sklearn import datasets, model_selection
    from sklearn.linear_model import SGDClassifier
    from sklearn.metrics import confusion_matrix
    import sklearn.metrics as sklearn_metrics
    from joblib import dump

    iris_dataset = datasets.load_iris()
    train_x, test_x, train_y, test_y = model_selection.train_test_split(
        iris_dataset["data"],
        iris_dataset["target"],
        test_size=test_samples_fraction,
    )

    classifier = SGDClassifier()
    classifier.fit(train_x, train_y)
    predictions = model_selection.cross_val_predict(
        classifier,
        train_x,
        train_y,
        cv=3,
    )
    metrics.log_confusion_matrix(
        ['Setosa', 'Versicolour', 'Virginica'],
        confusion_matrix(train_y, predictions).tolist()
    )

    with open(model.path, "wb") as fp:
        dump(classifier, fp)

    pred = classifier.predict(test_x)
    model.metadata["accuracy"] = sklearn_metrics.accuracy_score(pred, test_y)


@dsl.pipeline(name="iris-classifier")
def iris_classifier(test_sample_fraction: float = 0.3):
    iris_sgdclassifier_op = iris_sgdclassifier(
        test_samples_fraction=test_sample_fraction,
    )


if __name__ == "__main__":
    import os
    import tempfile
    import pathlib

    import kfp
    from kfp.compiler import Compiler
    from yaml import safe_load, safe_dump

    def manifest(pipeline_func):

        ref = os.environ.get("GITHUB_REF_NAME", "latest")
        name = pipeline_func.__qualname__
        with tempfile.NamedTemporaryFile(mode="w+t", suffix=".yaml") as open_file:
            Compiler(
                mode=kfp.dsl.PipelineExecutionMode.V2_COMPATIBLE,
            ).compile(pipeline_func=pipeline_func, package_path=open_file.name)
            with open(open_file.name, mode="rt") as open_file2:
                workflow = open_file2.read()

        name = name.replace("_", "-").lower()
        version_name = name + "-" + ref

        version = {
            "apiVersion": "kfp.jackhoman.com/v1alpha1",
            "kind": "PipelineVersion",
            "metadata": {"name": version_name},
            "spec": {
                "pipeline": name,
                "description": f"pipeline release {ref}",
                "workflow": safe_load(workflow),
            }
        }
        this_dir = pathlib.Path(__file__).parent
        this_dir.joinpath("pipeline").mkdir(parents=True, exist_ok=True)
        this_dir.joinpath("versions").mkdir(parents=True, exist_ok=True)
        base = this_dir.joinpath("pipeline", f"{name}.yaml")
        path = this_dir.joinpath("versions", f"{version_name}.yaml")

        with path.open(mode="wt") as open_file:
            safe_dump(version, stream=open_file, default_flow_style=False)

        if not base.exists():
            pipeline = {
                "apiVersion": "kfp.jackhoman.com/v1alpha1",
                "kind": "Pipeline",
                "metadata": {"name": name},
                "spec": {
                    "description": "pipeline declaration",
                }
            }
            base.parent.mkdir(parents=True, exist_ok=True)
            with base.open(mode="wt") as open_file:
                safe_dump(pipeline, stream=open_file, default_flow_style=False)

    manifest(pipeline_func=iris_classifier)
