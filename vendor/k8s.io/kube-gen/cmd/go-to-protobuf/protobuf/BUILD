package(default_visibility = ["//visibility:public"])

<<<<<<< HEAD
licenses(["notice"])

=======
>>>>>>> 66f5f2bce071b09222a7a83d1f196f60c34cd224
load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
    "go_test",
)

go_library(
    name = "go_default_library",
    srcs = [
        "cmd.go",
        "generator.go",
        "import_tracker.go",
        "namer.go",
        "package.go",
        "parser.go",
        "tags.go",
    ],
<<<<<<< HEAD
    tags = ["automanaged"],
=======
>>>>>>> 66f5f2bce071b09222a7a83d1f196f60c34cd224
    deps = [
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/spf13/pflag:go_default_library",
        "//vendor/k8s.io/gengo/args:go_default_library",
        "//vendor/k8s.io/gengo/generator:go_default_library",
        "//vendor/k8s.io/gengo/namer:go_default_library",
        "//vendor/k8s.io/gengo/parser:go_default_library",
        "//vendor/k8s.io/gengo/types:go_default_library",
        "//vendor/k8s.io/kube-gen/third_party/forked/golang/reflect:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["namer_test.go"],
    library = ":go_default_library",
<<<<<<< HEAD
    tags = ["automanaged"],
=======
>>>>>>> 66f5f2bce071b09222a7a83d1f196f60c34cd224
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
