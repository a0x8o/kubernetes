package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
)

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "register.go",
<<<<<<< HEAD
        "types.generated.go",
=======
>>>>>>> 66f5f2bce071b09222a7a83d1f196f60c34cd224
        "types.go",
        "zz_generated.deepcopy.go",
    ],
    deps = [
<<<<<<< HEAD
        "//vendor/github.com/ugorji/go/codec:go_default_library",
=======
>>>>>>> 66f5f2bce071b09222a7a83d1f196f60c34cd224
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/conversion:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime/schema:go_default_library",
<<<<<<< HEAD
        "//vendor/k8s.io/apimachinery/pkg/types:go_default_library",
=======
>>>>>>> 66f5f2bce071b09222a7a83d1f196f60c34cd224
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
