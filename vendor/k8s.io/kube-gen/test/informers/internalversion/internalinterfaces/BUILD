package(default_visibility = ["//visibility:public"])

<<<<<<< HEAD
licenses(["notice"])

=======
>>>>>>> 66f5f2bce071b09222a7a83d1f196f60c34cd224
load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
)

go_library(
    name = "go_default_library",
    srcs = ["factory_interfaces.go"],
<<<<<<< HEAD
    tags = ["automanaged"],
=======
>>>>>>> 66f5f2bce071b09222a7a83d1f196f60c34cd224
    deps = [
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
        "//vendor/k8s.io/client-go/tools/cache:go_default_library",
        "//vendor/k8s.io/kube-gen/test/clientset/internal:go_default_library",
    ],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [":package-srcs"],
    tags = ["automanaged"],
)
