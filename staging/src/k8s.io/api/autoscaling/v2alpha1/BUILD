package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
)

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "generated.pb.go",
        "register.go",
<<<<<<< HEAD
        "types.generated.go",
=======
>>>>>>> 66f5f2bce071b09222a7a83d1f196f60c34cd224
        "types.go",
        "types_swagger_doc_generated.go",
        "zz_generated.deepcopy.go",
    ],
    deps = [
        "//vendor/github.com/gogo/protobuf/proto:go_default_library",
<<<<<<< HEAD
        "//vendor/github.com/ugorji/go/codec:go_default_library",
=======
>>>>>>> 66f5f2bce071b09222a7a83d1f196f60c34cd224
        "//vendor/k8s.io/api/core/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/api/resource:go_default_library",
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

filegroup(
    name = "go_default_library_protos",
    srcs = ["generated.proto"],
    visibility = ["//visibility:public"],
)
