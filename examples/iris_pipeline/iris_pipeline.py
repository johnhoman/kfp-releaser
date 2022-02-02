"""
Copied from https://www.kubeflow.org/docs/components/pipelines/sdk/output-viewer/
"""

from kfp.v2.dsl import component
from kfp.v2.dsl import Output, ClassificationMetrics
from kfp.v2 import dsl


@component(packages_to_install=['sklearn'], base_image='python:3.9')
def iris_sgdclassifier(test_samples_fraction: float, metrics: Output[ClassificationMetrics]):
    from sklearn import datasets, model_selection
    from sklearn.linear_model import SGDClassifier
    from sklearn.metrics import confusion_matrix

    iris_dataset = datasets.load_iris()
    train_x, test_x, train_y, test_y = model_selection.train_test_split(
        iris_dataset['data'], iris_dataset['target'], test_size=test_samples_fraction)

    classifier = SGDClassifier()
    classifier.fit(train_x, train_y)
    predictions = model_selection.cross_val_predict(classifier, train_x, train_y, cv=3)
    metrics.log_confusion_matrix(
        ['Setosa', 'Versicolour', 'Virginica'],
        confusion_matrix(train_y, predictions).tolist()
    )


@dsl.pipeline(name='metrics-visualization-pipeline')
def metrics_visualization_pipeline():
    iris_sgdclassifier_op = iris_sgdclassifier(test_samples_fraction=0.3)


if __name__ == "__main__":
    from kfp.compiler import Compiler
    import kfp
    Compiler(mode=kfp.dsl.PipelineExecutionMode.V2_COMPATIBLE).compile(metrics_visualization_pipeline, package_path=__file__.replace(".py", ".yaml"))
